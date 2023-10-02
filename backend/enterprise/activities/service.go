package activities

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"postgres-explain/backend/enterprise/shared"
	sharedBackend "postgres-explain/backend/shared"
	"postgres-explain/proto"
	"time"
)

type Service struct {
	Repo          Repository
	MetricsRepo   shared.MetricsRepository
	WaitEventsMap map[string]WaitEvent
	log           *logrus.Entry

	proto.ActivitiesServer
}

func NewService(
	repo Repository,
	metricsRepo shared.MetricsRepository,
	waitEventsMap map[string]WaitEvent,
	log *logrus.Entry,
) *Service {
	return &Service{
		Repo:          repo,
		WaitEventsMap: waitEventsMap,
		MetricsRepo:   metricsRepo,
		log:           log.WithField("subcomponent", "activity_profiler"),
	}
}

func (aps *Service) GetProfile(ctx context.Context, in *proto.GetProfileRequest) (*proto.GetProfileResponse, error) {
	if err := shared.ValidateCommonRequestProps(shared.Validate{
		PeriodStartFrom: in.PeriodStartFrom,
		PeriodStartTo:   in.PeriodStartTo,
		ClusterName:     in.ClusterName,
	}); err != nil {
		return nil, err
	}

	results, err := aps.Repo.Select(ctx, QueryArgs{
		PeriodStartFromSec: in.PeriodStartFrom.Seconds,
		PeriodStartToSec:   in.PeriodStartTo.Seconds,
		ClusterName:        in.ClusterName,
	})
	if err != nil {
		aps.log.Errorf("error querying clickhouse: %v", err)
		return &proto.GetProfileResponse{}, fmt.Errorf("something went wrong")
	}

	if len(results) == 0 {
		return &proto.GetProfileResponse{}, nil
	}

	// TODO document this and maybe optimize
	// Double transformation, doing in one was too complex,
	// thus we transform to slot data structure to make it more convenient.
	// This could be optimized later
	slots, ascOrderedUniqueTimestamps := aps.getSlots(results)
	traces := aps.mapSlotsToTraces(slots, ascOrderedUniqueTimestamps)
	//TODO find a better method to set cpu_cores
	var cpuCores float32 = 0
	if len(results) > 0 {
		cpuCores = results[0].CpuCores
	}
	return &proto.GetProfileResponse{Traces: traces, CurrentCpuCores: cpuCores}, nil
}

func (aps *Service) GetTopQueries(ctx context.Context, in *proto.GetTopQueriesRequest) (*proto.GetTopQueriesResponse, error) {
	if err := shared.ValidateCommonRequestProps(shared.Validate{
		PeriodStartFrom: in.PeriodStartFrom,
		PeriodStartTo:   in.PeriodStartTo,
		ClusterName:     in.ClusterName,
	}); err != nil {
		return nil, err
	}

	args := QueryArgs{
		PeriodStartFromSec: in.PeriodStartFrom.Seconds,
		PeriodStartToSec:   in.PeriodStartTo.Seconds,
		ClusterName:        in.ClusterName,
	}
	queries, err := aps.Repo.GetQueriesByWaitEventCount(ctx, args)
	if err != nil {
		aps.log.Errorf("error querying clickhouse: %v", err)
		return &proto.GetTopQueriesResponse{}, fmt.Errorf("something went wrong")
	}

	if len(queries) == 0 {
		return &proto.GetTopQueriesResponse{}, nil
	}

	queriesMetrics, _, err := aps.getMetricsForTopQueries(ctx, args, queries)
	if err != nil {
		return nil, err
	}

	xValueGetter := func(query QueryDB) string {
		return query.Fingerprint
	}
	metadata := aps.getQueriesMetadata(ctx, queries, xValueGetter)
	traces := aps.mapQueriesToTraces(queries, xValueGetter)

	return &proto.GetTopQueriesResponse{
		Traces:          traces,
		QueriesMetrics:  queriesMetrics,
		QueriesMetadata: metadata,
	}, nil
}

