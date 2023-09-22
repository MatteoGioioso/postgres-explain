package analytics

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"postgres-explain/proto"
)

const (
	MetricsTableName = "metrics"
	requestsCap      = 100
	batchTimeout     = 500 * time.Millisecond
	batchErrorDelay  = time.Second
)

const insertSQL = `
  INSERT INTO metrics
  (
    queryid,
    cluster_name,
    instance_name,
    database,
    schema,
    tables,
    username,
    client_host,
    replication_set,
    environment,
    labels.key,
    labels.value,
    agent_id,
    period_start,
    period_length,
    fingerprint,
    is_truncated,
    num_queries_with_warnings,
    warnings.code,
    warnings.count,
    num_queries_with_errors,
    errors.code,
    errors.count,
    num_queries,
    m_query_time_cnt,
    m_query_time_sum,
    m_query_time_min,
    m_query_time_max,
    m_query_time_p99,
    m_rows_sent_cnt,
    m_rows_sent_sum,
    m_rows_sent_min,
    m_rows_sent_max,
    m_rows_sent_p99,
    m_shared_blks_hit_cnt,
    m_shared_blks_hit_sum,
    m_shared_blks_read_cnt,
    m_shared_blks_read_sum,
    m_shared_blks_dirtied_cnt,
    m_shared_blks_dirtied_sum,
    m_shared_blks_written_cnt,
    m_shared_blks_written_sum,
    m_local_blks_hit_cnt,
    m_local_blks_hit_sum,
    m_local_blks_read_cnt,
    m_local_blks_read_sum,
    m_local_blks_dirtied_cnt,
    m_local_blks_dirtied_sum,
    m_local_blks_written_cnt,
    m_local_blks_written_sum,
    m_temp_blks_read_cnt,
    m_temp_blks_read_sum,
    m_temp_blks_written_cnt,
    m_temp_blks_written_sum,
    m_blk_read_time_cnt,
    m_blk_read_time_sum,
    m_blk_write_time_cnt,
    m_blk_write_time_sum
   )
  VALUES (
    :queryid,
    :cluster_name,
	:instance_name,
    :database,
    :schema,
    :tables,
    :username,
    :client_host,
    :replication_set,
    :environment,
    :labels_key,
    :labels_value,
    :agent_id,
    :period_start_ts,
    :period_length_secs,
    :fingerprint,
    :is_truncated,
    :num_queries_with_warnings,
    :warnings_code,
    :warnings_count,
    :num_queries_with_errors,
    :errors_code,
    :errors_count,
    :num_queries,
    :m_query_time_cnt,
    :m_query_time_sum,
    :m_query_time_min,
    :m_query_time_max,
    :m_query_time_p99,
    :m_rows_sent_cnt,
    :m_rows_sent_sum,
    :m_rows_sent_min,
    :m_rows_sent_max,
    :m_rows_sent_p99,
    :m_shared_blks_hit_cnt,
    :m_shared_blks_hit_sum,
    :m_shared_blks_read_cnt,
    :m_shared_blks_read_sum,
    :m_shared_blks_dirtied_cnt,
    :m_shared_blks_dirtied_sum,
    :m_shared_blks_written_cnt,
    :m_shared_blks_written_sum,
    :m_local_blks_hit_cnt,
    :m_local_blks_hit_sum,
    :m_local_blks_read_cnt,
    :m_local_blks_read_sum,
    :m_local_blks_dirtied_cnt,
    :m_local_blks_dirtied_sum,
    :m_local_blks_written_cnt,
    :m_local_blks_written_sum,
    :m_temp_blks_read_cnt,
    :m_temp_blks_read_sum,
    :m_temp_blks_written_cnt,
    :m_temp_blks_written_sum,
    :m_blk_read_time_cnt,
    :m_blk_read_time_sum,
    :m_blk_write_time_cnt,
    :m_blk_write_time_sum
  )
`

// MetricsBucketExtended extends proto MetricsBucket to store converted data into db.
type MetricsBucketExtended struct {
	PeriodStart      time.Time `json:"period_start_ts"`
	LabelsKey        []string  `json:"labels_key"`
	LabelsValues     []string  `json:"labels_value"`
	WarningsCode     []uint64  `json:"warnings_code"`
	WarningsCount    []uint64  `json:"warnings_count"`
	ErrorsCode       []uint64  `json:"errors_code"`
	ErrorsCount      []uint64  `json:"errors_count"`
	IsQueryTruncated uint8     `json:"is_query_truncated"` // uint32 -> uint8
	*proto.MetricsBucket
}

