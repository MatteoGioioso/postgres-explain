package query_explainer

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("could not connectToClient: %v", err)
	}

	defer conn.Close()

	commandResponse, err := client.Command(ctx, &proto.CommandRequest{
		ActionType: proto.ActionTypes_EXPLAIN,
		Message:    &proto.CommandRequest_PlanRequest{PlanRequest: planRequest},
	})
	if err != nil {
		return nil, fmt.Errorf("could not run Command EXPLAIN: %v", err)
	}

	if commandResponse.ActionType == proto.ActionTypes_EXPLAIN {
		return commandResponse.GetPlanResponse(), nil
	}

	return nil, fmt.Errorf("could not retreive response for command EXPLAIN")
}

func (c CommandsClient) connectToClient(ctx context.Context, clusterName, instanceName string) (proto.CommandsClient, *grpc.ClientConn, error) {
	instance, err := c.cacheClient.GetInstance(ctx, clusterName, instanceName)
	if err != nil {
		return nil, nil, fmt.Errorf("could not GetInstance: %v", err)
	}

	c.log.Infof("connecting to collector host %v", instance.CollectorHost)

	grpcConn, err := grpc.Dial(instance.CollectorHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("could not Dial collector %v: %v", instance.CollectorHost, err)
	}

	return proto.NewCommandsClient(grpcConn), grpcConn, nil
}
