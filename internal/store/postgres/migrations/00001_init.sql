-- +goose Up
CREATE TABLE IF NOT EXISTS contexts (
    id          UUID        PRIMARY KEY,
    name        TEXT        NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    fields      JSONB       NOT NULL DEFAULT '[]'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flags (
    key           TEXT        PRIMARY KEY,
    enabled       BOOLEAN     NOT NULL DEFAULT FALSE,
    default_value BOOLEAN     NOT NULL DEFAULT FALSE,
    context_id    UUID        REFERENCES contexts(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rules (
    id                 TEXT    NOT NULL,
    flag_key           TEXT    NOT NULL REFERENCES flags(key) ON DELETE CASCADE,
    expression         TEXT    NOT NULL,
    rollout_percentage INTEGER NOT NULL DEFAULT 100,
    rollout_bucket_by  TEXT    NOT NULL DEFAULT '',
    position           INTEGER NOT NULL,
    PRIMARY KEY (flag_key, id)
);

CREATE INDEX IF NOT EXISTS idx_rules_flag_key_position
    ON rules(flag_key, position);

-- +goose Down
DROP INDEX IF EXISTS idx_rules_flag_key_position;
DROP TABLE IF EXISTS rules;
DROP TABLE IF EXISTS flags;
DROP TABLE IF EXISTS contexts;
