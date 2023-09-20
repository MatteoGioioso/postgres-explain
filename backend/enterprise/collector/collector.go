package collector

import (
	"github.com/borealisdb/commons/credentials"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/modules"
	"postgres-explain/proto"
)

const ModuleName = "collector"

type Module struct {
	Log                 *logrus.Entry
	CredentialsProvider credentials.Credentials
}

func (m *Module) Register(log *logrus.Entry, db *sqlx.DB, credentialsProvider credentials.Credentials, params modules.Params) {
	m.Log = log.WithField("module", ModuleName)
	m.Log.Infof("registered")
}

func (m *Module) Init(initArgs modules.InitArgs) error {
	service := Service{
		log:         m.Log,
		cacheClient: initArgs.Cache,
	}

	proto.RegisterCollectorServer(initArgs.GrpcServer, &service)
	m.Log.Infof("initialized")
	return nil
}
