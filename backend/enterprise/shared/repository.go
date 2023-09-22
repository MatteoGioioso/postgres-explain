package shared

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	OptimalAmountOfPoint = 120
	MinFullTimeFrame     = 2 * time.Hour
)

// MetricsRepository represents methods to work with metrics.
type MetricsRepository struct {
	db *sqlx.DB
}

// NewMetricsRepository initialize MetricsRepository with db instance.
func NewMetricsRepository(db *sqlx.DB) MetricsRepository {
	return MetricsRepository{db: db}
}

type MetricsGetArgs struct {
	PeriodStartFromSec, PeriodStartToSec int64
	Filter, Group                        string
	Dimensions, Labels                   map[string][]string
	Totals                               bool
}

const queryMetricsTmpl = `
SELECT

SUM(num_queries) AS num_queries,
SUM(num_queries_with_errors) AS num_queries_with_errors,
SUM(num_queries_with_warnings) AS num_queries_with_warnings,

SUM(m_query_time_cnt) AS m_query_time_cnt,
SUM(m_query_time_sum) AS m_query_time_sum,
MIN(m_query_time_min) AS m_query_time_min,
MAX(m_query_time_max) AS m_query_time_max,
AVG(m_query_time_p99) AS m_query_time_p99,

SUM(m_rows_sent_cnt) AS m_rows_sent_cnt,
SUM(m_rows_sent_sum) AS m_rows_sent_sum,
MIN(m_rows_sent_min) AS m_rows_sent_min,
MAX(m_rows_sent_max) AS m_rows_sent_max,
AVG(m_rows_sent_p99) AS m_rows_sent_p99,

SUM(m_shared_blks_hit_sum) AS m_shared_blks_hit_sum,
SUM(m_shared_blks_read_sum) AS m_shared_blks_read_sum,
SUM(m_shared_blks_dirtied_sum) AS m_shared_blks_dirtied_sum,
SUM(m_shared_blks_written_sum) AS m_shared_blks_written_sum,

SUM(m_local_blks_hit_sum) AS m_local_blks_hit_sum,
SUM(m_local_blks_read_sum) AS m_local_blks_read_sum,
SUM(m_local_blks_dirtied_sum) AS m_local_blks_dirtied_sum,
SUM(m_local_blks_written_sum) AS m_local_blks_written_sum,

SUM(m_temp_blks_read_sum) AS m_temp_blks_read_sum,
SUM(m_temp_blks_written_sum) AS m_temp_blks_written_sum,
SUM(m_blk_read_time_sum) AS m_blk_read_time_sum,
SUM(m_blk_write_time_sum) AS m_blk_write_time_sum

FROM metrics
WHERE period_start >= :period_start_from AND period_start <= :period_start_to
{{ if not .Totals }} AND {{ .Group }} = '{{ .DimensionVal }}' {{ end }}
{{ if .Dimensions }}
    {{range $key, $vals := .Dimensions }}
        AND {{ $key }} IN ( '{{ StringsJoin $vals "', '" }}' )
    {{ end }}
{{ end }}
{{ if .Labels }}{{$i := 0}}
    AND ({{range $key, $vals := .Labels }}{{ $i = inc $i}}
        {{ if gt $i 1}} OR {{ end }} has(['{{ StringsJoin $vals "', '" }}'], labels.value[indexOf(labels.key, '{{ $key }}')])
    {{ end }})
{{ end }}
{{ if not .Totals }} GROUP BY {{ .Group }} {{ end }}
	WITH TOTALS;
`

