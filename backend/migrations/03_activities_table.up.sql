CREATE TABLE activities
(
    `query_id` LowCardinality(String),
    `fingerprint`        String,
    `query_sha`          String,
    `datname`            String,
    `pid`                UInt32,
    `usesysid`           String,
    `usename`            String,
    `application_name`   String,
    `backend_type`       String,
    `client_hostname` LowCardinality(String),
    `wait_event_type`    String,
    `wait_event`         String,
    `parsed_query`       String,
    `query`              String,
    `is_query_truncated` UInt8 COMMENT 'Indicates if query is too long and was truncated from postgres',
    `is_not_explainable` UInt8 COMMENT 'Indicates if the query cannot be run with EXPLAIN',
    `state`              String,
    `query_start`        UInt32,
    `duration`           Float32,
    `period_start`       DateTime COMMENT 'Time when collection of the sample started',
    `period_length`      UInt32 COMMENT 'Duration of collected sample (usually 1 second)',
    `current_timestamp`  DateTime COMMENT 'Time when the sample was taken in the db',
    `cluster_name`       String,
    `instance_name`      String,
    `cpu_cores`          Float32
) ENGINE = MergeTree PARTITION BY toYYYYMMDD(period_start)
      ORDER BY
          (
           query_id,
           datname,
           usename,
           client_hostname,
           period_start,
           current_timestamp
              ) SETTINGS index_granularity = 8192;