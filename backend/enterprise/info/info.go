package info

import (
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
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
	m.Log = log.WithField("module", ModuleName)
	m.DB = db
	m.CredentialsProvider = credentialsProvider
	m.Log.Infof("registered")
}

func (m *Module) Init(initArgs modules.InitArgs) error {
	repository := Repository{credentialsProvider: m.CredentialsProvider, log: m.Log, cacheClient: initArgs.Cache}
	service := Service{
		log:                 m.Log,
		Repo:                repository,
		credentialsProvider: m.CredentialsProvider,
	}

	proto.RegisterInfoServer(initArgs.GrpcServer, &service)
	if err := proto.RegisterInfoHandlerFromEndpoint(initArgs.Ctx, initArgs.Mux, initArgs.GrpcAddress, initArgs.Opts); err != nil {
		return fmt.Errorf("could not register InfoHandlerFromEndpoint: %v", err)
	}
	m.Log.Infof("initialized")

	return nil
}