// Get select metrics for specific queryid, hostname, etc.
// If totals = true, the function will return only totals and it will skip filters
// to differentiate it from empty filters.
func (m *MetricsRepository) Get(ctx context.Context, gArgs MetricsGetArgs) ([]M, error) {
	arg := map[string]interface{}{
		"period_start_from": gArgs.PeriodStartFromSec,
		"period_start_to":   gArgs.PeriodStartToSec,
	}

	tmplArgs := struct {
		PeriodStartFrom int64
		PeriodStartTo   int64
		PeriodDuration  int64
		Dimensions      map[string][]string
		Labels          map[string][]string
		DimensionVal    string
		Group           string
		Totals          bool
	}{
		PeriodStartFrom: gArgs.PeriodStartFromSec,
		PeriodStartTo:   gArgs.PeriodStartToSec,
		PeriodDuration:  gArgs.PeriodStartToSec - gArgs.PeriodStartFromSec,
		Dimensions:      EscapeColonsInMap(gArgs.Dimensions),
		Labels:          EscapeColonsInMap(gArgs.Labels),
		DimensionVal:    EscapeColons(gArgs.Filter),
		Group:           gArgs.Group,
		Totals:          gArgs.Totals,
	}
	var queryBuffer bytes.Buffer
	if tmpl, err := template.New("queryMetricsTmpl").Funcs(FuncMap).Parse(queryMetricsTmpl); err != nil {
		log.Fatalln(err)
	} else if err = tmpl.Execute(&queryBuffer, tmplArgs); err != nil {
		log.Fatalln(err)
	}
	var results []M
	query, args, err := sqlx.Named(queryBuffer.String(), arg)
	if err != nil {
		return results, errors.Wrap(err, cannotPrepare)
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return results, errors.Wrap(err, cannotPopulate)
	}
	query = m.db.Rebind(query)

	queryCtx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	rows, err := m.db.QueryxContext(queryCtx, query, args...)
	if err != nil {
		return results, errors.Wrap(err, CannotExecute)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		result := make(M)
		err = rows.MapScan(result)
		if err != nil {
			logrus.Errorf("DimensionMetrics Scan error: %v", err)
		}
		results = append(results, result)
	}
	rows.NextResultSet()
	total := make(M)
	for rows.Next() {
		err = rows.MapScan(total)
		if err != nil {
			logrus.Errorf("DimensionMetrics Scan TOTALS error: %v", err)
		}
		results = append(results, total)
	}

	return results, err
}

const queryMetricsTimeseries = `
SELECT num_queries,
       (m_query_time_sum/metrics.num_queries) AS m_query_time_avg_per_call,
       m_rows_sent_sum,
       m_shared_blks_read_sum,
       m_shared_blks_written_sum,
       m_shared_blks_hit_sum,
       period_start
FROM metrics
WHERE period_start >= :period_start_from AND period_start <= :period_start_to AND queryid = :queryid AND cluster = :cluster_name
ORDER BY period_start;
`

type QueryMetricPointDB struct {
	Timestamp           time.Time `json:"period_start"`
	QueryTimeAvgPerCall float64   `json:"m_query_time_avg_per_call"`
	NumQueries          int       `json:"num_queries"`
	RowSent             int       `json:"m_rows_sent_sum"`
	SharedBlocksRead    int       `json:"m_shared_blks_read_sum"`
	SharedBlocksWritten int       `json:"m_shared_blks_written_sum"`
	SharedBlocksHit     int       `json:"m_shared_blks_hit_sum"`
}

func (m *MetricsRepository) SelectQueryMetricsTimeseriesByQueryID(
	ctx context.Context,
	in MetricsGetArgs,
	queryID string,
	clusterName string,
) ([]QueryMetricPointDB, error) {
	arg := map[string]interface{}{
		"period_start_from": in.PeriodStartFromSec,
		"period_start_to":   in.PeriodStartToSec,
		"queryid":           queryID,
		"cluster_name":      clusterName,
	}

	var queryBuffer bytes.Buffer
	queryBuffer.WriteString(queryMetricsTimeseries)

	var results []QueryMetricPointDB
	query, args, err := processQuery(queryBuffer, arg)
	if err != nil {
		return nil, fmt.Errorf("could not process query: %v", err)
	}
	query = m.db.Rebind(query)

	queryCtx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	rows, err := m.db.QueryxContext(queryCtx, query, args...)
	if err != nil {
		return results, errors.Wrap(err, CannotExecute)
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		result := QueryMetricPointDB{}
		err = rows.StructScan(&result)
		if err != nil {
			logrus.Errorf("DimensionMetrics Scan error: %v", err)
		}
		results = append(results, result)
	}

	return results, err
}