// MetricsBucket implements models to store metrics bucket
type MetricsBucket struct {
	db         *sqlx.DB
	l          *logrus.Entry
	requestsCh chan *proto.StatementsCollectRequest
}

func NewMetricsBucket(db *sqlx.DB, log *logrus.Entry) *MetricsBucket {
	requestsCh := make(chan *proto.StatementsCollectRequest, requestsCap)

	mb := &MetricsBucket{
		db:         db,
		l:          log,
		requestsCh: requestsCh,
	}

	return mb
}

// Run stores incoming data until context is canceled.
// It exits when all data is stored.
func (mb *MetricsBucket) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		mb.l.Warn("Closing requests channel.")
		close(mb.requestsCh)
	}()

	for ctx.Err() == nil {
		if err := mb.insertBatch(batchTimeout); err != nil {
			mb.l.Errorf("could not insert batch: %v", err)
			time.Sleep(batchErrorDelay)
		}
	}

	// insert one last final batch
	_ = mb.insertBatch(0)
}

func (mb *MetricsBucket) insertBatch(timeout time.Duration) (err error) {
	// wait for first request before doing anything, ignore timeout
	req, ok := <-mb.requestsCh
	if !ok {
		mb.l.Warn("Requests channel closed, nothing to store.")
		return
	}

	var buckets int
	start := time.Now()
	defer func() {
		d := time.Since(start)

		if err == nil {
			mb.l.Infof("Saved %d buckets in %s.", buckets, d)
		} else {
			mb.l.Errorf("Failed to save %d buckets in %s: %s.", buckets, d, err)
		}
	}()

	// begin "transaction" and commit or rollback it on exit
	var tx *sqlx.Tx
	if tx, err = mb.db.Beginx(); err != nil {
		err = errors.Wrap(err, "failed to begin transaction")
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "failed to commit transaction")
		}
	}()

	// prepare INSERT statement and close it on exit
	var stmt *sqlx.NamedStmt
	if stmt, err = tx.PrepareNamed(insertSQL); err != nil {
		err = errors.Wrap(err, "failed to prepare statement")
		return
	}
	defer func() {
		if e := stmt.Close(); e != nil && err == nil {
			err = errors.Wrap(e, "failed to close statement")
		}
	}()

	// limit only by time, not by batch size, because large batches already handled by the driver
	// ("block_size" query parameter)
	var timeoutCh <-chan time.Time
	if timeout > 0 {
		t := time.NewTimer(timeout)
		defer t.Stop()
		timeoutCh = t.C
	}

	for {
		// INSERT buckets from current request
		for _, metricsBucket := range req.MetricsBucket {
			buckets++

			lk, lv := mapToArrsStrStr(metricsBucket.Labels)
			wk, wv := mapToArrsIntInt(metricsBucket.Warnings)
			ek, ev := mapToArrsIntInt(metricsBucket.Errors)

			var truncated uint8
			if metricsBucket.IsTruncated {
				truncated = 1
			}

			q := MetricsBucketExtended{
				time.Unix(int64(metricsBucket.GetPeriodStartUnixSecs()), 0).UTC(),
				lk,
				lv,
				wk,
				wv,
				ek,
				ev,
				truncated,
				metricsBucket,
			}

			if _, err = stmt.Exec(q); err != nil {
				err = errors.Wrap(err, "failed to exec")
				return
			}
		}

		// wait for the next request or exit on timer
		select {
		case req, ok = <-mb.requestsCh:
			if !ok {
				mb.l.Warn("Requests channel closed, exiting.")
				return
			}
		case <-timeoutCh:
			return
		}
	}
}

// Save store metrics bucket received from agent into db.
func (mb *MetricsBucket) Save(agentMsg *proto.StatementsCollectRequest) error {
	if len(agentMsg.MetricsBucket) == 0 {
		mb.l.Warnf("Nothing to save - no metrics buckets.")
		return nil
	}

	mb.requestsCh <- agentMsg
	return nil
}

// mapToArrsStrStr converts map into two lists.
func mapToArrsStrStr(m map[string]string) (keys []string, values []string) {
	keys = make([]string, 0, len(m))
	values = make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}

	return
}

// mapToArrsIntInt converts map into two lists.
func mapToArrsIntInt(m map[uint64]uint64) (keys []uint64, values []uint64) {
	keys = make([]uint64, 0, len(m))
	values = make([]uint64, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}

	return
}
