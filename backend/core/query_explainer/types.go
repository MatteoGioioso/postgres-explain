package query_explainer

import (
	"database/sql"
	"time"
)

type PlanRequest struct {
	Query    string `json:"query"`
	QueryID  string `json:"query_id"`
	Database string `json:"datname"`
}

type PlanEntity struct {
	PlanID           string         `json:"id"`
	Alias            sql.NullString `json:"alias"`
	Plan             string         `json:"plan"`
	OriginalPlan     string         `json:"original_plan"`
	Query            string         `json:"query"`
	QueryID          sql.NullString `json:"queryid"`
	QueryFingerprint string         `json:"query_fingerprint"`
	ClusterName      string         `json:"cluster"`
	Database         string         `json:"database"`
	PeriodStart      time.Time      `json:"period_start"`
	Username         string         `json:"username"`
}

type QueryArgs struct {
	PeriodStartFromSec int64
	PeriodStartToSec   int64
	ClusterName        string
}
