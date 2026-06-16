-- +goose Up
CREATE TABLE IF NOT EXISTS environments (
    id          UUID        PRIMARY KEY,
    key         TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID        REFERENCES users(id) ON DELETE SET NULL,
    deleted_by  UUID        REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO environments (id, key, name, description)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'production',
    'Production',
    'Default environment created during the environments migration.'
)
ON CONFLICT (key) DO NOTHING;

ALTER TABLE flags
    ADD COLUMN IF NOT EXISTS environment_id UUID;

UPDATE flags
SET environment_id = '00000000-0000-0000-0000-000000000001'
WHERE environment_id IS NULL;

ALTER TABLE flags
    ALTER COLUMN environment_id SET NOT NULL;

ALTER TABLE rules
    ADD COLUMN IF NOT EXISTS environment_id UUID;

UPDATE rules
SET environment_id = '00000000-0000-0000-0000-000000000001'
WHERE environment_id IS NULL;

ALTER TABLE rules
    ALTER COLUMN environment_id SET NOT NULL;

ALTER TABLE rules
    DROP CONSTRAINT IF EXISTS rules_flag_key_fkey;

ALTER TABLE flags
    DROP CONSTRAINT IF EXISTS flags_pkey;

ALTER TABLE flags
    ADD CONSTRAINT flags_pkey PRIMARY KEY (environment_id, key),
    ADD CONSTRAINT flags_environment_id_fkey
        FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE RESTRICT;

ALTER TABLE rules
    DROP CONSTRAINT IF EXISTS rules_pkey;

ALTER TABLE rules
    ADD CONSTRAINT rules_pkey PRIMARY KEY (environment_id, flag_key, id),
    ADD CONSTRAINT rules_environment_flag_fkey
        FOREIGN KEY (environment_id, flag_key)
        REFERENCES flags(environment_id, key)
        ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_rules_flag_key_position;
CREATE INDEX IF NOT EXISTS idx_rules_environment_flag_position
    ON rules(environment_id, flag_key, position);

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS environment_id UUID;

UPDATE api_keys
SET environment_id = '00000000-0000-0000-0000-000000000001'
WHERE environment_id IS NULL;

ALTER TABLE api_keys
    ALTER COLUMN environment_id SET NOT NULL,
    ADD CONSTRAINT api_keys_environment_id_fkey
        FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_api_keys_environment_id
    ON api_keys(environment_id);

-- +goose Down
DROP INDEX IF EXISTS idx_api_keys_environment_id;

ALTER TABLE api_keys
    DROP CONSTRAINT IF EXISTS api_keys_environment_id_fkey,
    DROP COLUMN IF EXISTS environment_id;

ALTER TABLE rules
    DROP CONSTRAINT IF EXISTS rules_environment_flag_fkey,
    DROP CONSTRAINT IF EXISTS rules_pkey;

ALTER TABLE flags
    DROP CONSTRAINT IF EXISTS flags_environment_id_fkey,
    DROP CONSTRAINT IF EXISTS flags_pkey;

ALTER TABLE flags
    ADD CONSTRAINT flags_pkey PRIMARY KEY (key);

ALTER TABLE rules
    ADD CONSTRAINT rules_pkey PRIMARY KEY (flag_key, id),
    ADD CONSTRAINT rules_flag_key_fkey
        FOREIGN KEY (flag_key) REFERENCES flags(key) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_rules_environment_flag_position;
CREATE INDEX IF NOT EXISTS idx_rules_flag_key_position
    ON rules(flag_key, position);

ALTER TABLE rules
    DROP COLUMN IF EXISTS environment_id;

ALTER TABLE flags
    DROP COLUMN IF EXISTS environment_id;

DROP TABLE IF EXISTS environments;
