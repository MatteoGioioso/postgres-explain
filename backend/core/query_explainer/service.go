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
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (aps *Service) GetQueryPlansList(ctx context.Context, request *proto.GetQueryPlansListRequest) (*proto.GetQueryPlansListResponse, error) {
	list, err := aps.Repo.GetPlansList(ctx, PlansSearchRequest{
		PeriodStartFrom:  request.PeriodStartFrom.AsTime(),
		PeriodStartTo:    request.PeriodStartTo.AsTime(),
		ClusterName:      request.ClusterName,
		Limit:            int(request.Limit),
		Order:            request.Order,
		QueryFingerprint: request.QueryFingerprint,
		TrackingId:       request.TrackingId,
	})
	if err != nil {
		return nil, fmt.Errorf("could not GetPlansList: %v", err)
	}

	items := make([]*proto.PlanItem, 0)
	for _, entity := range list {
		items = append(items, &proto.PlanItem{
			Id:          entity.PlanID,
			Alias:       entity.Alias.String,
			PeriodStart: timestamppb.New(entity.PeriodStart),
			Query:       entity.Query,
			TrackingId:  entity.TrackingID,
		})
	}

	return &proto.GetQueryPlansListResponse{Plans: items}, nil
}

func (aps *Service) SaveQueryPlan(ctx context.Context, request *proto.SaveQueryPlanRequest) (*proto.SaveQueryPlanResponse, error) {
	pg := postgresql.V2{}
	pgCreds, err := aps.credentialsProvider.GetPostgresCredentials(ctx, request.ClusterName, "", credentials.Options{})
	if err != nil {
		return nil, fmt.Errorf("could not GetPostgresCredentials for cluster %v: %v", request.ClusterName, err)
	}
	endpoint, err := aps.credentialsProvider.GetClusterEndpoint(ctx, request.ClusterName, "")
	if err != nil {
		return nil, fmt.Errorf("could not GetClusterEndpoint for cluster %v: %v", request.ClusterName, err)
	}
	conn, err := pg.GetConnection(postgresql.Args{
		Username: pgCreds.Username,
		Password: pgCreds.Password,
		Database: request.Database,
		Port:     endpoint.Port,
		Host:     endpoint.Hostname,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get postgres connection pg.GetConnection: %v", err)
	}
	defer conn.Close()

	planRequest := PlanRequest{
		Query:    request.Query,
		QueryID:  request.QueryId,
		Database: request.Database,
	}
	if len(request.Parameters) > 0 {
		planRequest.paramsFromRequest(request.Parameters)
		planRequest.Query, err = shared.ConvertQueryWithParams(planRequest.Query, planRequest.Parameters)
		if err != nil {
			return nil, fmt.Errorf("could not ConvertQueryWithParams: %v", err)
		}
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
		TrackingID:       uuid.New().String(),
		QueryID:          shared.ToSqlNullString(planRequest.QueryID),
		QueryFingerprint: fingerprint,
		OriginalPlan:     plan,
		ClusterName:      request.ClusterName,
		Database:         planRequest.Database,
		Plan:             string(marshalPlan),
		PeriodStart:      time.Now(),
		Username:         pgCreds.Username,
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

	aps.log.Debugf("explaining: %v", query.Query)

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
