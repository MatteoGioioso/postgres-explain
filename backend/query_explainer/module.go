package query_explainer

import (
	"github.com/borealisdb/commons/credentials"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"postgres-explain/core/proto"
)

const ModuleName = "query_explainer"

type Module struct {
	DB                  *sqlx.DB
	Log                 *logrus.Entry
	CredentialsProvider credentials.Credentials
}

func (m Module) Init(grpcServer *grpc.Server) error {
	repository := Repository{DB: m.DB}
	service := Service{
		log:                 m.Log.WithField("subcomponent", "query-optimization"),
		Repo:                repository,
		credentialsProvider: m.CredentialsProvider,
	}

	proto.RegisterQueryExplainerServer(grpcServer, &service)
	return nil
}
