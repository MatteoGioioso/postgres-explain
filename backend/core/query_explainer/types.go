package query_explainer

import "time"

type PlanRequest struct {
	Query    string `json:"query"`
	QueryID  string `json:"query_id"`
	Database string `json:"datname"`
}

type PlanEntity struct {
	Plan             string    `json:"plan"`
	PlanID           string    `json:"plan_id"`
	OriginalPlan     string    `json:"original_plan"`
	Query            string    `json:"query"`
	QueryID          string    `json:"query_id"`
	QueryFingerprint string    `json:"query_fingerprint"`
	ClusterName      string    `json:"cluster_name"`
	Database         string    `json:"datname"`
	PeriodStart      time.Time `json:"period_start"`
}

type QueryArgs struct {
	PeriodStartFromSec int64
	PeriodStartToSec   int64
	ClusterName        string
}
