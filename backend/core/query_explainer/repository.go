package query_explainer

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
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
   period_start
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
   period_start
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
	:period_start
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
