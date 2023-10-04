package info

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (aps *Service) GetDatabases(ctx context.Context, in *proto.GetDatabasesRequest) (*proto.GetDatabasesResponse, error) {
	client, grpcConn, instanceName, err := aps.connectToClient(ctx, in.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("could not connectToClient: %v", err)
	}

	defer grpcConn.Close()

	commandResponse, err := client.Command(ctx, &proto.CommandRequest{
		ActionType: proto.ActionTypes_GET_DATABASES,
		Message: &proto.CommandRequest_GetDatabasesRequest{GetDatabasesRequest: &proto.GetDatabasesCommandRequest{
			InstanceName: instanceName,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("could not run Command GET_DATABASES: %v", err)
	}

	if commandResponse.ActionType == proto.ActionTypes_GET_DATABASES {
		databases := make([]*proto.Database, 0)
		for _, db := range commandResponse.GetGetDatabasesResponse().Databases {
			databases = append(databases, &proto.Database{Name: db.Name})
		}

		return &proto.GetDatabasesResponse{Databases: databases}, nil
	}

	return nil, fmt.Errorf("could not retreive response for command GET_DATABASES")
}

func (aps *Service) connectToClient(ctx context.Context, clusterName string) (proto.CommandsClient, *grpc.ClientConn, string, error) {
	clusterInstances, err := aps.cacheClient.GetClusterInstances(ctx, clusterName)
	if err != nil {
		return nil, nil, "", fmt.Errorf("could not GetInstance: %v", err)
	}

	if len(clusterInstances) == 0 {
		return nil, nil, "", fmt.Errorf("no instances found with cluster name %v", clusterName)
	}

	// Get any instance, it does not matter which one in this case
	var instance cache.Instance
	for _, inst := range clusterInstances {
		instance = inst
		break
	}

	aps.log.Infof("connecting to collector host %v", instance.CollectorHost)

	grpcConn, err := grpc.Dial(instance.CollectorHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, "", fmt.Errorf("could not Dial collector %v: %v", instance.CollectorHost, err)
	}

	return proto.NewCommandsClient(grpcConn), grpcConn, instance.Name, nil
}
