package query_explainer

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"postgres-explain/backend/cache"
	"postgres-explain/proto"
)

type CommandsClient struct {
	log         *logrus.Entry
	cacheClient *cache.Client
}

func (c CommandsClient) Explain(ctx context.Context, clusterName, instanceName string, planRequest *proto.PlanRequest) (*proto.PlanResponse, error) {
	client, conn, err := c.connectToClient(ctx, clusterName, instanceName)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	commandResponse, err := client.Command(ctx, &proto.CommandRequest{
		ActionType: proto.ActionTypes_EXPLAIN,
		Message:    &proto.CommandRequest_PlanRequest{PlanRequest: planRequest},
	})
	if err != nil {
		return nil, err
	}

	if commandResponse.ActionType == proto.ActionTypes_EXPLAIN {
		return commandResponse.GetPlanResponse(), nil
	}

	return nil, nil
}

func (c CommandsClient) connectToClient(ctx context.Context, clusterName, instanceName string) (proto.CommandsClient, *grpc.ClientConn, error) {
	instance, err := c.cacheClient.GetInstance(ctx, clusterName, instanceName)
	if err != nil {
		return nil, nil, err
	}
	grpcConn, err := grpc.Dial(instance.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	return proto.NewCommandsClient(grpcConn), grpcConn, nil
}
