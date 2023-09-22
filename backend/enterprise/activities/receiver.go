package activities

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"postgres-explain/proto"
)

type ActivityCollectorService struct {
	proto.ActivityCollectorServer
	ActivitySampler *ActivitySampler
	Log             *logrus.Entry
}

func (s ActivityCollectorService) Collect(ctx context.Context, request *proto.ActivityCollectRequest) (*proto.ActivityCollectResponse, error) {
	s.Log.Infof("Received %+v activity samples", len(request.ActivitySamples))

	if err := s.ActivitySampler.Save(request); err != nil {
		return nil, fmt.Errorf("could not Save: %v", err)
	}

	return &proto.ActivityCollectResponse{}, nil
}
