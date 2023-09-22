package analytics

import (
	"github.com/borealisdb/commons/credentials"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/modules"
	"postgres-explain/proto"
)

const ModuleName = "analytics"

type Module struct {
	Log *logrus.Entry
	DB  *sqlx.DB

	modules.Params
}

func (m *Module) Register(log *logrus.Entry, db *sqlx.DB, credentialsProvider credentials.Credentials, params modules.Params) {
	m.Log = log.WithField("module", ModuleName)
	m.DB = db
	m.Log.Infof("registered")
	m.Params = params
}

func (m *Module) Init(initArgs modules.InitArgs) error {
	metricsBucket := NewMetricsBucket(m.DB, m.Log)

	receiver := &Receiver{
		MetricsBucket: metricsBucket,
		Log:           m.Log,
	}

	go func() {
		metricsBucket.Run(initArgs.Ctx)
	}()

	proto.RegisterStatementsCollectorServer(initArgs.GrpcServer, receiver)

	return nil
}
