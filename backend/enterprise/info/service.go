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
	proto.InfoServer
}

func (aps *Service) GetClusterInstances(ctx context.Context, request *proto.GetClusterInstancesRequest) (*proto.GetClusterInstancesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (aps *Service) GetClusters(ctx context.Context, in *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	clusters, err := aps.Repo.GetActiveClusters(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.GetClustersResponse{
		Clusters: clusters,
	}, nil
}
