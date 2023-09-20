package info

import (
	"context"
	"github.com/borealisdb/commons/credentials"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/cache"
	"postgres-explain/proto"
)

type Repository struct {
	credentialsProvider credentials.Credentials
	log                 *logrus.Entry
	cacheClient         *cache.Client
}

func (r Repository) GetActiveClusters(ctx context.Context) ([]*proto.Cluster, error) {
	clusters, err := r.cacheClient.GetClusters(ctx)
	if err != nil {
		return nil, err
	}

	clustersProto := make([]*proto.Cluster, 0)
	for _, name := range clusters {
		cluster := &proto.Cluster{
			Id:     name,
			Name:   name,
			Status: statusOnline,
		}

		clustersProto = append(clustersProto, cluster)
	}

	return clustersProto, nil
}
