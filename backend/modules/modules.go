package modules

import (
	"context"
	"github.com/borealisdb/commons/credentials"
	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Module interface {
	Register(log *logrus.Entry, db *sqlx.DB, credentialsProvider credentials.Credentials, params Params)
	Init(ctx context.Context, grpcServer *grpc.Server, mux *grpc_gateway.ServeMux, grpcAddress string, opts []grpc.DialOption) error
}

type Params struct {
}
