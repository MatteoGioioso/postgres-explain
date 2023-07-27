package query_explainer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/borealisdb/commons/postgresql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"log"
	"postgres-explain/core/proto"
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
		postgresql.Options{ // TODO this should NOT user Admin user
			Database:        query.Database,
			SSLRootCertPath: "/borealis/root.crt",
			SSLDownload:     true,
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

	enrichedPlan, err := aps.enrichPlan(plan)
	if err != nil {
		return nil, fmt.Errorf("could not enrich plan: %v", err)
	}

	//stats, err := aps.getStatsFromPlan(plan)
	//if err != nil {
	//	return nil, fmt.Errorf("could get stats from plan: %v", err)
	//}

	return &proto.GetQueryPlanResponse{
		QueryId:   in.QueryId,
		QueryPlan: enrichedPlan,
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
		return "", err
	}
	defer tx.Rollback()

	rows, err := tx.Query(fmt.Sprintf("EXPLAIN (ANALYZE, COSTS, VERBOSE, BUFFERS, FORMAT JSON) %v", query.Query))
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			log.Fatal(err)
		}
		sb.WriteString(s)
		sb.WriteString("\n")
	}

	// In case of UPDATE or INSERT we don't want to change anything
	if err := tx.Rollback(); err != nil {
		return "", err
	}

	aps.log.Debugf("found plan %v for query %v", sb.String(), query.Query)

	return sb.String(), nil
}

func (aps *Service) enrichPlan(plan string) (string, error) {
	type Plans []struct {
		Plan map[string]interface{} `json:"Plan"`
	}

	p := Plans{}
	if err := json.Unmarshal([]byte(plan), &p); err != nil {
		return "", fmt.Errorf("could not unmarshal plan: %v", err)
	}
	NewPlanEnricher().AnalyzePlan(p[0].Plan)

	marshalledEnrichedPlan, err := json.Marshal(p[0].Plan)
	if err != nil {
		return "", fmt.Errorf("could not marshal enriched plan: %v", err)
	}

	return string(marshalledEnrichedPlan), nil
}

func (aps *Service) getStatsFromPlan(plan string) (string, error) {
	type Stats struct {
		ExecutionTime float64 `json:"Execution Time"`
		PlanningTime  float64 `json:"Planning Time"`
	}
	type Plans []Stats

	p := Plans{}
	if err := json.Unmarshal([]byte(plan), &p); err != nil {
		return "", fmt.Errorf("could not unmarshal plan: %v", err)
	}

	marshalledStats, err := json.Marshal(p[0])
	if err != nil {
		return "", fmt.Errorf("could not marshal enriched plan: %v", err)
	}

	return string(marshalledStats), nil
}
