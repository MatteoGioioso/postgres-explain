package info

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"postgres-explain/backend/cache"
	"postgres-explain/proto"
)

type Service struct {
	log         *logrus.Entry
	cacheClient *cache.Client

	proto.InfoServer
}

func (aps *Service) GetClusterInstances(ctx context.Context, request *proto.GetClusterInstancesRequest) (*proto.GetClusterInstancesResponse, error) {
	instances, err := aps.cacheClient.GetClusterInstances(ctx, request.ClusterName)
	if err != nil {
		return &proto.GetClusterInstancesResponse{}, fmt.Errorf("could not GetClusterInstances: %v", err)
	}

	var instancesProto []*proto.Instance
	for _, instance := range instances {
		hostname, port, err := net.SplitHostPort(instance.Host)
		if err != nil {
			return &proto.GetClusterInstancesResponse{}, fmt.Errorf("could not SplitHostPort: %v", err)
		}

		instancesProto = append(instancesProto, &proto.Instance{
			Id:          instance.Name,
			Name:        instance.Name,
			Hostname:    hostname,
			Port:        port,
			Status:      "",
			StatusError: "",
		})
	}

	return &proto.GetClusterInstancesResponse{Instances: instancesProto}, nil
}

func (aps *Service) GetClusters(ctx context.Context, in *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	clusters, err := aps.cacheClient.GetClusters(ctx)
	if err != nil {
		return &proto.GetClustersResponse{}, err
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

	return &proto.GetClustersResponse{
		Clusters: clustersProto,
	}, nil
}
