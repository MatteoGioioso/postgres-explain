package modules

import (
	"context"
	"github.com/borealisdb/commons/credentials"
	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"postgres-explain/backend/cache"
)

type Module interface {
	Register(log *logrus.Entry, db *sqlx.DB, credentialsProvider credentials.Credentials, params Params)
	Init(initArgs InitArgs) error
}

type Params struct {
	WaitEventsMapFilePath string `json:"waitEventsMapFilePath"`
}

type InitArgs struct {
	Ctx         context.Context
	GrpcServer  *grpc.Server
	Mux         *grpc_gateway.ServeMux
	GrpcAddress string
	Cache       *cache.Client
	Opts        []grpc.DialOption
}
