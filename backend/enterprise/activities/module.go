package activities

import (
	"github.com/borealisdb/commons/credentials"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/enterprise/shared"
	"postgres-explain/backend/modules"
	"postgres-explain/proto"
)

const ModuleName = "activities"

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
	activitiesProfilerService := NewService(
		NewActivitiesRepository(m.DB),
		shared.NewMetricsRepository(m.DB),
		LoadWaitEventsMapFromFile(m.WaitEventsMapFilePath),
		m.Log,
	)
	activityCollectorService := &ActivityCollectorService{
		ActivitySampler: NewActivitySampler(m.DB, m.Log),
		Log:             m.Log,
	}
	proto.RegisterActivityCollectorServer(initArgs.GrpcServer, activityCollectorService)
	proto.RegisterActivitiesServer(initArgs.GrpcServer, activitiesProfilerService)

	return nil
}
