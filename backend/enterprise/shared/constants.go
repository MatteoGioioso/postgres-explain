package shared

import "time"

const QueryTimeout = 30 * time.Second
const cannotPrepare = "cannot prepare query"
const cannotPopulate = "cannot populate query arguments"
const CannotExecute = "cannot execute query"

var commonColumnNames = map[string]struct{}{
	"query_time":   {},
	"rows_sent":    {},
	"query_length": {},
}

var sumColumnNames = map[string]struct{}{
	"shared_blks_hit":     {},
	"shared_blks_read":    {},
	"shared_blks_dirtied": {},
	"shared_blks_written": {},
	"local_blks_hit":      {},
	"local_blks_read":     {},
	"local_blks_dirtied":  {},
	"local_blks_written":  {},
	"temp_blks_read":      {},
	"temp_blks_written":   {},
	"blk_read_time":       {},
	"blk_write_time":      {},
	"plans_calls":         {},
	"wal_records":         {},
	"wal_fpi":             {},
	"plan_time":           {},
	"wal_bytes":           {},
}
