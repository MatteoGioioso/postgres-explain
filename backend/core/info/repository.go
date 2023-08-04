package info

import (
	"context"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	"github.com/borealisdb/commons/postgresql"
	"github.com/sirupsen/logrus"
	"os"
	"postgres-explain/proto"
	"strings"
)

type Repository struct {
	credentialsProvider credentials.Credentials
	log                 *logrus.Entry
}

func (r Repository) GetActiveClusters(ctx context.Context) ([]*proto.Cluster, error) {
	pg := postgresql.V2{}
	clusters := make([]*proto.Cluster, 0)
	env, ok := os.LookupEnv("CLUSTERS")
	if !ok {
		return clusters, fmt.Errorf("no cluster/s present you maybe forget to set CLUSTERS environmental variable")
	}
	for _, name := range strings.Split(env, ",") {
		endpoint, err := r.credentialsProvider.GetClusterEndpoint(ctx, name, "")
		if err != nil {
			return nil, fmt.Errorf("could not GetClusterEndpoint for cluster %v: %v", name, err)
		}
		creds, err := r.credentialsProvider.GetPostgresCredentials(ctx, name, "", credentials.Options{})
		if err != nil {
			return nil, fmt.Errorf("could not GetPostgresCredentials for cluster %v: %v", name, err)
		}

		cluster := &proto.Cluster{
			Id:       name,
			Name:     name,
			Hostname: endpoint.Hostname,
			Port:     endpoint.Port,
			Status:   statusOnline,
		}

		conn, err := pg.GetConnection(postgresql.Args{
			Username: creds.Username,
			Password: creds.Password,
			Port:     endpoint.Port,
			Host:     endpoint.Hostname,
		})
		if err != nil {
			r.log.Errorf(
				"could not GetConnection for cluster %v with host %v:%v : %v",
				name,
				endpoint.Hostname,
				endpoint.Port,
				err,
			)

			cluster.Status = statusOffline
			cluster.StatusError = err.Error()
		}
		if err := conn.Ping(); err != nil {
			r.log.Errorf(
				"could not Ping cluster %v with host %v:%v : %v",
				name,
				endpoint.Hostname,
				endpoint.Port,
				err,
			)

			cluster.Status = statusOffline
			cluster.StatusError = err.Error()
		}

		clusters = append(clusters, cluster)
	}

	return clusters, nil
}
