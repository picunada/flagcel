-- +goose Up
CREATE TABLE IF NOT EXISTS flag_usage_buckets (
    environment_id  UUID        NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    flag_key        TEXT        NOT NULL,
    bucket_start    TIMESTAMPTZ NOT NULL,
    value_type      TEXT        NOT NULL,
    value_key       TEXT        NOT NULL,
    value           JSONB       NOT NULL,
    reason          TEXT        NOT NULL,
    matched_rule_id TEXT        NOT NULL DEFAULT '',
    api_key_id      TEXT        NOT NULL DEFAULT '',
    source          TEXT        NOT NULL DEFAULT '',
    count           BIGINT      NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (
        environment_id,
        flag_key,
        bucket_start,
        value_key,
        reason,
        matched_rule_id,
        api_key_id,
        source
    ),
    FOREIGN KEY (environment_id, flag_key)
        REFERENCES flags(environment_id, key)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_flag_usage_buckets_flag_time
    ON flag_usage_buckets (environment_id, flag_key, bucket_start DESC);

CREATE TABLE IF NOT EXISTS flag_evaluation_events (
    id              UUID        PRIMARY KEY,
    environment_id  UUID        NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    flag_key        TEXT        NOT NULL,
    observed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    value_type      TEXT        NOT NULL,
    value           JSONB       NOT NULL,
    reason          TEXT        NOT NULL,
    matched_rule_id TEXT        NOT NULL DEFAULT '',
    api_key_id      TEXT        NOT NULL DEFAULT '',
    source          TEXT        NOT NULL DEFAULT '',
    latency_ms      DOUBLE PRECISION NOT NULL DEFAULT 0,
    context         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    FOREIGN KEY (environment_id, flag_key)
        REFERENCES flags(environment_id, key)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_flag_evaluation_events_flag_time
    ON flag_evaluation_events (environment_id, flag_key, observed_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_flag_evaluation_events_flag_time;
DROP TABLE IF EXISTS flag_evaluation_events;
DROP INDEX IF EXISTS idx_flag_usage_buckets_flag_time;
DROP TABLE IF EXISTS flag_usage_buckets;
