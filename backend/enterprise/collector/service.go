package collector

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"postgres-explain/backend/cache"
	"postgres-explain/proto"
)

type Service struct {
	log         *logrus.Entry
	cacheClient *cache.Client

	proto.CollectorServer
}

func (s Service) Register(ctx context.Context, request *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	s.log.Infof("received registration from cluster %v and instance %v", request.ClusterName, request.InstanceName)

	if err := s.cacheClient.SetInstance(ctx, cache.Instance{
		ClusterName:   request.ClusterName,
		Name:          request.InstanceName,
		Host:          request.InstanceHost,
		CollectorHost: request.CollectorHost,
	}); err != nil {
		return &proto.RegisterResponse{}, fmt.Errorf("could not SetInstance in cache: %v", err)
	}

	return &proto.RegisterResponse{}, nil
}
