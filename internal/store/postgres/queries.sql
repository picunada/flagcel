-- name: GetFlag :one
SELECT key, enabled, default_value, context_id, updated_at
FROM flags
WHERE key = $1;

-- name: ListFlags :many
SELECT key, enabled, default_value, context_id, updated_at
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

-- name: TouchFlag :execrows
UPDATE flags
SET updated_at = NOW()
WHERE key = $1;

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

-- name: UpsertUserByOIDC :one
INSERT INTO users (id, oidc_subject, email, name, password_hash, admin, updated_at)
VALUES ($1, $2, $3, $4, '', $5, NOW())
ON CONFLICT (oidc_subject) DO UPDATE SET
    email      = EXCLUDED.email,
    name       = EXCLUDED.name,
    admin      = EXCLUDED.admin,
    updated_at = NOW()
RETURNING id, oidc_subject, email, name, admin;

-- name: UpsertLocalAdmin :one
INSERT INTO users (id, oidc_subject, email, name, password_hash, admin, updated_at)
VALUES ($1, $2, $3, $4, $5, TRUE, NOW())
ON CONFLICT (oidc_subject) DO UPDATE SET
    email         = EXCLUDED.email,
    name          = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    admin         = TRUE,
    updated_at    = NOW()
RETURNING id, oidc_subject, email, name, admin;

-- name: GetUserByEmail :one
SELECT id, oidc_subject, email, name, password_hash, admin
FROM users
WHERE lower(email) = lower($1);

-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetUserBySessionHash :one
SELECT u.id, u.oidc_subject, u.email, u.name, u.admin
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = $1
  AND s.expires_at > NOW();

-- name: DeleteSessionByHash :exec
DELETE FROM sessions WHERE token_hash = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= NOW();

-- name: CreateAPIKey :one
INSERT INTO api_keys (id, name, prefix, secret_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, name, prefix, created_at, last_used_at, revoked_at;

-- name: ListAPIKeys :many
SELECT id, name, prefix, created_at, last_used_at, revoked_at
FROM api_keys
ORDER BY created_at DESC;

-- name: GetActiveAPIKeyByHash :one
SELECT id, name, prefix, created_at, last_used_at, revoked_at
FROM api_keys
WHERE secret_hash = $1
  AND revoked_at IS NULL;

-- name: RevokeAPIKey :execrows
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;

-- name: TouchAPIKey :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1;
