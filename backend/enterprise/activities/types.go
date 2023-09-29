package activities

import (
	"database/sql"
	"postgres-explain/proto"
	"time"
)

type QueryRank struct {
	ID    string
	Total float32
	Query string
}

type QuerySlot map[string]float32
type QueriesSlots map[string]QuerySlot

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type waitEventsGroup map[string]Group

var waitEventsGroupsMap = waitEventsGroup{
	"application_name": Group{
		ID:   "application_name",
		Name: "Application",
	},
	"usename": Group{
		ID:   "usename",
		Name: "User",
	},
	"datname": Group{
		ID:   "datname",
		Name: "Database",
	},
	"instance_name": Group{
		ID:   "instance_name",
		Name: "Instance",
	},
}

type QueryMetadataRequest struct {
	Fingerprint string `json:"fingerprint"`
	ID          string `json:"query_id"`
}

type QueryMetadata struct {
	Database         string `json:"datname"`
	Query            string `json:"query"`
	ParsedQuery      string `json:"parsed_query"`
	IsQueryTruncated uint8  `json:"is_query_truncated"`
}

type QueryArgs struct {
	PeriodStartFromSec int64
	PeriodStartToSec   int64
	ClusterName        string
	Fingerprint        string
}

type SlotDB struct {
	Timestamp      time.Time `json:"slot"`
	WaitEventCount int       `json:"wait_event_count"`
	WaitEventName  string    `json:"wait_event"`
	CpuCores       float32   `json:"cpu_cores"`
}

type QueryDB struct {
	Fingerprint        string             `json:"fingerprint"`
	CPULoadWaitEvents  map[string]float64 `json:"cpu_load_wait_events"`
	CPULoadTotal       float32            `json:"cpu_load_total"`
	ParsedQuery        sql.NullString     `json:"parsed_query"`
	Query              string             `json:"query"`
	QuerySha           string             `json:"query_sha"`
	IsQueryTruncated   uint8              `json:"is_query_truncated"`
	IsQueryExplainable bool               `json:"is_query_explainable"`
}

func (q QueryDB) GetSQL() string {
	if q.ParsedQuery.Valid && q.ParsedQuery.String != "" {
		return q.ParsedQuery.String
	}

	return q.Query
}

type QueryByFingerprintDB struct {
	CPULoadWaitEvents map[string]float64 `json:"cpu_load_wait_events"`
	CPULoadTotal      float32            `json:"cpu_load_total"`
	Query             string             `json:"query"`
}

type ActivitySampleDB struct {
	PeriodStart      time.Time `json:"period_start"`
	PeriodLength     uint32    `json:"period_length"`
	CurrentTimestamp time.Time `json:"current_timestamp"`
	IsQueryTruncated uint8     `json:"is_query_truncated"`
	*proto.ActivitySample
}

func (s *ActivitySampleDB) FromActivitySample(sample *proto.ActivitySample) {
	isTruncated := 0
	if sample.IsQueryTruncated {
		isTruncated = 1
	}

	s.PeriodStart = time.Unix(int64(sample.GetPeriodStartUnixSecs()), 0).UTC()
	s.PeriodLength = sample.GetPeriodLengthSecs()
	s.CurrentTimestamp = time.Unix(int64(sample.GetCurrentTimestamp()), 0).UTC()
	s.IsQueryTruncated = uint8(isTruncated)
	s.ActivitySample = sample
}