func (aps *Service) GetTopQueriesByFingerprint(ctx context.Context, in *proto.GetTopQueriesRequest) (*proto.GetTopQueriesByFingerprintResponse, error) {
	if err := shared.ValidateCommonRequestProps(shared.Validate{
		PeriodStartFrom: in.PeriodStartFrom,
		PeriodStartTo:   in.PeriodStartTo,
		ClusterName:     in.ClusterName,
	}); err != nil {
		return nil, err
	}

	args := QueryArgs{
		PeriodStartFromSec: in.PeriodStartFrom.Seconds,
		PeriodStartToSec:   in.PeriodStartTo.Seconds,
		ClusterName:        in.ClusterName,
		Fingerprint:        in.Fingerprint,
	}
	queries, err := aps.Repo.GetTopQueriesByFingerprint(ctx, args)
	if err != nil {
		return &proto.GetTopQueriesByFingerprintResponse{}, fmt.Errorf("could not GetTopQueriesByFingerprint: %v", err)
	}

	if len(queries) == 0 {
		return &proto.GetTopQueriesByFingerprintResponse{}, nil
	}

	queriesMetrics, _, err := aps.getMetricsForTopQueries(ctx, args, queries)
	if err != nil {
		return nil, fmt.Errorf("could not getMetricsForTopQueries: %v", err)
	}

	xValueGetter := func(query QueryDB) string {
		return query.QuerySha
	}
	metadata := aps.getQueriesMetadata(ctx, queries, xValueGetter)
	traces := aps.mapQueriesToTraces(queries, xValueGetter)

	return &proto.GetTopQueriesByFingerprintResponse{
		Traces:          traces,
		QueriesMetadata: metadata,
		QueryMetrics:    queriesMetrics[queries[0].Fingerprint],
	}, nil
}

func (aps *Service) GetQueryDetails(ctx context.Context, in *proto.GetQueryDetailsRequest) (*proto.GetQueryDetailsResponse, error) {
	if in.PeriodStartFrom == nil || in.PeriodStartTo == nil {
		return nil, fmt.Errorf("from-date: %s or to-date: %s cannot be empty", in.PeriodStartFrom, in.PeriodStartTo)
	}

	periodStartFromSec := in.PeriodStartFrom.Seconds
	periodStartToSec := in.PeriodStartTo.Seconds
	if periodStartFromSec > periodStartToSec {
		return nil, fmt.Errorf("from-date %s cannot be bigger then to-date %s", in.PeriodStartFrom, in.PeriodStartTo)
	}

	queryMetricsByFingerprint, err := aps.MetricsRepo.SelectQueryMetricsByFingerprint(
		ctx,
		shared.MetricsGetArgs{PeriodStartFromSec: periodStartFromSec, PeriodStartToSec: periodStartToSec},
		in.QueryFingerprint,
		in.ClusterName,
	)
	if err != nil {
		return nil, fmt.Errorf("could not query metrics from clickhouse: %v", err)
	}

	traces := aps.toTrace(queryMetricsByFingerprint)

	return &proto.GetQueryDetailsResponse{Traces: traces}, nil
}

func (aps *Service) getMetricsForTopQueries(ctx context.Context, args QueryArgs, queries []QueryDB) (
	map[string]*proto.QueriesMetrics,
	map[string]*proto.MetricValues,
	error,
) {
	totalsList, err := aps.MetricsRepo.Get(
		ctx,
		shared.MetricsGetArgs{
			PeriodStartFromSec: args.PeriodStartFromSec,
			PeriodStartToSec:   args.PeriodStartToSec,
			Totals:             true, // get Totals
		},
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot get get metrics totals")
	}

	totalLen := len(totalsList)
	if totalLen < 2 { // TODO don't know why this, manual query just return one result like it should be
		return nil, nil, fmt.Errorf("totals not found for filter: %s and group: %s in given time range", "", "queryid")
	}

	// Get totals for given filter
	totals := totalsList[totalLen-1]
	durationSec := args.PeriodStartToSec - args.PeriodStartFromSec

	queriesMetrics := make(map[string]*proto.QueriesMetrics)
	for _, query := range queries {
		metricsList, err := aps.MetricsRepo.Get(ctx, shared.MetricsGetArgs{
			PeriodStartFromSec: args.PeriodStartFromSec,
			PeriodStartToSec:   args.PeriodStartToSec,
			Filter:             query.Fingerprint,
			Group:              "fingerprint",
			Totals:             false,
		})
		if err != nil {
			return nil, nil, err
		}
		if len(metricsList) > 0 {
			metrics := shared.MakeMetrics(metricsList[0], totals, durationSec)
			metrics["cpu_total_load"] = &proto.MetricValues{Sum: query.CPULoadTotal}

			queriesMetrics[query.Fingerprint] = &proto.QueriesMetrics{Metrics: metrics}
		}
	}

	return queriesMetrics, shared.MakeMetrics(totals, totals, durationSec), nil
}

