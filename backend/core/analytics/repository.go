package analytics

import (
	"context"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/borealisdb/commons/postgresql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/shared"
)

type Repository struct {
	conn                *sqlx.DB
	log                 *logrus.Entry
	credentialsProvider credentials.Credentials
}

const getQueryMetricsTmpl = `
SELECT queryid,
       query,
       plans,
       total_plan_time,
       min_plan_time,
       max_plan_time,
       mean_plan_time,
       stddev_plan_time,
       calls,
       total_exec_time,
       min_exec_time,
       max_exec_time,
       mean_exec_time,
       stddev_exec_time,
       rows,
       shared_blks_hit,
       shared_blks_read,
       shared_blks_dirtied,
       shared_blks_written,
       local_blks_hit,
       local_blks_read,
       local_blks_dirtied,
       local_blks_written,
       temp_blks_read,
       temp_blks_written,
       blk_read_time,
       blk_write_time,
       temp_blk_read_time,
       temp_blk_write_time,
       wal_records,
       wal_fpi,
       wal_bytes,
       jit_functions,
       jit_generation_time,
       jit_inlining_count,
       jit_inlining_time,
       jit_optimization_count,
       jit_optimization_time,
       jit_emission_count,
       jit_emission_time
FROM pg_stat_statements
ORDER BY {{.OrderBy}} {{.OrderDir}}
LIMIT :limit
`

func (r Repository) GetMetrics(ctx context.Context, request QueriesMetricsRequest) ([]MetricsEntity, error) {
	pg := postgresql.V2{}
	pgCreds, err := r.credentialsProvider.GetPostgresCredentials(ctx, request.ClusterName, "", credentials.Options{})
	if err != nil {
		return nil, fmt.Errorf("could not GetPostgresCredentials for cluster %v: %v", request.ClusterName, err)
	}
	endpoint, err := r.credentialsProvider.GetClusterEndpoint(ctx, request.ClusterName, "")
	if err != nil {
		return nil, fmt.Errorf("could not GetClusterEndpoint for cluster %v: %v", request.ClusterName, err)
	}
	conn, err := pg.GetConnection(postgresql.Args{
		Username: pgCreds.Username,
		Password: pgCreds.Password,
		Port:     endpoint.Port,
		Host:     endpoint.Hostname,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get postgres connection pg.GetConnection: %v", err)
	}
	defer conn.Close()

	query, queryArgs, err := shared.ProcessQueryWithTemplate(request.GetTemplateArgs(), request.GetQueryArgs(), getQueryMetricsTmpl)
	if err != nil {
		return nil, err
	}

	query = conn.Rebind(query)

	r.log.Debugf("query: %v, args: %v", query, queryArgs)

	queryCtx, cancel := context.WithTimeout(ctx, shared.QueryTimeout)
	defer cancel()

	rows, err := conn.QueryxContext(queryCtx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("QueryxContext error: %v", err)
	}
	defer rows.Close()

	queriesMetrics := make([]MetricsEntity, 0)
	for rows.Next() {
		metrics := MetricsEntity{}
		if err := rows.MapScan(metrics); err != nil {
			return nil, fmt.Errorf("could not MapScan to MetricsEntity: %v", err)
		}
		queriesMetrics = append(queriesMetrics, metrics)
	}

	return queriesMetrics, nil
}
