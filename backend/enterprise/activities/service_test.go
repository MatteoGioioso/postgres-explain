package activities

import (
	"database/sql"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"postgres-explain/proto"
	"reflect"
	"testing"
)

// We replicate this type because we cannot unmarshal sql.NullString
type QueryDBTest struct {
	ID                string             `json:"query_id"`
	CPULoadWaitEvents map[string]float64 `json:"cpu_load_wait_events"`
	CPULoadTotal      float32            `json:"cpu_load_total"`
	ParsedQuery       string             `json:"parsed_query"`
	Query             string             `json:"query"`
}

func loadFixtures(t *testing.T) []QueryDB {
	file, err := ioutil.ReadFile("../../../testdata/fixtures/queries.json")
	if err != nil {
		t.Fatal(err)
	}

	f := make([]QueryDBTest, 0)
	if err := json.Unmarshal(file, &f); err != nil {
		t.Fatal(err)
	}

	f2 := make([]QueryDB, 0)
	for _, q := range f {
		f2 = append(f2, QueryDB{
			CPULoadWaitEvents: q.CPULoadWaitEvents,
			CPULoadTotal:      q.CPULoadTotal,
			ParsedQuery: sql.NullString{
				String: q.ParsedQuery,
				Valid:  true,
			},
			Query: q.Query,
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
			if got := aps.mapQueriesToTraces(tt.args.queries, nil); !reflect.DeepEqual(got, tt.want) {
				for _, q := range got["AddinShmemInit"].XValuesString {
					t.Log(q)
				}
				t.Log(len(got["AddinShmemInit"].XValuesString))

				//t.Errorf("mapQueriesToTraces() = %v, want %v", got, tt.want)
			}
		})
	}
}
