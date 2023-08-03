package analytics

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/borealisdb/commons/credentials"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/sirupsen/logrus"
	"math"
	"postgres-explain/proto"
	"strconv"
)

type Service struct {
	credentialsProvider credentials.Credentials
	log                 *logrus.Entry
	repo                Repository
	proto.QueryAnalyticsServer
}

func (s Service) GetQueriesList(ctx context.Context, request *proto.GetQueriesListRequest) (*proto.GetQueriesListResponse, error) {
	metricsSet, err := s.repo.GetMetrics(ctx, QueriesMetricsRequest{
		PeriodStartFrom: request.PeriodStartFrom.AsTime(),
		PeriodStartTo:   request.PeriodStartTo.AsTime(),
		ClusterName:     request.ClusterName,
		Limit:           int(request.Limit),
		Order:           request.Order,
	})
	if err != nil {
		return nil, err
	}

	queries := make([]*proto.Query, 0)
	for _, metrics := range metricsSet {
		queryText := metrics["query"].(string)
		fingerprint, err := pg_query.Fingerprint(queryText)
		if err != nil {
			return nil, fmt.Errorf("could not calculate query Fingerprint %v", err)
		}

		mts := make([]*proto.MetricValues, 0)
		for _, mapping := range MetricsMappingsSimple {
			value := metrics[mapping.Key]
			mts = append(mts, &proto.MetricValues{Value: convertMetricValueToFloat32(value)})
		}

		queries = append(queries, &proto.Query{
			Id:          strconv.FormatInt(metrics["queryid"].(int64), 10),
			Fingerprint: fingerprint,
			Text:        queryText,
			Parameters:  nil,
			PlanIds:     nil,
			Metrics:     mts,
		})
	}

	return &proto.GetQueriesListResponse{Queries: queries, Mappings: MetricsMappingsSimple}, nil
}

func convertMetricValueToFloat32(val interface{}) float32 {
	switch val.(type) {
	case int64:
		return float32(val.(int64))
	case float64:
		return float32(val.(float64))
	case []uint8:
		endian := binary.LittleEndian.Uint32(val.([]uint8))
		return math.Float32frombits(endian)
	default:
		return float32(val.(int64))
	}
}
