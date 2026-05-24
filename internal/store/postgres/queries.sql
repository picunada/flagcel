-- name: GetFlag :one
SELECT key, enabled, default_value, context_id
FROM flags
WHERE key = $1;

-- name: ListFlags :many
SELECT key, enabled, default_value, context_id
FROM flags
ORDER BY key;

-- name: ListRulesForFlag :many
SELECT * FROM rules
WHERE flag_key = $1
ORDER BY position;

-- name: ListAllRules :many
SELECT * FROM rules
ORDER BY flag_key, position;

-- name: UpsertFlag :exec
INSERT INTO flags (key, enabled, default_value, context_id, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (key) DO UPDATE SET
    enabled       = EXCLUDED.enabled,
    default_value = EXCLUDED.default_value,
    context_id    = EXCLUDED.context_id,
    updated_at    = NOW();

-- name: DeleteRulesForFlag :exec
DELETE FROM rules WHERE flag_key = $1;

-- name: InsertRule :exec
INSERT INTO rules (id, flag_key, expression, rollout_percentage, rollout_bucket_by, position)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DeleteFlag :exec
DELETE FROM flags WHERE key = $1;

-- name: GetRule :one
SELECT * FROM rules
WHERE flag_key = $1 AND id = $2;

-- name: InsertRuleAtEnd :exec
INSERT INTO rules (id, flag_key, expression, rollout_percentage, rollout_bucket_by, position)
VALUES (
    $1, $2, $3, $4, $5,
    COALESCE((SELECT MAX(position) + 1 FROM rules WHERE flag_key = $2), 0)
);

-- name: UpdateRule :execrows
UPDATE rules
SET expression         = $3,
    rollout_percentage = $4,
    rollout_bucket_by  = $5
WHERE flag_key = $1 AND id = $2;

-- name: DeleteRule :execrows
DELETE FROM rules
WHERE flag_key = $1 AND id = $2;

-- name: SetRulePosition :execrows
UPDATE rules
SET position = $3
WHERE flag_key = $1 AND id = $2;

-- name: ListContexts :many
SELECT id, name, description, fields
FROM contexts
ORDER BY name;

-- name: GetContext :one
SELECT id, name, description, fields
FROM contexts
WHERE id = $1;

-- name: InsertContext :exec
INSERT INTO contexts (id, name, description, fields)
VALUES ($1, $2, $3, $4);

-- name: UpdateContext :execrows
UPDATE contexts
SET name        = $2,
    description = $3,
    fields      = $4,
    updated_at  = NOW()
WHERE id = $1;

-- name: DeleteContext :execrows
DELETE FROM contexts WHERE id = $1;
