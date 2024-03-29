package query_explainer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/matoous/go-nanoid/v2"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"postgres-explain/backend/enterprise/activities"
	"postgres-explain/backend/shared"
	"postgres-explain/core/pkg"
	"postgres-explain/proto"
	"time"
)

type Service struct {
	log            *logrus.Entry
	Repo           Repository
	ActivitiesRepo activities.Repository
	CommandsClient CommandsClient

	proto.QueryExplainerServer
}

func (aps *Service) GetOptimizationsList(ctx context.Context, request *proto.GetOptimizationsListRequest) (*proto.GetOptimizationsListResponse, error) {
	list, err := aps.Repo.GetOptimizations(ctx, PlansSearchRequest{
		PeriodStartFrom:  request.PeriodStartFrom.AsTime(),
		PeriodStartTo:    request.PeriodStartTo.AsTime(),
		ClusterName:      request.ClusterName,
		Limit:            int(request.Limit),
		Order:            request.Order,
		QueryFingerprint: request.QueryFingerprint,
		OptimizationId:   request.OptimizationId,
	})
	if err != nil {
		return nil, fmt.Errorf("could not GetPlansList: %v", err)
	}

	items := make([]*proto.PlanItem, 0)
	for _, entity := range list {
		var plan pkg.Explained
		if err := json.Unmarshal([]byte(entity.Plan), &plan); err != nil {
			return nil, fmt.Errorf("could not Unmarshal to pkg.Explained: %v", err)
		}

		items = append(items, &proto.PlanItem{
			Id:               entity.PlanID,
			Alias:            entity.Alias.String,
			PeriodStart:      timestamppb.New(entity.PeriodStart),
			Query:            entity.Query,
			OptimizationId:   entity.OptimizationId,
			QueryFingerprint: entity.QueryFingerprint,
			ExecutionTime:    float32(plan.Stats.ExecutionTime),
			PlanningTime:     float32(plan.Stats.PlanningTime),
		})
	}

	return &proto.GetOptimizationsListResponse{Plans: items}, nil
}

func (aps *Service) GetQueryPlansList(ctx context.Context, request *proto.GetQueryPlansListRequest) (*proto.GetQueryPlansListResponse, error) {
	list, err := aps.Repo.GetPlansList(ctx, PlansSearchRequest{
		PeriodStartFrom: request.PeriodStartFrom.AsTime(),
		PeriodStartTo:   request.PeriodStartTo.AsTime(),
		ClusterName:     request.ClusterName,
		Limit:           int(request.Limit),
		Order:           request.Order,
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
		})
	}

	return &proto.GetQueryPlansListResponse{Plans: items}, nil
}

func (aps *Service) SaveQueryPlan(ctx context.Context, request *proto.SaveQueryPlanRequest) (*proto.SaveQueryPlanResponse, error) {
	if err := aps.validateSaveRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	planRequest, err := aps.makePlanRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("could not makePlanRequest: %v", err)
	}

	plan, err := aps.CommandsClient.Explain(ctx, request.ClusterName, request.InstanceName, planRequest)
	if err != nil {
		return nil, fmt.Errorf("could not run explain: %v", err)
	}

	enrichedPlan, err := aps.processPlan(plan.Plan)
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

	planId, err := gonanoid.New(11)
	if err != nil {
		return nil, fmt.Errorf("could not generate nano id: %v", err)
	}

	planEntity := PlanEntity{
		Alias:            shared.ToSqlNullString(request.Alias),
		Query:            planRequest.Query,
		PlanID:           planId,
		QuerySha:         "",
		QueryFingerprint: fingerprint,
		OriginalPlan:     plan.Plan,
		ClusterName:      request.ClusterName,
		Database:         planRequest.Database,
		Plan:             string(marshalPlan),
		PeriodStart:      time.Now(),
		Username:         "",
	}

	if request.OptimizationId == "" {
		planEntity.OptimizationId = planId
	} else {
		planEntity.OptimizationId = request.OptimizationId
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
		PeriodStart:       timestamppb.New(plan.PeriodStart),
		OptimizationId:    plan.OptimizationId,
	}, err
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
	nodesStats := statsGather.ComputeNodesStats(node)
	tablesStats := statsGather.ComputeTablesStats(node)
	jitStats := statsGather.ComputeJITStats()
	triggersStats := statsGather.ComputeTriggersStats()
	summary := pkg.NewSummary().Do(node, stats)

	return pkg.Explained{
		Summary:       summary,
		Stats:         stats,
		IndexesStats:  indexesStats,
		TablesStats:   tablesStats,
		NodesStats:    nodesStats,
		JITStats:      jitStats,
		TriggersStats: triggersStats,
	}, nil
}

func (aps *Service) makePlanRequest(ctx context.Context, request *proto.SaveQueryPlanRequest) (*proto.PlanRequest, error) {
	planRequest := &proto.PlanRequest{
		InstanceName: request.InstanceName,
		Database:     request.Database,
		Query:        request.Query,
	}

	if request.QueryFingerprint != "" {
		queryMetadata, err := aps.ActivitiesRepo.GetQueryMetadataByFingerprint(ctx, request.QueryFingerprint)
		if err != nil {
			return nil, fmt.Errorf("could not GetQueryMetadataByFingerprint: %v", err)
		}
		if queryMetadata == nil {
			return nil, fmt.Errorf("query metadata not found for fingerprint %v", request.QueryFingerprint)
		}
		if queryMetadata.IsQueryTruncated == 1 && request.Query == "" {
			return nil, fmt.Errorf("query is truncated, cannot run an incomplete query")
		}

		if len(request.Parameters) > 0 {
			var err error
			planRequest.Query, err = shared.ConvertQueryWithParams(queryMetadata.ParsedQuery, paramsFromRequest(request.Parameters))
			if err != nil {
				return nil, fmt.Errorf("could not ConvertQueryWithParams: %v", err)
			}
		} else {
			planRequest.Query = queryMetadata.ParsedQuery
		}

		planRequest.Database = queryMetadata.Database
	}

	if request.QuerySha != "" {
		queryMetadata, err := aps.ActivitiesRepo.GetQueryMetadataBySha(ctx, request.QuerySha)
		if err != nil {
			return nil, fmt.Errorf("could not GetQueryMetadataBySha: %v", err)
		}
		if queryMetadata == nil {
			return nil, fmt.Errorf("query metadata not found for sha %v", request.QuerySha)
		}
		if queryMetadata.IsQueryTruncated == 1 && request.Query == "" {
			return nil, fmt.Errorf("query is truncated, cannot run an incomplete query")
		}

		planRequest.Query = queryMetadata.Query
		planRequest.Database = queryMetadata.Database
	}

	return planRequest, nil
}

func (aps *Service) validateSaveRequest(request *proto.SaveQueryPlanRequest) error {
	if request.ClusterName == "" {
		return fmt.Errorf("cluster_name is required")
	}

	if request.QueryFingerprint == "" && request.QuerySha == "" && request.Query == "" {
		return fmt.Errorf("at least one of the following property is required: query_fingerprint, query_id, query")
	}

	if (request.QueryFingerprint == "" && request.QuerySha == "") && request.Database == "" {
		return fmt.Errorf("if you are not suppling query_fingerprint or query_id, database is required to know where to run the query")
	}

	if request.QuerySha != "" && request.QueryFingerprint != "" {
		return fmt.Errorf("both query_sha and query_fingerprint cannot be present at the same time")
	}

	return nil
}

func paramsFromRequest(params []string) []interface{} {
	s := make([]interface{}, 0)
	for _, v := range params {
		s = append(s, v)
	}

	return s
}
