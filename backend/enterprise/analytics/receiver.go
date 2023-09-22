package analytics

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"postgres-explain/proto"
)

type Receiver struct {
	proto.StatementsCollectorServer
	MetricsBucket *MetricsBucket
	Log           *logrus.Entry
}

func (s Receiver) Collect(ctx context.Context, request *proto.StatementsCollectRequest) (*proto.StatementsCollectResponse, error) {
	s.Log.Infof("Received %+v statements samples", len(request.MetricsBucket))

	if err := s.MetricsBucket.Save(request); err != nil {
		return nil, fmt.Errorf("could not Save: %v", err)
	}

	return &proto.StatementsCollectResponse{}, nil
}
