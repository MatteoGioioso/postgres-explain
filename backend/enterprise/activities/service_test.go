package activities

import (
	"database/sql"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"postgres-explain/proto"
	"testing"
)

// We replicate this type because we cannot unmarshal sql.NullString
type QueryDBTest struct {
	QueryDB
	ParsedQuery        string `json:"parsed_query"`
	IsQueryTruncated   string `json:"is_query_truncated"`
	IsQueryExplainable string `json:"is_query_explainable"`
}

func loadFixtures(t *testing.T) []QueryDB {
	file, err := os.ReadFile("../../fixtures/queries_001.json")
	if err != nil {
		t.Fatal(err)
	}

	f := make([]QueryDBTest, 0)
	if err := json.Unmarshal(file, &f); err != nil {
		t.Fatal(err)
	}

	f2 := make([]QueryDB, 0)
	for _, fx := range f {
		f2 = append(f2, QueryDB{
			Fingerprint:       fx.Fingerprint,
			CPULoadWaitEvents: fx.CPULoadWaitEvents,
			CPULoadTotal:      fx.CPULoadTotal,
			ParsedQuery: sql.NullString{
				String: fx.ParsedQuery,
				Valid:  true,
			},
			Query:    fx.Query,
			QuerySha: fx.QuerySha,
		})
	}

	return f2
}

func TestService_mapQueriesToPlotlyTraces(t *testing.T) {
	type fields struct {
		Repo                     Repository
		WaitEventsMap            map[string]WaitEvent
		log                      *logrus.Entry
		ActivitiesProfilerServer proto.ActivitiesServer
	}
	type args struct {
		queries []QueryDB
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]*proto.Trace
	}{
		{
			name: "map queries to trace",
			fields: fields{
				Repo:                     Repository{},
				WaitEventsMap:            LoadWaitEventsMapFromFile("./"),
				log:                      &logrus.Entry{Logger: logrus.New()},
				ActivitiesProfilerServer: nil,
			},
			args: args{queries: loadFixtures(t)},
			want: map[string]*proto.Trace{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aps := &Service{
				Repo:             tt.fields.Repo,
				WaitEventsMap:    tt.fields.WaitEventsMap,
				log:              tt.fields.log,
				ActivitiesServer: tt.fields.ActivitiesProfilerServer,
			}
			res := aps.mapQueriesToTraces(tt.args.queries, func(db QueryDB) string {
				return db.Fingerprint
			})

			assert.Equal(t, len(res), 13)
		})
	}
}
