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

	return &proto.GetClustersResponse{}, nil
}
