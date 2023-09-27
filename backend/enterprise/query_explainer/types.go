package query_explainer

import (
	"database/sql"
	"time"
)

type PlanEntity struct {
	PlanID           string         `json:"id"`
	OptimizationId   string         `json:"optimization_id"`
	Alias            sql.NullString `json:"alias"`
	Plan             string         `json:"plan"` // Explained object
	OriginalPlan     string         `json:"original_plan"`
	Query            string         `json:"query"`
	QueryID          sql.NullString `json:"queryid"`
	QueryFingerprint string         `json:"query_fingerprint"`
	ClusterName      string         `json:"cluster"`
	Database         string         `json:"database"`
	PeriodStart      time.Time      `json:"period_start"`
	Username         string         `json:"username"`
}

type PlansSearchRequest struct {
	PeriodStartFrom  time.Time `json:"period_start_from"`
	PeriodStartTo    time.Time `json:"period_start_to"`
	ClusterName      string    `json:"cluster_name"`
	Limit            int       `json:"limit"`
	Order            string    `json:"order"`
	QueryFingerprint string    `json:"query_fingerprint"`
	OptimizationId   string    `json:"optimization_id"`
}

func (r PlansSearchRequest) ToQueryArgs() map[string]interface{} {
	if r.Limit == 0 {
		r.Limit = 100
	}

	m := map[string]interface{}{
		"cluster":           r.ClusterName,
		"limit":             r.Limit,
		"query_fingerprint": r.QueryFingerprint,
		"optimization_id":   r.OptimizationId,
	}

	return m
}

func (r PlansSearchRequest) ToTmplArgs() interface{} {
	type tmplArgs struct {
		OrderDir string
		OrderBy  string
	}

	orderByMap := map[string]string{
		"latest": "period_start",
		"oldest": "period_start",
	}
	orderDirMap := map[string]string{
		"latest": "DESC",
		"oldest": "ASC",
	}

	if r.Order == "" {
		r.Order = "latest"
	}

	return tmplArgs{
		OrderDir: orderDirMap[r.Order],
		OrderBy:  orderByMap[r.Order],
	}
}

type QueryMetadataRequest struct {
	Fingerprint string `json:"fingerprint"`
	ID          string `json:"query_id"`
}

type QueryMetadata struct {
	Database    string `json:"datname"`
	Query       string `json:"query"`
	ParsedQuery string `json:"parsed_query"`
}
