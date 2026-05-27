-- +goose Up
ALTER TABLE flags
    ADD COLUMN value_type TEXT NOT NULL DEFAULT 'boolean';

ALTER TABLE flags
    ALTER COLUMN default_value DROP DEFAULT,
    ALTER COLUMN default_value TYPE JSONB USING to_jsonb(default_value),
    ALTER COLUMN default_value SET DEFAULT 'false'::jsonb;

ALTER TABLE rules
    ADD COLUMN value JSONB NOT NULL DEFAULT 'true'::jsonb;

-- +goose Down
ALTER TABLE rules
    DROP COLUMN value;

ALTER TABLE flags
    ALTER COLUMN default_value DROP DEFAULT,
    ALTER COLUMN default_value TYPE BOOLEAN USING CASE
        WHEN jsonb_typeof(default_value) = 'boolean' THEN (default_value #>> '{}')::boolean
        ELSE FALSE
    END,
    ALTER COLUMN default_value SET DEFAULT FALSE;

ALTER TABLE flags
    DROP COLUMN value_type;
