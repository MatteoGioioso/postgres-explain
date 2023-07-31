package info

import (
	"context"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"postgres-explain/backend/modules"
	"postgres-explain/proto"
)

const ModuleName = "info"

type Module struct {
	DB                  *sqlx.DB
	Log                 *logrus.Entry
	CredentialsProvider credentials.Credentials
}

func (m *Module) Register(log *logrus.Entry, db *sqlx.DB, credentialsProvider credentials.Credentials, params modules.Params) {
	m.Log = log
	m.DB = db
	m.CredentialsProvider = credentialsProvider
}

func (m *Module) Init(ctx context.Context, grpcServer *grpc.Server, mux *runtime.ServeMux, address string, opts []grpc.DialOption) error {
	repository := Repository{DB: m.DB}
	service := Service{
		log:                 m.Log.WithField("module", ModuleName),
		Repo:                repository,
		credentialsProvider: m.CredentialsProvider,
	}

	proto.RegisterQueryExplainerServer(grpcServer, &service)
	if err := proto.RegisterInfoHandlerFromEndpoint(ctx, mux, address, opts); err != nil {
		return fmt.Errorf("could not register InfoHandlerFromEndpoint: %v", err)
	}
	return nil
}
