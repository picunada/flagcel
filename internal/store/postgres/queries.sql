-- name: GetFlag :one
SELECT key, value_type, enabled, default_value, context_id, description, created_at, updated_at, created_by, deleted_by
FROM flags
WHERE key = $1;

-- name: ListFlags :many
SELECT key, value_type, enabled, default_value, context_id, description, created_at, updated_at, created_by, deleted_by
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
INSERT INTO flags (key, value_type, enabled, default_value, context_id, description, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (key) DO UPDATE SET
    value_type    = EXCLUDED.value_type,
    enabled       = EXCLUDED.enabled,
    default_value = EXCLUDED.default_value,
    context_id    = EXCLUDED.context_id,
    description   = EXCLUDED.description,
    updated_at    = NOW();

-- name: DeleteRulesForFlag :exec
DELETE FROM rules WHERE flag_key = $1;

-- name: InsertRule :exec
INSERT INTO rules (id, flag_key, expression, rollout_percentage, rollout_bucket_by, position, value, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: DeleteFlag :exec
DELETE FROM flags WHERE key = $1;

-- name: TouchFlag :execrows
UPDATE flags
SET updated_at = NOW()
WHERE key = $1;

-- name: GetRule :one
SELECT * FROM rules
WHERE flag_key = $1 AND id = $2;

-- name: InsertRuleAtEnd :exec
INSERT INTO rules (id, flag_key, expression, rollout_percentage, rollout_bucket_by, position, value, description)
VALUES (
    $1, $2, $3, $4, $5,
    COALESCE((SELECT MAX(position) + 1 FROM rules WHERE flag_key = $2), 0),
    $6, $7
);

-- name: UpdateRule :execrows
UPDATE rules
SET expression         = $3,
    rollout_percentage = $4,
    rollout_bucket_by  = $5,
    value              = $6,
    description        = $7,
    updated_at         = NOW()
WHERE flag_key = $1 AND id = $2;

-- name: DeleteRule :execrows
DELETE FROM rules
WHERE flag_key = $1 AND id = $2;

-- name: SetRulePosition :execrows
UPDATE rules
SET position = $3,
    updated_at = NOW()
WHERE flag_key = $1 AND id = $2;

-- name: ListContexts :many
SELECT id, name, description, fields, created_at, updated_at, created_by, deleted_by
FROM contexts
ORDER BY name;

-- name: GetContext :one
SELECT id, name, description, fields, created_at, updated_at, created_by, deleted_by
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

-- name: UpsertUserByOIDC :one
INSERT INTO users (id, oidc_subject, email, name, password_hash, admin, updated_at)
VALUES ($1, $2, $3, $4, '', $5, NOW())
ON CONFLICT (oidc_subject) DO UPDATE SET
    email      = EXCLUDED.email,
    name       = EXCLUDED.name,
    admin      = EXCLUDED.admin,
    updated_at = NOW()
RETURNING id, oidc_subject, email, name, description, admin, created_at, updated_at, created_by, deleted_by;

-- name: UpsertLocalAdmin :one
INSERT INTO users (id, oidc_subject, email, name, password_hash, admin, updated_at)
VALUES ($1, $2, $3, $4, $5, TRUE, NOW())
ON CONFLICT (oidc_subject) DO UPDATE SET
    email         = EXCLUDED.email,
    name          = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    admin         = TRUE,
    updated_at    = NOW()
RETURNING id, oidc_subject, email, name, description, admin, created_at, updated_at, created_by, deleted_by;

-- name: GetUserByEmail :one
SELECT id, oidc_subject, email, name, password_hash, description, admin, created_at, updated_at, created_by, deleted_by
FROM users
WHERE lower(email) = lower($1);

-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetUserBySessionHash :one
SELECT u.id, u.oidc_subject, u.email, u.name, u.description, u.admin, u.created_at, u.updated_at, u.created_by, u.deleted_by
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = $1
  AND s.expires_at > NOW();

-- name: DeleteSessionByHash :exec
DELETE FROM sessions WHERE token_hash = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= NOW();

-- name: CreateAPIKey :one
INSERT INTO api_keys (id, name, description, prefix, secret_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, prefix, created_at, updated_at, last_used_at, revoked_at, created_by, deleted_by;

-- name: ListAPIKeys :many
SELECT id, name, description, prefix, created_at, updated_at, last_used_at, revoked_at, created_by, deleted_by
FROM api_keys
ORDER BY created_at DESC;

-- name: GetActiveAPIKeyByHash :one
SELECT id, name, description, prefix, created_at, updated_at, last_used_at, revoked_at, created_by, deleted_by
FROM api_keys
WHERE secret_hash = $1
  AND revoked_at IS NULL;

-- name: RevokeAPIKey :execrows
UPDATE api_keys
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;

-- name: TouchAPIKey :exec
UPDATE api_keys
SET last_used_at = NOW(),
    updated_at = NOW()
WHERE id = $1;