func (aps *Service) getQueriesMetadata(ctx context.Context, queries []QueryDB, keyGetter func(db QueryDB) string) map[string]*proto.QueryMetadata {
	m := make(map[string]*proto.QueryMetadata)
	for _, query := range queries {
		m[keyGetter(query)] = &proto.QueryMetadata{
			Fingerprint:           query.Fingerprint,
			Parameters:            sharedBackend.QueryParameterPlaceholder.FindAllString(query.ParsedQuery.String, -1),
			Text:                  query.GetSQL(),
			IsQueryTruncated:      query.IsQueryTruncated == 1,
			QuerySha:              query.QuerySha,
			IsQueryNotExplainable: query.IsQueryNotExplainable,
		}
	}

	return m
}

// the output from the db is {'fingerprint': 'iend09030...', 'cpu_load_wait_events': {'transactionid':0.00008680555555555556,'tuple':0.00001736111111111111, ...}, ...}
// we want to transform into {'transactionid': {'x_values_string': [...<query>], 'y_values_float': [...]}, ...}
// output map[<wait_event_name>]{ x_value_string: ['SELECT * FROM table...', 'SELECT id FROM customers...'], y_values_float: [0.1, 0.0023, ...]}
func (aps *Service) mapQueriesToTraces(queries []QueryDB, xValueGetter func(db QueryDB) string) map[string]*proto.Trace {
	traces := aps.prefillTraces()
	// This is done to remove all wait events that contains only zero values
	hasWaitEventNonZeroValues := make(map[string]bool)

	for _, query := range queries {
		for waitEventName := range query.CPULoadWaitEvents {
			if _, ok := traces[waitEventName]; !ok {
				aps.log.Warningf("trace does not exist for wait event name: %v", waitEventName)
				continue
			}

			hasWaitEventNonZeroValues[waitEventName] = true
			trace := traces[waitEventName]
			traceYValue := float32(query.CPULoadWaitEvents[waitEventName])
			trace.XValuesString = append(trace.XValuesString, xValueGetter(query))
			trace.YValuesFloat = append(trace.YValuesFloat, traceYValue)

			traces[waitEventName] = trace
		}

		for waitEventName := range aps.WaitEventsMap {
			if _, ok := query.CPULoadWaitEvents[waitEventName]; ok {
				continue
			}

			trace := traces[waitEventName]
			trace.XValuesString = append(trace.XValuesString, xValueGetter(query))
			trace.YValuesFloat = append(trace.YValuesFloat, 0)

			traces[waitEventName] = trace
		}
	}

	for waitEventName, _ := range traces {
		if !hasWaitEventNonZeroValues[waitEventName] {
			delete(traces, waitEventName)
			continue
		}
	}

	return traces
}

// This function output time slots, each slot correspond to 1 minute of aggregated wait events count,
// Since we return a map the order is not guaranteed, thus this method will also return
// an ordered (ASC) unique timestamps array to use it later.
func (aps *Service) getSlots(results []SlotDB) (Slots, []time.Time) {
	ascOrderedUniqueTimestamps := make([]time.Time, 0)
	timestampsMap := make(map[time.Time]bool)
	slots := make(Slots)
	for _, slotDB := range results {
		if !timestampsMap[slotDB.Timestamp] {
			ascOrderedUniqueTimestamps = append(ascOrderedUniqueTimestamps, slotDB.Timestamp)
			timestampsMap[slotDB.Timestamp] = true
		}
		if slot, ok := slots[slotDB.Timestamp]; ok {
			slot[slotDB.WaitEventName] = float32(slotDB.WaitEventCount)
			slots[slotDB.Timestamp] = slot
		} else {
			slot := make(Slot)
			slot[slotDB.WaitEventName] = float32(slotDB.WaitEventCount)
			slots[slotDB.Timestamp] = slot
		}
	}
	return slots, ascOrderedUniqueTimestamps
}

