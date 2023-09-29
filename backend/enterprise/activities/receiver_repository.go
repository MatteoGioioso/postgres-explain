package activities

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"postgres-explain/proto"
)

const ActivitiesTableName = "activities"

const insertActivitySQL = `
  INSERT INTO activities
  (
    current_timestamp,
    period_start,
    period_length,
	fingerprint,
	query_id,
    datname,
    pid,
	usesysid,
    usename,
    application_name,
    backend_type,
    client_hostname,
    wait_event_type,
    wait_event,
    parsed_query,
    query,
    state,
    query_start,
    duration,
	cluster_name,
    instance_name,
    cpu_cores,
	is_query_truncated,
	query_sha,
	is_not_explainable
   )
  VALUES (
    :current_timestamp,
	:period_start,
    :period_length,
  	:fingerprint,
   	:query_id,
    :datname,
    :pid,
    :usesysid,
    :usename,
	:application_name,
    :backend_type,
    :client_hostname,
    :wait_event_type,
    :wait_event,
    :parsed_query,
    :query,
    :state,
	:query_start,
    :duration
	:cluster_name,
    :instance_name,
    :cpu_cores,
	:is_query_truncated
	:query_sha,
	:is_not_explainable
  )
`

type ActivitySampler struct {
	db *sqlx.DB
	l  *logrus.Entry
}

func NewActivitySampler(db *sqlx.DB, log *logrus.Entry) *ActivitySampler {
	return &ActivitySampler{db: db, l: log.WithField("subcomponent", "activity-collector")}
}

// Save store metrics bucket received from agent into db.
func (as *ActivitySampler) Save(agentMsg *proto.ActivityCollectRequest) error {
	if len(agentMsg.ActivitySamples) == 0 {
		as.l.Warnf("Nothing to save - no metrics buckets.")
		return nil
	}

	if err := as.insertBatch(agentMsg.ActivitySamples); err != nil {
		return err
	}

	return nil
}

func (as *ActivitySampler) insertBatch(samples []*proto.ActivitySample) (err error) {
	// begin "transaction" and commit or rollback it on exit
	var tx *sqlx.Tx
	if tx, err = as.db.Beginx(); err != nil {
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
	if stmt, err = tx.PrepareNamed(insertActivitySQL); err != nil {
		err = errors.Wrap(err, "failed to prepare statement")
		return
	}
	defer func() {
		if e := stmt.Close(); e != nil && err == nil {
			err = errors.Wrap(e, "failed to close statement")
		}
	}()

	savedSamplesCounter := 0
	for _, sample := range samples {
		q := ActivitySampleDB{}
		q.FromActivitySample(sample)

		if _, err = stmt.Exec(q); err != nil {
			err = errors.Wrap(err, "failed to exec")
			return
		}
		savedSamplesCounter++
	}

	as.l.Infof("Saved %v samples", savedSamplesCounter)
	return nil
}
