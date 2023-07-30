package info

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"os"
	"strings"
)

type Repository struct {
	DB *sqlx.DB
}

func (r Repository) GetActiveClusters(ctx context.Context) ([]Cluster, error) {
	clusters := make([]Cluster, 0)
	env, ok := os.LookupEnv("CLUSTERS")
	if !ok {
		return clusters, fmt.Errorf("no cluster/s present you maybe forget to set CLUSTERS environmental variable")
	}
	for _, name := range strings.Split(env, ",") {
		clusters = append(clusters, Cluster{
			ID:   name,
			Name: name,
		})
	}

	return clusters, nil
}
