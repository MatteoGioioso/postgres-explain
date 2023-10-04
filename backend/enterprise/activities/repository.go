package activities

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"postgres-explain/backend/enterprise/shared"
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

type Repository struct {
	DB *sqlx.DB
}

func NewActivitiesRepository(db *sqlx.DB) Repository {
	return Repository{DB: db}
}

func (ar Repository) Select(ctx context.Context, args QueryArgs) ([]SlotDB, error) {
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
       fingerprint,
       groupArray(acs.is_query_truncated)[1] AS is_query_truncated,
       groupArray(acs.is_not_explainable)[1] AS is_query_not_explainable
FROM final
         LEFT JOIN activities acs ON final.fingerprint = acs.fingerprint
GROUP BY cpu_load_wait_events, cpu_load_total, fingerprint
ORDER BY cpu_load_total DESC
LIMIT 25`

func (ar Repository) GetQueriesByWaitEventCount(ctx context.Context, args QueryArgs) ([]QueryDB, error) {
	queryArgs := map[string]interface{}{
		"period_start_from": args.PeriodStartFromSec,
		"period_start_to":   args.PeriodStartToSec,
		"period_duration":   args.PeriodStartToSec - args.PeriodStartFromSec,
		"cluster_name":      args.ClusterName,
	}
	return ar.getTopQueries(ctx, queryArgs, topQueriesSQLTemplate)
}

const getTopQueriesByFingerprintTmpl = `WITH final AS (WITH grouping AS (SELECT query,
                                        groupArray(cpu_cores)[1] AS cc,
                                        (count() / :period_duration) / cc    AS cpu_load_by_wait_event,
                                        wait_event
                                 FROM activities
                                 WHERE period_start > :period_start_from
                                   AND period_start < :period_start_to
                                   AND cluster_name = :cluster_name
                                   AND fingerprint = :fingerprint
                                 GROUP BY wait_event, query)
               SELECT query,
                      maxMap(map(wait_event, cpu_load_by_wait_event)) AS cpu_load_wait_events,
                      sum(cpu_load_by_wait_event)                     AS cpu_load_total
               FROM grouping
               GROUP BY query
               ORDER BY cpu_load_total DESC)
SELECT acs.query                             AS query,
       cpu_load_total,
       cpu_load_wait_events,
       groupArray(acs.is_query_truncated)[1] AS is_query_truncated,
       acs.query_sha                         AS query_sha,
       groupArray(acs.is_not_explainable)[1] AS is_query_not_explainable,
       fingerprint
FROM final
         LEFT JOIN activities acs ON final.query = acs.query
GROUP BY cpu_load_wait_events, cpu_load_total, fingerprint, acs.query_sha, acs.query
ORDER BY cpu_load_total DESC
LIMIT 25`

func (ar Repository) GetTopQueriesByFingerprint(ctx context.Context, args QueryArgs) ([]QueryDB, error) {
	queryArgs := map[string]interface{}{
		"period_start_from": args.PeriodStartFromSec,
		"period_start_to":   args.PeriodStartToSec,
		"period_duration":   args.PeriodStartToSec - args.PeriodStartFromSec,
		"cluster_name":      args.ClusterName,
		"fingerprint":       args.Fingerprint,
	}

	return ar.getTopQueries(ctx, queryArgs, getTopQueriesByFingerprintTmpl)
}

func (ar Repository) getTopQueries(ctx context.Context, args map[string]interface{}, tmpl string) ([]QueryDB, error) {
	rows, err := ar.DB.NamedQueryContext(ctx, tmpl, args)
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

const getQueryMetadataByFingerprintTmpl = `SELECT datname, parsed_query, is_query_truncated FROM activities WHERE fingerprint = :fingerprint LIMIT 1`
const getQueryMetadataByShaTmpl = `SELECT datname, query, is_query_truncated FROM activities WHERE query_sha = :query_sha LIMIT 1`

func (ar Repository) GetQueryMetadataByFingerprint(ctx context.Context, fingerprint string) (*QueryMetadata, error) {
	metadata, err := ar.getQueryMetadata(ctx, getQueryMetadataByFingerprintTmpl, struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, fmt.Errorf("could not getQueryMetadata with fingerprint %v: %v", fingerprint, err)
	}

	if len(metadata) > 0 {
		return metadata[0], nil
	}

	return nil, nil
}

func (ar Repository) GetQueryMetadataBySha(ctx context.Context, sha string) (*QueryMetadata, error) {
	metadata, err := ar.getQueryMetadata(ctx, getQueryMetadataByShaTmpl, struct {
		Sha string `json:"query_sha"`
	}{
		Sha: sha,
	})
	if err != nil {
		return nil, fmt.Errorf("could not getQueryMetadata with sha %v: %v", sha, err)
	}

	if len(metadata) > 0 {
		return metadata[0], nil
	}

	return nil, nil
}

func (ar Repository) getQueryMetadata(ctx context.Context, tmpl string, args interface{}) ([]*QueryMetadata, error) {
	queryCtx, cancel := context.WithTimeout(ctx, shared.QueryTimeout)
	defer cancel()

	rows, err := ar.DB.NamedQueryContext(queryCtx, tmpl, args)
	if err != nil {
		return nil, fmt.Errorf("could not NamedQueryContext: %v", err)
	}

	defer rows.Close()

	metadata := make([]*QueryMetadata, 0)
	for rows.Next() {
		meta := &QueryMetadata{}
		if err := rows.StructScan(&meta); err != nil {
			return nil, fmt.Errorf("could not StructScan: %v", err)
		}

		metadata = append(metadata, meta)
	}

	return metadata, nil
}
