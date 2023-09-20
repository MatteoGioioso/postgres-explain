package analytics

import (
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/modules"
	"postgres-explain/proto"
)

const ModuleName = "analytics"

type Module struct {
	DB                  *sqlx.DB
	Log                 *logrus.Entry
	CredentialsProvider credentials.Credentials
}

func (m *Module) Register(log *logrus.Entry, db *sqlx.DB, credentialsProvider credentials.Credentials, params modules.Params) {
	m.Log = log.WithField("module", ModuleName)
	m.DB = db
	m.CredentialsProvider = credentialsProvider
	m.Log.Infof("registered")
}

func (m *Module) Init(initArgs modules.InitArgs) error {
	repository := Repository{
		log:                 m.Log,
		credentialsProvider: m.CredentialsProvider,
	}
	service := Service{
		log:                 m.Log,
		repo:                repository,
		credentialsProvider: m.CredentialsProvider,
	}

	proto.RegisterQueryAnalyticsServer(initArgs.GrpcServer, &service)
	if err := proto.RegisterQueryAnalyticsHandlerFromEndpoint(initArgs.Ctx, initArgs.Mux, initArgs.GrpcAddress, initArgs.Opts); err != nil {
		return fmt.Errorf("could not register QueryExplainerHandlerFromEndpoint: %v", err)
	}
	m.Log.Infof("initialized")
	return nil
}
