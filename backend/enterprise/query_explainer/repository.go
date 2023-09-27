package query_explainer

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/shared"
)

type Repository struct {
	DB  *sqlx.DB
	Log *logrus.Entry
}

const selectQueryPlan = `
SELECT 	id, 
   alias,
   query_fingerprint,
   queryid,
   plan,
   original_plan,
   query,
   database,
   username,
   cluster,
   period_start,
   optimization_id
FROM plans
WHERE id = :plan_id;`

func (ar Repository) GetQueryPlan(ctx context.Context, planID string) (PlanEntity, error) {
	queryArgs := map[string]interface{}{
		"plan_id": planID,
	}
	rows, err := ar.DB.NamedQueryContext(ctx, selectQueryPlan, queryArgs)
	if err != nil {
		return PlanEntity{}, fmt.Errorf("could not NamedQueryContext for selectQueryPlan: %v", err)
	}

	planEntities := make([]PlanEntity, 0)
	for rows.Next() {
		planEntity := PlanEntity{}
		if err := rows.StructScan(&planEntity); err != nil {
			return PlanEntity{}, fmt.Errorf("could not scan struct PlanEntity: %v", err)
		}

		planEntities = append(planEntities, planEntity)
	}

	if len(planEntities) == 0 {
		return PlanEntity{}, fmt.Errorf("plan with id: %v, not found", planID)
	}

	return planEntities[0], nil
}

const insertQueryPlan = `
  INSERT INTO plans
  (
	id, 
   alias,
   query_fingerprint,
   queryid,
   plan,
   original_plan,
   query,
   database,
   username,
   cluster,
   period_start,
   optimization_id
   )
VALUES (
    :id,
	:alias,
	:query_fingerprint,
	:queryid,
	:plan,
	:original_plan,
	:query,
	:database,
	:username,
	:cluster,
	:period_start,
	:optimization_id
  )
`

func (ar Repository) SaveQueryPlan(ctx context.Context, entity PlanEntity) error {
	query, args, err := ar.DB.BindNamed(insertQueryPlan, entity)
	if err != nil {
		return fmt.Errorf("could not BindNamed for insertQueryPlan with PlanEntity: %v", err)
	}

	if _, err := ar.DB.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("could not execute statement: %v", err)
	}

	return nil
}

const getPlansTmpl = `
SELECT id, alias, period_start, query, optimization_id, query_fingerprint 
FROM plans 
WHERE cluster = :cluster
ORDER BY {{ .OrderBy}} {{ .OrderDir }} 
LIMIT :limit
`

func (ar Repository) GetPlansList(ctx context.Context, request PlansSearchRequest) ([]PlanEntity, error) {
	return ar.getPlansList(ctx, request, getPlansTmpl)
}

const getOptimizationsTmpl = `
SELECT id, alias, period_start, query, optimization_id, query_fingerprint, plan
FROM plans 
WHERE cluster = :cluster AND (query_fingerprint = :query_fingerprint OR optimization_id = :optimization_id)
ORDER BY {{ .OrderBy}} {{ .OrderDir }} 
LIMIT :limit
`

func (ar Repository) GetOptimizations(ctx context.Context, request PlansSearchRequest) ([]PlanEntity, error) {
	return ar.getPlansList(ctx, request, getOptimizationsTmpl)
}

func (ar Repository) getPlansList(ctx context.Context, request PlansSearchRequest, queryTemplate string) ([]PlanEntity, error) {
	query, queryArgs, err := shared.ProcessQueryWithTemplate(request.ToTmplArgs(), request.ToQueryArgs(), queryTemplate)
	if err != nil {
		return nil, fmt.Errorf("could not ProcessQueryWithTemplate: %v", err)
	}

	query = ar.DB.Rebind(query)

	ar.Log.Debugf("query: %v, args: %v", query, queryArgs)

	queryCtx, cancel := context.WithTimeout(ctx, shared.QueryTimeout)
	defer cancel()

	rows, err := ar.DB.QueryxContext(queryCtx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("QueryxContext error: %v", err)
	}
	defer rows.Close()

	plans := make([]PlanEntity, 0)
	for rows.Next() {
		planEntity := PlanEntity{}
		if err := rows.StructScan(&planEntity); err != nil {
			return nil, fmt.Errorf("could not StructScan, PlanEntity: %v", err)
		}
		plans = append(plans, planEntity)
	}

	return plans, nil
}

const getQueryMetadataByFingerprintTmpl = `SELECT datname, parsed_query FROM activities WHERE fingerprint = :fingerprint LIMIT 1`
const getQueryMetadataByQueryIDTmpl = `SELECT datname, query FROM activities WHERE query_id = :query_id LIMIT 1;`

func (ar Repository) GetQueryMetadataByFingerprint(ctx context.Context, fingerprint string) (*QueryMetadata, error) {
	metadata, err := ar.getQueryMetadata(ctx, getQueryMetadataByFingerprintTmpl, struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, fmt.Errorf("could not getQueryMetadata with fingerprint template: %v", err)
	}

	if len(metadata) > 0 {
		return metadata[0], nil
	}

	return nil, nil
}

func (ar Repository) GetQueryMetadataByID(ctx context.Context, id string) (*QueryMetadata, error) {
	metadata, err := ar.getQueryMetadata(ctx, getQueryMetadataByQueryIDTmpl, struct {
		ID string `json:"query_id"`
	}{
		ID: id,
	})
	if err != nil {
		return nil, fmt.Errorf("could not getQueryMetadata with id template: %v", err)
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
