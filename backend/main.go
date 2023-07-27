package main

import (
	"context"
	"fmt"
	"github.com/borealisdb/commons/constants"
	"github.com/borealisdb/commons/credentials"
	env "github.com/borealisdb/commons/environment"
	"github.com/borealisdb/commons/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/alecthomas/kingpin.v2"
	stdLog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"postgres-explain/backend/auth"
	"postgres-explain/backend/middlewares"
	"postgres-explain/backend/query_explainer"
	"postgres-explain/proto"
	"sync"
	"time"

	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type Module interface {
	Init(grpcServer *grpc.Server) error
}

const (
	shutdownTimeout = 3 * time.Second
	maxIdleConns    = 5
	maxOpenConns    = 10
)

var (
	dataRetentionDays = kingpin.Flag("data-retention-days", "").
				Envar("DATA_RETENTION").
				Default("30").
				Uint()
	clickhouseHost = kingpin.Flag("clickhouse-host", "").
			Envar("CLICKHOUSE_HOST").
			Default("localhost").
			String()
	clickhousePort = kingpin.Flag("clickhouse-port", "").
			Envar("CLICKHOUSE_PORT").
			Default("9000").
			String()
	grpcServerPort = kingpin.Flag("grpc-server-port", "").
			Envar("GRPC_SERVER_PORT").
			Default(constants.MonitoringAPIGRPCPort).
			String()
	httpServerPort = kingpin.Flag("http-server-port", "").
			Envar("HTTP_SERVER_PORT").
			Default(constants.MonitoringAPIPort).
			String()
	clientID = kingpin.Flag("client-id", "").
			Envar("CLIENT_ID").
			Default("borealis").
			String()
	appHost = kingpin.Flag("app-host", "").
		Envar("APP_HOST").
		Default("http://localhost:8080").
		String()
	rootUrlPath = kingpin.Flag("root-url-path", "").
			Envar("ROOT_URL_PATH").
			Default("/borealis").
			String()
	authType = kingpin.Flag("auth-type", "type of authentication").
			Envar("AUTH_TYPE").
			Default(auth.Oauth2Type).
			String()
	environment = kingpin.Flag("environment", "").
			Envar("ENVIRONMENT").
			Enum(env.Kubernetes, env.VM, env.Mock)
	logLevelRaw = kingpin.Flag("log-level", "").
			Envar("LOG_LEVEL").
			Default("info").
			Enum("debug", "info", "warning")
)

// Workaround for http.Server
type logrusErrorWriter struct {
	Log *logrus.Entry
}

func (w logrusErrorWriter) Write(p []byte) (n int, err error) {
	w.Log.Errorf("%s", string(p))
	return len(p), nil
}

func main() {
	kingpin.Parse()
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	log := logger.NewDefaultLogger(*logLevelRaw, "monitoring-server")
	log.Infof("Starting")

	authFactory := auth.Factory{Providers: map[string]auth.Auth{
		auth.Oauth2Type: &auth.Oauth2{
			Log: log,
		},
		auth.DisabledType: &auth.Disabled{
			Log: log,
		},
	}}
	authProvider := authFactory.Get(*authType)

	err := authProvider.Init(ctx, auth.Params{
		IssuerUrl: fmt.Sprintf("%v%v/identity", *appHost, *rootUrlPath),
		ClientID:  *clientID,
	})
	if err != nil {
		log.Fatalln(err)
	}

	determinedEnvironment, err := env.DetermineEnvironment(*environment)
	if err != nil {
		log.Fatalln(err)
	}

	credentialsFactory := credentials.Factory{Providers: map[string]credentials.Credentials{
		env.Kubernetes: &credentials.Kubernetes{},
		env.VM:         &credentials.VM{},
	}}
	credentialsProvider := credentialsFactory.Get(determinedEnvironment)
	if err := credentialsProvider.Init(); err != nil {
		log.Fatalln(err)
	}

	clickhouseDSN := fmt.Sprintf("clickhouse://%v:%v/bmserver", *clickhouseHost, *clickhousePort)
	db := NewDB(clickhouseDSN, maxIdleConns, maxOpenConns, log, "/migrations")

	modulesList := []string{query_explainer.ModuleName}
	modulesMap := map[string]Module{
		query_explainer.ModuleName: &query_explainer.Module{
			DB:                  db,
			Log:                 log,
			CredentialsProvider: credentialsProvider,
		},
	}

	// handle termination signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		s := <-signals
		signal.Stop(signals)
		log.Printf("Got %s, shutting down...\n", unix.SignalName(s.(unix.Signal)))
		cancel()
	}()

	ticker := time.NewTicker(24 * time.Hour)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// Drop old partitions once in 24h.
			DropOldPartition(db, "plans", *dataRetentionDays, log)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// nothing
			}
		}
	}()
	grpcAddress := ":" + (*grpcServerPort)
	httpAddress := ":" + (*httpServerPort)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runGRPCServer(
			modulesMap,
			modulesList,
			grpcAddress,
			log,
		)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runHTTPServer(ctx, grpcAddress, httpAddress, authProvider.AuthMiddleware, log)
	}()

	wg.Wait()
}

func runGRPCServer(
	modulesMap map[string]Module,
	modulesList []string,
	grpcAddress string,
	log *logrus.Entry,
) {
	listen, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalln(err)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	for _, moduleName := range modulesList {
		module, ok := modulesMap[moduleName]
		if ok {
			if err := module.Init(grpcServer); err != nil {
				log.Fatalln("Module failed to initialize: %v", err)
			}
		}
	}

	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalln(err)
	}
}

func runHTTPServer(
	ctx context.Context,
	grpcAddress, httpAddress string,
	authMiddleware func(h http.HandlerFunc) http.HandlerFunc,
	log *logrus.Entry,
) {
	l := logrus.WithField("component", "JSON")
	l.Infof("Starting server on http://0.0.0.0:%s/ ...", httpAddress)

	marshaller := &grpc_gateway.JSONPb{
		MarshalOptions: protojson.MarshalOptions{ //nolint:exhaustivestruct
			UseEnumNumbers:  false,
			EmitUnpopulated: false,
			UseProtoNames:   true,
			Indent:          "  ",
		},
		UnmarshalOptions: protojson.UnmarshalOptions{ //nolint:exhaustivestruct
			DiscardUnknown: true,
		},
	}

	proxyMux := grpc_gateway.NewServeMux(
		grpc_gateway.WithMarshalerOption(grpc_gateway.MIMEWildcard, marshaller),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	type registrar func(context.Context, *grpc_gateway.ServeMux, string, []grpc.DialOption) error
	for _, r := range []registrar{
		proto.RegisterQueryExplainerHandlerFromEndpoint,
	} {
		if err := r(ctx, proxyMux, grpcAddress, opts); err != nil {
			l.Panic(err)
		}
	}

	stack := []middlewares.Middleware{
		authMiddleware,
		middlewares.CORSMiddleware,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", middlewares.CompileMiddleware(proxyMux.ServeHTTP, stack))

	server := &http.Server{
		Addr:     httpAddress,
		ErrorLog: stdLog.New(logrusErrorWriter{Log: log}, "", 0),
		Handler:  mux,
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			l.Panic(err)
		}
		l.Println("Server stopped.")
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	if err := server.Shutdown(ctx); err != nil {
		l.Errorf("Failed to shutdown gracefully: %s \n", err)
		server.Close()
	}
	cancel()
}
