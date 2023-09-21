CREATE TABLE plans
(
    `id` String COMMENT 'unique plan id',
    `alias` String COMMENT 'custom alias for the plan',
    `query_fingerprint` LowCardinality(String) COMMENT 'query fingerprint',
    `queryid` LowCardinality(String) COMMENT 'hash of query fingerprint from postgres',
    `plan` String COMMENT 'JSON string of the query plan',
    `original_plan` String COMMENT 'JSON string of the original plan',
    `query` String COMMENT 'query',
    `database` LowCardinality(String) COMMENT 'PostgreSQL: database',
    `schema` LowCardinality(String) COMMENT 'PostgreSQL: schema',
    `username` LowCardinality(String) COMMENT 'client user name',
    `cluster` LowCardinality(String) COMMENT 'Cluster name',
    `period_start`  DateTime COMMENT 'Time when collection of bucket started',
    `period_length` UInt32 COMMENT 'Duration of collection bucket',
    `optimization_id` String COMMENT 'for tracking optimizations'
) ENGINE = MergeTree PARTITION BY toYYYYMMDD(period_start)
      ORDER BY
          (
              period_start
              ) SETTINGS index_granularity = 8192;
