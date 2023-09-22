package activities

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"postgres-explain/backend/enterprise/shared"
	"time"
)

const waitEventProfilerSQLTemplate = `
SELECT toStartOfMinute(period_start) AS slot, 
       count() AS wait_event_count, 
       wait_event, 
       groupArray(cpu_cores)[1] as cpu_cores
FROM activities
WHERE period_start > :period_start_from 
  AND period_start < :period_start_to 
  AND cluster_name = :cluster_name
GROUP BY slot, wait_event
ORDER BY slot ASC;`

const topQueriesSQLTemplate = `
WITH final AS (WITH grouping AS (SELECT fingerprint,
                                        groupArray(cpu_cores)[1]  AS cc,
                                        (count() / :period_duration ) / cc AS cpu_load_by_wait_event,
                                        wait_event
                                 FROM activities
                                 WHERE period_start > :period_start_from
                                   AND period_start < :period_start_to 
                                   AND cluster_name = :cluster_name
                                 GROUP BY wait_event, fingerprint)
               SELECT fingerprint,
                      maxMap(map(wait_event, cpu_load_by_wait_event)) AS cpu_load_wait_events,
                      sum(cpu_load_by_wait_event)                     AS cpu_load_total
               FROM grouping
               GROUP BY fingerprint
               ORDER BY cpu_load_total DESC)
SELECT groupArray(acs.parsed_query)[1] AS parsed_query,
       groupArray(acs.query)[1]        AS query,
       cpu_load_total,
       cpu_load_wait_events,
       fingerprint
FROM final
         LEFT JOIN activities acs ON final.fingerprint = acs.fingerprint
GROUP BY cpu_load_wait_events, cpu_load_total, fingerprint
ORDER BY cpu_load_total DESC
LIMIT 25`

const topPropSQLTemplate = `
WITH grouping AS (SELECT {{ .Prop }},
                         groupArray(cpu_cores)[1] AS cc,
                         (count() / 60) / cc      AS cpu_load_by_wait_event,
                         wait_event
                  FROM bmserver.activities
                  WHERE period_start > :period_start_from 
                    AND period_start < :period_start_to 
                    AND cluster_name = :cluster_name
                  GROUP BY wait_event, {{ .Prop }})
SELECT {{ .Prop }} AS name,
       maxMap(map(wait_event, cpu_load_by_wait_event)) AS cpu_load_wait_events,
       sum(cpu_load_by_wait_event)                     AS cpu_load_total
FROM grouping
GROUP BY {{ .Prop }}
ORDER BY cpu_load_total DESC
LIMIT 25`

const topQueriesByIDTemplate = `
WITH grouping AS (SELECT query,
                         groupArray(cpu_cores)[1] AS cc,
                         (count() / :period_duration) / cc    AS cpu_load_by_wait_event,
                         wait_event
                  FROM bmserver.activities
                  WHERE period_start > :period_start_from 
                    AND period_start < :period_start_from 
                    AND cluster_name = :cluster_name
                    AND query_id = :query_id
                  GROUP BY wait_event, query)
SELECT query,
       maxMap(map(wait_event, cpu_load_by_wait_event)) AS cpu_load_wait_events,
       sum(cpu_load_by_wait_event)                     AS cpu_load_total
FROM grouping
GROUP BY query
ORDER BY cpu_load_total DESC;`

type QueryArgs struct {
	PeriodStartFromSec int64
	PeriodStartToSec   int64
	ClusterName        string
}

type SlotDB struct {
	Timestamp      time.Time `json:"slot"`
	WaitEventCount int       `json:"wait_event_count"`
	WaitEventName  string    `json:"wait_event"`
	CpuCores       float32   `json:"cpu_cores"`
}

type QueryDB struct {
	Fingerprint       string             `json:"fingerprint"`
	CPULoadWaitEvents map[string]float64 `json:"cpu_load_wait_events"`
	CPULoadTotal      float32            `json:"cpu_load_total"`
	ParsedQuery       sql.NullString     `json:"parsed_query"`
	Query             string             `json:"query"`
}

type PropDB struct {
	Name              string             `json:"name"`
	CPULoadWaitEvents map[string]float64 `json:"cpu_load_wait_events"`
	CPULoadTotal      float32            `json:"cpu_load_total"`
}

func (q QueryDB) GetSQL() string {
	if q.ParsedQuery.Valid && q.ParsedQuery.String != "" {
		return q.ParsedQuery.String
	}

	return q.Query
}

type ActivitiesRepository struct {
	DB *sqlx.DB
}

func NewActivitiesRepository(db *sqlx.DB) ActivitiesRepository {
	return ActivitiesRepository{DB: db}
}

func (ar ActivitiesRepository) Select(ctx context.Context, args QueryArgs) ([]SlotDB, error) {
	queryArgs := map[string]interface{}{
		"period_start_from": args.PeriodStartFromSec,
		"period_start_to":   args.PeriodStartToSec,
		"cluster_name":      args.ClusterName,
	}

	queryCtx, cancel := context.WithTimeout(ctx, shared.QueryTimeout)
	defer cancel()

	rows, err := ar.DB.NamedQueryContext(queryCtx, waitEventProfilerSQLTemplate, queryArgs)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	slots := make([]SlotDB, 0)
	for rows.Next() {
		slotDB := SlotDB{}
		if err := rows.StructScan(&slotDB); err != nil {
			return nil, err
		}

		slots = append(slots, slotDB)
	}

	return slots, err
}

func (ar ActivitiesRepository) GetQueriesByWaitEventCount(ctx context.Context, args QueryArgs) ([]QueryDB, error) {
	queryArgs := map[string]interface{}{
		"period_start_from": args.PeriodStartFromSec,
		"period_start_to":   args.PeriodStartToSec,
		"period_duration":   args.PeriodStartToSec - args.PeriodStartFromSec,
		"cluster_name":      args.ClusterName,
	}
	rows, err := ar.DB.NamedQueryContext(ctx, topQueriesSQLTemplate, queryArgs)
	if err != nil {
		return nil, err
	}

	rankedQueries := make([]QueryDB, 0)
	for rows.Next() {
		rankedQuery := QueryDB{}
		if err := rows.StructScan(&rankedQuery); err != nil {
			return nil, err
		}

		rankedQueries = append(rankedQueries, rankedQuery)
	}

	return rankedQueries, nil
}

func (ar ActivitiesRepository) GetTopWaitEventsLoadGroupByPropName(ctx context.Context, args QueryArgs, propName string) ([]PropDB, error) {
	queryArgs := map[string]interface{}{
		"period_start_from": args.PeriodStartFromSec,
		"period_start_to":   args.PeriodStartToSec,
		"cluster_name":      args.ClusterName,
	}

	tmplArgs := struct {
		Prop string
	}{
		Prop: propName,
	}
	query, vals, err := shared.ProcessQueryWithTemplate(tmplArgs, queryArgs, topPropSQLTemplate)
	if err != nil {
		return nil, err
	}

	var results []PropDB

	query = ar.DB.Rebind(query)

	queryCtx, cancel := context.WithTimeout(ctx, shared.QueryTimeout)
	defer cancel()

	rows, err := ar.DB.QueryxContext(queryCtx, query, vals...)
	if err != nil {
		return results, errors.Wrap(err, shared.CannotExecute)
	}
	defer rows.Close()

	for rows.Next() {
		result := PropDB{}
		err = rows.StructScan(&result)
		if err != nil {
			return nil, fmt.Errorf("PropDB Scan error: %v", err)
		}
		results = append(results, result)
	}

	return results, nil
}
