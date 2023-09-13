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
   tracking_id
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
   tracking_id,
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
    :tracking_id,
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
SELECT id, alias, period_start, query, tracking_id 
FROM plans 
WHERE cluster = :cluster
ORDER BY :order_by {{ .OrderDir }} 
LIMIT :limit
`

func (ar Repository) GetPlansList(ctx context.Context, request PlansSearchRequest) ([]PlanEntity, error) {
	query, queryArgs, err := shared.ProcessQueryWithTemplate(request.ToTmplArgs(), request.ToQueryArgs(), getPlansTmpl)
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
