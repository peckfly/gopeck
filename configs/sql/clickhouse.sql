CREATE DATABASE gopeck;

-- create
drop table gopeck.stress_log;
CREATE TABLE gopeck.stress_log
(
    plan_id                       UInt64,
    task_id                       UInt64,
    url                           String,
    timestamp                     Int64,
    total_num                     Int64,
    total_response_content_length Int64,
    duration_map Map(Int32, Int64),
    status_map Map(Int32, Int64),
    error_map Map(String, Int64),
    body_check_result_map Map(String, Int64),
    latency_map Map(String, Int32)
) ENGINE = MergeTree
      ORDER BY (timestamp)
      PARTITION BY toYYYYMMDD(toDateTime(timestamp))
      TTL toDateTime(timestamp) + INTERVAL 20 DAY;

-- select
select * FROM gopeck.stress_log ;

