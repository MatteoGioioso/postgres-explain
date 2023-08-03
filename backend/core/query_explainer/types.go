package query_explainer

import (
	"database/sql"
	"time"
)

type PlanRequest struct {
	Query      string        `json:"query"`
	QueryID    string        `json:"query_id"`
	Database   string        `json:"datname"`
	Parameters []interface{} `json:"parameters"`
}

func (p *PlanRequest) paramsFromRequest(params []string) {
	s := make([]interface{}, 0)
	for _, v := range params {
		s = append(s, v)
	}
	p.Parameters = s
}

type PlanEntity struct {
	PlanID           string         `json:"id"`
	TrackingID       string         `json:"tracking_id"`
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

type PlansSearchRequest struct {
	PeriodStartFrom  time.Time `json:"period_start_from"`
	PeriodStartTo    time.Time `json:"period_start_to"`
	ClusterName      string    `json:"cluster_name"`
	Limit            int       `json:"limit"`
	Order            string    `json:"order"`
	QueryFingerprint string    `json:"query_fingerprint"`
	TrackingId       string    `json:"tracking_id"`
}

func (r PlansSearchRequest) ToQueryArgs() map[string]interface{} {
	orderByMap := map[string]string{
		"latest": "period_start",
		"oldest": "period_start",
	}

	if r.Order == "" {
		r.Order = "latest"
	}
	if r.Limit == 0 {
		r.Limit = 100
	}

	m := map[string]interface{}{
		"cluster":  r.ClusterName,
		"order_by": orderByMap[r.Order],
		"limit":    r.Limit,
	}

	return m
}

func (r PlansSearchRequest) ToTmplArgs() interface{} {
	type tmplArgs struct {
		OrderDir string
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
	}
}
