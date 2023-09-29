package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	env "github.com/borealisdb/commons/environment"
	"github.com/borealisdb/commons/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/alecthomas/kingpin.v2"
	stdLog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"postgres-explain/backend/auth"
	"postgres-explain/backend/cache"
	"postgres-explain/backend/middlewares"
	"postgres-explain/backend/modules"
	"runtime/debug"
	"sync"
	"time"

	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

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
			Default("8081").
			String()
	httpServerPort = kingpin.Flag("http-server-port", "").
			Envar("HTTP_SERVER_PORT").
			Default("8082").
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
	credentialProvider = kingpin.Flag("credential-provider", "").
				Envar("ENVIRONMENT").
				Default(credentials.EnvironmentProvider).
				Enum(credentials.EnvironmentProvider, credentials.KubernetesProvider)
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

	log := logger.NewDefaultLogger(*logLevelRaw, "backend")
	log.Infof("Starting")

	cacheClient, err := cache.New(cache.Params{})
	if err != nil {
		log.Fatalln(err)
	}

	authFactory := auth.Factory{Providers: map[string]auth.Auth{
		auth.Oauth2Type: &auth.Oauth2{
			Log: log,
		},
		auth.DisabledType: &auth.Disabled{
			Log: log,
		},
	}}
	authProvider := authFactory.Get(*authType)

	err = authProvider.Init(ctx, auth.Params{
		IssuerUrl: fmt.Sprintf("%v%v/identity", *appHost, *rootUrlPath),
		ClientID:  *clientID,
	})
	if err != nil {
		log.Fatalln(err)
	}

	credentialsFactory := credentials.Factory{Providers: map[string]credentials.Credentials{
		credentials.KubernetesProvider:  &credentials.Kubernetes{},
		credentials.EnvironmentProvider: &credentials.Environment{},
	}}
	credentialsProvider := credentialsFactory.Get(*credentialProvider)
	if err := credentialsProvider.Init(); err != nil {
		log.Fatalln(err)
	}

	clickhouseDSN := fmt.Sprintf("clickhouse://%v:%v/backend", *clickhouseHost, *clickhousePort)
	db := NewDB(clickhouseDSN, maxIdleConns, maxOpenConns, log, "/migrations")

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
			// Drop old partitions every 24h.
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

	listen, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalln(err)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				return status.Errorf(codes.Unknown, "panic triggered: %v, %v", p, string(debug.Stack()))
			})),
		)),
	)

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

	// Modules instantiation
	modulesMap, err := GetModules()
	if err != nil {
		log.Fatalln(err)
	}
	for _, module := range modulesMap {
		module.Register(log, db, credentialsProvider, modules.Params{
			WaitEventsMapFilePath: "/",
		})
		if err := module.Init(modules.InitArgs{
			Ctx:         ctx,
			GrpcServer:  grpcServer,
			Mux:         proxyMux,
			GrpcAddress: grpcAddress,
			Opts:        opts,
			Cache:       cacheClient,
		}); err != nil {
			log.Fatalln("Module failed to initialize: %v", err)
		}
	}

	reflection.Register(grpcServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		go func() {
			if err := grpcServer.Serve(listen); err != nil {
				log.Fatalln(err)
			}
		}()

		<-ctx.Done()
		grpcServer.Stop()
		listen.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runHTTPServer(ctx, proxyMux, httpAddress, authProvider.AuthMiddleware, log)
	}()

	wg.Wait()
}

func runHTTPServer(
	ctx context.Context,
	proxyMux *grpc_gateway.ServeMux,
	httpAddress string,
	authMiddleware func(h http.Handler) http.Handler,
	log *logrus.Entry,
) {
	l := logrus.WithField("component", "JSON")
	l.Infof("Starting server on http://0.0.0.0:%s/ ...", httpAddress)

	stack := []middlewares.Middleware{
		authMiddleware,
		middlewares.CORSMiddleware,
	}

	mux := http.NewServeMux()
	mux.Handle("/", middlewares.CompileMiddleware(proxyMux, stack))

	server := &http.Server{
		Addr:     httpAddress,
		ErrorLog: stdLog.New(logrusErrorWriter{Log: log}, "server", 0),
		Handler:  mux,
	}
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
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
