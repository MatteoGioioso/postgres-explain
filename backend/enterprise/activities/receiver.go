package activities

import (
	"context"
	"github.com/borealis/monitoring-commons/proto"
	"github.com/sirupsen/logrus"
)

type ActivityCollectorService struct {
	proto.ActivityCollectorServer
	ActivitySampler *ActivitySampler
	Log             *logrus.Entry
}

func (s ActivityCollectorService) Collect(ctx context.Context, request *proto.ActivityCollectRequest) (*proto.ActivityCollectResponse, error) {
	s.Log.Infof("Received %+v activity samples\n", len(request.ActivitySamples))

	if err := s.ActivitySampler.Save(request); err != nil {
		return nil, err
	}

	return &proto.ActivityCollectResponse{}, nil
}