// This method will format data for Plotly:
// https://plotly.com/javascript/reference/index/
// it will use the previously created unique ordered timestamps to maintain the sorting.
// If a wait event is missing in the current timestamp we will assign 0 to its value.
func (aps *Service) mapSlotsToTraces(slots Slots, ascOrderedTimeStamps []time.Time) map[string]*proto.Trace {
	traces := aps.prefillTraces()
	hasWaitEventNonZeroValues := make(map[string]bool)

	for _, timestamp := range ascOrderedTimeStamps {
		slot := slots[timestamp]
		for waitEventName := range slot {
			if _, ok := traces[waitEventName]; !ok {
				aps.log.Println("trace does not exist for wait event name:", waitEventName)
				continue
			}

			hasWaitEventNonZeroValues[waitEventName] = true

			trace := traces[waitEventName]
			trace.XValuesTimestamp = append(trace.XValuesTimestamp, timestamppb.New(timestamp))
			trace.YValuesFloat = append(trace.YValuesFloat, slot.GetWaitEventFraction(waitEventName))
			traces[waitEventName] = trace
		}

		for waitEventName := range aps.WaitEventsMap {
			if _, ok := slot[waitEventName]; ok {
				continue
			}

			trace := traces[waitEventName]
			trace.XValuesTimestamp = append(trace.XValuesTimestamp, timestamppb.New(timestamp))
			trace.YValuesFloat = append(trace.YValuesFloat, 0)
			traces[waitEventName] = trace
		}
	}

	for waitEventName, _ := range traces {
		if !hasWaitEventNonZeroValues[waitEventName] {
			delete(traces, waitEventName)
			continue
		}
	}

	return traces
}

// To avoid empty or missing traces (for example a missing wait event for the aggregated minute), we prefill traces
func (aps *Service) prefillTraces() map[string]*proto.Trace {
	traces := make(map[string]*proto.Trace)
	for waitEventName, val := range aps.WaitEventsMap {
		timestamps := make([]*timestamppb.Timestamp, 0)
		floats := make([]float32, 0)
		strings := make([]string, 0)
		traces[waitEventName] = &proto.Trace{
			XValuesTimestamp: timestamps,
			XValuesFloat:     floats,
			XValuesString:    strings,
			YValuesFloat:     floats,
			Color:            val.Color,
		}
	}
	return traces
}

func (aps *Service) toTrace(metrics []shared.QueryMetricDB) map[string]*proto.Trace {
	const (
		rowsSent           = "rows_sent"
		numQueries         = "num_queries"
		queryTimePerCall   = "query_time_per_call"
		totalBlocksRead    = "total_blks_read"
		totalBlocksWritten = "total_blks_written"
		totalBlocksHit     = "total_blks_hit"
	)

	var allowedMetrics = []string{
		rowsSent,
		numQueries,
		queryTimePerCall,
		totalBlocksRead,
		totalBlocksWritten,
		totalBlocksHit,
	}

	var metricsMap = make(map[string]*proto.Trace)
	for _, metricName := range allowedMetrics {
		metricsMap[metricName] = &proto.Trace{}
	}

	timestamps := make([]*timestamppb.Timestamp, 0)
	for _, m := range metrics {
		timestamps = append(timestamps, timestamppb.New(m.Timestamp))

		metricsMap[numQueries].YValuesFloat = append(metricsMap[numQueries].YValuesFloat, float32(m.NumQueries))
		metricsMap[rowsSent].YValuesFloat = append(metricsMap[rowsSent].YValuesFloat, float32(m.RowSent))
		metricsMap[queryTimePerCall].YValuesFloat = append(metricsMap[queryTimePerCall].YValuesFloat, float32(m.QueryTimeAvgPerCall))
		metricsMap[totalBlocksRead].YValuesFloat = append(metricsMap[totalBlocksRead].YValuesFloat, float32(m.TotalBlocksRead))
		metricsMap[totalBlocksWritten].YValuesFloat = append(metricsMap[totalBlocksWritten].YValuesFloat, float32(m.TotalBlocksWritten))
		metricsMap[totalBlocksHit].YValuesFloat = append(metricsMap[totalBlocksHit].YValuesFloat, float32(m.TotalBlocksHit))
	}

	for _, metricName := range allowedMetrics {
		metricsMap[metricName].XValuesTimestamp = timestamps
	}

	return metricsMap
}
