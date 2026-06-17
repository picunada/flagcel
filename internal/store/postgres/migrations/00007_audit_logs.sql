-- +goose Up
CREATE TABLE IF NOT EXISTS audit_logs (
    id             UUID        PRIMARY KEY,
    environment_id UUID        NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    resource_type  TEXT        NOT NULL,
    resource_id    TEXT        NOT NULL,
    action         TEXT        NOT NULL,
    version        INT         NOT NULL,
    snapshot       JSONB,
    actor_id       UUID        REFERENCES users(id) ON DELETE SET NULL,
    actor_label    TEXT,
    summary        TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_resource
    ON audit_logs (environment_id, resource_type, resource_id, version DESC);

ALTER TABLE flags
    ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE rules
    ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE rules
    DROP COLUMN IF EXISTS updated_by;

ALTER TABLE flags
    DROP COLUMN IF EXISTS updated_by;

DROP INDEX IF EXISTS idx_audit_logs_resource;

DROP TABLE IF EXISTS audit_logs;
