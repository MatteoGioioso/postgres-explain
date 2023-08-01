package query_explainer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/borealisdb/commons/postgresql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/shared"
	"postgres-explain/core/pkg"
	"postgres-explain/proto"
	"strings"
	"time"
)

type Service struct {
	log                 *logrus.Entry
	Repo                Repository
	credentialsProvider credentials.Credentials
	proto.QueryExplainerServer
}

func (aps *Service) SaveQueryPlan(ctx context.Context, request *proto.SaveQueryPlanRequest) (*proto.SaveQueryPlanResponse, error) {
	pg := postgresql.PG{CredentialsProvider: aps.credentialsProvider}
	conn, err := pg.GetConnection(ctx, request.ClusterName, "", postgresql.Options{
		Database: request.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get postgres connection pg.GetConnection: %v", err)
	}

	planRequest := PlanRequest{}

	// Custom query
	if request.QueryId == "" {
		planRequest.Query = request.Query
		planRequest.Database = request.Database
	} else {
		planRequest.QueryID = request.QueryId
		// Get query from the pg_stats_statements
		// planRequest.Query =
		// planRequest.Database =
	}

	plan, err := aps.runExplain(ctx, conn, planRequest)
	if err != nil {
		return nil, fmt.Errorf("could not run explain: %v", err)
	}

	aps.log.Debugf("found plan for query %v: \n %v", planRequest.Query, plan)

	enrichedPlan, err := aps.processPlan(plan)
	if err != nil {
		return nil, fmt.Errorf("could not enrich plan: %v", err)
	}

	marshalPlan, err := json.Marshal(enrichedPlan)
	if err != nil {
		return nil, fmt.Errorf("could not marshal plan: %v", err)
	}

	// Calculate fingerprint
	// https://pganalyze.com/blog/pg-query-2-0-postgres-query-parser#why-did-we-create-our-own-query-fingerprint-concept
	// https://github.com/pganalyze/pg_query_go
	fingerprint, err := pg_query.Fingerprint(planRequest.Query)
	if err != nil {
		return nil, fmt.Errorf("could not calculate query Fingerprint %v", err)
	}

	planEntity := PlanEntity{
		Query:            planRequest.Query,
		PlanID:           uuid.New().String(),
		QueryID:          shared.ToSqlNullString(planRequest.QueryID),
		QueryFingerprint: fingerprint,
		OriginalPlan:     plan,
		ClusterName:      request.ClusterName,
		Database:         planRequest.Database,
		Plan:             string(marshalPlan),
		PeriodStart:      time.Now(),
	}

	if err := aps.Repo.SaveQueryPlan(ctx, planEntity); err != nil {
		return nil, fmt.Errorf("could not SaveQueryPlan: %v", err)
	}

	return &proto.SaveQueryPlanResponse{PlanId: planEntity.PlanID}, nil
}

func (aps *Service) GetQueryPlan(ctx context.Context, request *proto.GetQueryPlanRequest) (*proto.GetQueryPlanResponse, error) {
	plan, err := aps.Repo.GetQueryPlan(ctx, request.PlanId)
	if err != nil {
		return nil, err
	}

	return &proto.GetQueryPlanResponse{
		PlanId:            plan.PlanID,
		QueryId:           plan.QueryID.String,
		QueryPlan:         plan.Plan,
		QueryOriginalPlan: plan.OriginalPlan,
		QueryFingerprint:  plan.QueryFingerprint,
		Query:             plan.Query,
	}, err
}

func (aps *Service) runExplain(
	ctx context.Context,
	conn *sqlx.DB,
	query PlanRequest,
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
