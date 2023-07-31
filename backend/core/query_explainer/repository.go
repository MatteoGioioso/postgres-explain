package query_explainer

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

const longestQueryByID = `
SELECT query_id,
       query, 
       duration,
       datname
FROM activities
WHERE period_start > :period_start_from 
  AND period_start < :period_start_to 
  AND query_id = :query_id
  AND cluster_name = :cluster_name
ORDER BY duration DESC 
LIMIT 1;`

type Repository struct {
	DB *sqlx.DB
}

func (ar Repository) GetLongestQueryByID(ctx context.Context, args QueryArgs, queryID string) (PlanRequest, error) {
	queryArgs := map[string]interface{}{
		"period_start_from": args.PeriodStartFromSec,
		"period_start_to":   args.PeriodStartToSec,
		"query_id":          queryID,
		"cluster_name":      args.ClusterName,
	}
	rows, err := ar.DB.NamedQueryContext(ctx, longestQueryByID, queryArgs)
	if err != nil {
		return PlanRequest{}, err
	}
	longestQueriesDB := make([]PlanRequest, 0)
	for rows.Next() {
		longestQueryDB := PlanRequest{}
		if err := rows.StructScan(&longestQueryDB); err != nil {
			return PlanRequest{}, err
		}

		longestQueriesDB = append(longestQueriesDB, longestQueryDB)
	}

	if len(longestQueriesDB) == 0 {
		return PlanRequest{}, fmt.Errorf("query with id: %v, not found", queryID)
	}

	return longestQueriesDB[0], nil
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
   schema,
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
	:schema,
	:username,
	:cluster,
	:period_start
  )
`

func (ar Repository) SaveQueryPlan(ctx context.Context, entity PlanEntity) error {
	stmt, err := ar.DB.PrepareNamed(insertQueryPlan)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}

	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, entity); err != nil {
		return fmt.Errorf("could not execute statement: %v", err)
	}

	return nil
}
