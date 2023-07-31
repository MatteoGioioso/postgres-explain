package info

import (
	"context"
	"github.com/borealisdb/commons/credentials"
	"github.com/sirupsen/logrus"
	"postgres-explain/proto"
)

type Service struct {
	log                 *logrus.Entry
	Repo                Repository
	credentialsProvider credentials.Credentials
	proto.QueryExplainerServer
}

func (aps *Service) GetClusters(ctx context.Context, in *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	clusters, err := aps.Repo.GetActiveClusters(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.GetClustersResponse{
		Clusters: toClusters(clusters),
	}, nil
}

func toClusters(clusters []Cluster) []*proto.Clusters {
	cls := make([]*proto.Clusters, 0)
	for _, cluster := range clusters {
		cls = append(cls, &proto.Clusters{
			Id:   cluster.ID,
			Name: cluster.Name,
		})
	}

	return cls
}
