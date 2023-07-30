package query_explainer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/borealisdb/commons/postgresql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/core/pkg"
	"postgres-explain/proto"
	"strings"
)

type Service struct {
	log                 *logrus.Entry
	Repo                Repository
	credentialsProvider credentials.Credentials
	proto.QueryExplainerServer
}

func (aps *Service) GetQueryPlan(ctx context.Context, in *proto.GetQueryPlanRequest) (*proto.GetQueryPlanResponse, error) {
	if in.PeriodStartFrom == nil || in.PeriodStartTo == nil {
		return nil, fmt.Errorf("from-date: %s or to-date: %s cannot be empty", in.PeriodStartFrom, in.PeriodStartTo)
	}

	periodStartFromSec := in.PeriodStartFrom.Seconds
	periodStartToSec := in.PeriodStartTo.Seconds
	if periodStartFromSec > periodStartToSec {
		return nil, fmt.Errorf("from-date %s cannot be bigger then to-date %s", in.PeriodStartFrom, in.PeriodStartTo)
	}

	query, err := aps.Repo.GetLongestQueryByID(
		ctx,
		QueryArgs{
			PeriodStartFromSec: periodStartFromSec,
			PeriodStartToSec:   periodStartToSec,
			ClusterName:        in.ClusterName,
		},
		in.QueryId,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get longest query from clickhouse: %v", err)
	}

	pg := postgresql.PG{CredentialsProvider: aps.credentialsProvider}
	conn, err := pg.GetConnection(
		ctx,
		in.ClusterName,
		postgresql.Options{
			Database: query.Database,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not get postgres connection pg.GetConnection: %v", err)
	}

	plan, err := aps.runExplain(ctx, conn, query)
	if err != nil {
		return nil, fmt.Errorf("could not run explain: %v", err)
	}

	aps.log.Debugf("found plan for query %v: \n %v", query.Query, plan)

	enrichedPlan, err := aps.processPlan(plan)
	if err != nil {
		return nil, fmt.Errorf("could not enrich plan: %v", err)
	}

	marshalPlan, err := json.Marshal(enrichedPlan)
	if err != nil {
		return nil, fmt.Errorf("could not marshal plan: %v", err)
	}

	return &proto.GetQueryPlanResponse{
		QueryId:   in.QueryId,
		QueryPlan: string(marshalPlan),
	}, nil
}

func (aps *Service) runExplain(
	ctx context.Context,
	conn *sqlx.DB,
	query LongestQueryDB,
) (string, error) {
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("could not run transaction: %v", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query(fmt.Sprintf("EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON) %v", query.Query))
	if err != nil {
		return "", fmt.Errorf("could not run EXPLAIN query: %v", err)
	}

	var sb strings.Builder

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return "", fmt.Errorf("could not scan row: %v", err)
		}
		sb.WriteString(s)
		sb.WriteString("\n")
	}

	// In case of UPDATE, DELETE or INSERT we don't want to persist the changes
	if err := tx.Rollback(); err != nil {
		return "", fmt.Errorf("could not roll back transaction: %v", err)
	}

	aps.log.Debugf("found plan %v for query %v", sb.String(), query.Query)

	return sb.String(), nil
}

func (aps *Service) processPlan(plan string) (pkg.Explained, error) {
	node, err := pkg.GetRootNodeFromPlans(plan)
	if err != nil {
		return pkg.Explained{}, fmt.Errorf("could not get root node from plan: %v", err)
	}

	pkg.NewPlanEnricher().AnalyzePlan(node)

	statsGather := pkg.NewStatsGather()
	if err := statsGather.GetStatsFromPlans(plan); err != nil {
		return pkg.Explained{}, fmt.Errorf("could not get stats from plan from plan: %v", err)
	}

	stats := statsGather.ComputeStats(node)
	indexesStats := statsGather.ComputeIndexesStats(node)
	summary := pkg.NewSummary().Do(node, stats)

	return pkg.Explained{
		Summary:      summary,
		Stats:        stats,
		IndexesStats: indexesStats,
	}, nil
}
