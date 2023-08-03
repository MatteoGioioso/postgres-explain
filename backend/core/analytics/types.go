package analytics

import (
	"postgres-explain/core/pkg"
	"postgres-explain/proto"
	"time"
)

type QueriesMetricsRequest struct {
	PeriodStartFrom time.Time `json:"period_start_from"`
	PeriodStartTo   time.Time `json:"period_start_to"`
	ClusterName     string    `json:"cluster_name"`
	Limit           int       `json:"limit"`
	Order           string    `json:"order"`
}

func (e QueriesMetricsRequest) GetQueryArgs() map[string]interface{} {
	if e.Limit == 0 {
		e.Limit = 10
	}
	return map[string]interface{}{
		"limit": e.Limit,
	}
}

func (e QueriesMetricsRequest) GetTemplateArgs() interface{} {
	type tmplArgs struct {
		OrderBy  string
		OrderDir string
	}
	s := ordersMap[e.Order]

	return tmplArgs{
		OrderBy:  s.OrderBy,
		OrderDir: s.OrderDir,
	}
}

var ordersMap = map[string]struct {
	OrderDir string
	OrderBy  string
}{
	"most_calls":         {OrderDir: "DESC", OrderBy: "calls"},
	"longest_avg_time":   {OrderDir: "DESC", OrderBy: "mean_exec_time"},
	"most_rows_returned": {OrderDir: "DESC", OrderBy: "rows"},
	"":                   {OrderDir: "DESC", OrderBy: "mean_exec_time"},
}

type MetricsEntity map[string]interface{}

var MetricsMappings = []*proto.MetricInfo{
	{
		Key:   "calls",
		Type:  "sum",
		Kind:  pkg.Quantity,
		Title: "Calls",
	},
	{
		Key:   "rows",
		Type:  "sum",
		Kind:  pkg.Quantity,
		Title: "Rows returned",
	},
	{
		Key:   "total_exec_time",
		Type:  "sum",
		Kind:  pkg.Timing,
		Title: "Time",
	},
	{
		Key:   "mean_exec_time",
		Type:  "sum",
		Kind:  pkg.Timing,
		Title: "Time spent per call average in ms",
	},
	{
		Key:   "mean_plan_time",
		Type:  "sum",
		Kind:  pkg.Timing,
		Title: "Mean time",
	},
	{
		Key:   "plans",
		Type:  "sum",
		Kind:  pkg.Quantity,
		Title: "Plans",
	},
	{
		Key:   "shared_blks_read",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "Blocks read",
	},
	{
		Key:   "shared_blks_dirtied",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "shared_blks_hit",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "Buffer hits",
	},
	{
		Key:   "shared_blks_written",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "Blocks written",
	},
	{
		Key:   "local_blks_hit",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "local_blks_read",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "local_blks_dirtied",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "local_blks_written",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "temp_blks_read",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "temp_blks_written",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "blk_read_time",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
	{
		Key:   "blk_write_time",
		Type:  "sum",
		Kind:  pkg.Blocks,
		Title: "",
	},
}

var MetricsMappingsSimple = []*proto.MetricInfo{
	{
		Key:   "calls",
		Type:  "sum",
		Kind:  pkg.Quantity,
		Title: "Calls",
	},
	{
		Key:   "rows",
		Type:  "sum",
		Kind:  pkg.Quantity,
		Title: "Rows returned",
	},
	{
		Key:   "total_exec_time",
		Type:  "sum",
		Kind:  pkg.Timing,
		Title: "Total time",
	},
	{
		Key:   "mean_exec_time",
		Type:  "sum",
		Kind:  pkg.Timing,
		Title: "Mean Time",
	},
	{
		Key:   "mean_plan_time",
		Type:  "sum",
		Kind:  pkg.Timing,
		Title: "Mean plan time",
	},
	{
		Key:   "plans",
		Type:  "sum",
		Kind:  pkg.Quantity,
		Title: "Plans",
	},
}
