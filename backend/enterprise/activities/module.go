package activities

import (
	"github.com/borealisdb/commons/credentials"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/enterprise/core"
	"postgres-explain/backend/modules"
	"postgres-explain/proto"
)

const ModuleName = "activities"

type Module struct {
	Log                   *logrus.Entry
	DB                    *sqlx.DB
	WaitEventsMapFilePath string
	CredentialProvider    credentials.Credentials
	MetricsRepository     core.MetricsRepository
}

func (m *Module) Register(log *logrus.Entry, db *sqlx.DB, credentialsProvider credentials.Credentials, params modules.Params) {
	m.Log = log.WithField("module", ModuleName)
	m.DB = db
	m.CredentialsProvider = credentialsProvider
	m.Log.Infof("registered")
}

func (m Module) Init(initArgs modules.InitArgs) error {
	waitEventsMap := LoadWaitEventsMapFromFile(m.WaitEventsMapFilePath)
	activitiesRepository := NewActivitiesRepository(m.DB)
	activitySampler := NewActivitySampler(m.DB, m.Log)
	activitiesProfilerService := NewService(
		activitiesRepository,
		m.MetricsRepository,
		waitEventsMap,
		m.CredentialProvider,
		m.Log,
	)
	activityCollectorService := ActivityCollectorService{
		ActivitySampler: activitySampler,
		Log:             m.Log.WithField("subcomponent", "activities-collector"),
	}
	proto.RegisterActivityCollectorServer(grpcServer, &activityCollectorService)
	proto.RegisterActivitiesProfilerServer(grpcServer, activitiesProfilerService)

	return nil
}
