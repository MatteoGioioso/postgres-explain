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

func (ar Repository) GetLongestQueryByID(ctx context.Context, args QueryArgs, queryID string) (LongestQueryDB, error) {
	queryArgs := map[string]interface{}{
		"period_start_from": args.PeriodStartFromSec,
		"period_start_to":   args.PeriodStartToSec,
		"query_id":          queryID,
		"cluster_name":      args.ClusterName,
	}
	rows, err := ar.DB.NamedQueryContext(ctx, longestQueryByID, queryArgs)
	if err != nil {
		return LongestQueryDB{}, err
	}
	longestQueriesDB := make([]LongestQueryDB, 0)
	for rows.Next() {
		longestQueryDB := LongestQueryDB{}
		if err := rows.StructScan(&longestQueryDB); err != nil {
			return LongestQueryDB{}, err
		}

		longestQueriesDB = append(longestQueriesDB, longestQueryDB)
	}

	if len(longestQueriesDB) == 0 {
		return LongestQueryDB{}, fmt.Errorf("query with id: %v, not found", queryID)
	}

	return longestQueriesDB[0], nil
}

type LongestQueryDB struct {
	Query    string  `json:"query"`
	ID       string  `json:"query_id"`
	Database string  `json:"datname"`
	Duration float32 `json:"duration"`
}

type QueryArgs struct {
	PeriodStartFromSec int64
	PeriodStartToSec   int64
	ClusterName        string
}
