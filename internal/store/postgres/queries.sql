-- name: ListEnvironments :many
SELECT id, key, name, description, created_at, updated_at, created_by, deleted_by
FROM environments
ORDER BY key;

-- name: GetEnvironment :one
SELECT id, key, name, description, created_at, updated_at, created_by, deleted_by
FROM environments
WHERE id = $1;

-- name: GetEnvironmentByKey :one
SELECT id, key, name, description, created_at, updated_at, created_by, deleted_by
FROM environments
WHERE key = $1;

-- name: InsertEnvironment :exec
INSERT INTO environments (id, key, name, description)
VALUES ($1, $2, $3, $4);

-- name: UpdateEnvironment :execrows
UPDATE environments
SET key = $2,
    name = $3,
    description = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteEnvironment :execrows
DELETE FROM environments
WHERE id = $1;

-- name: GetFlag :one
SELECT environment_id, key, value_type, enabled, default_value, context_id, description, created_at, updated_at, created_by, updated_by, deleted_by
FROM flags
WHERE environment_id = $1 AND key = $2;

-- name: ListFlags :many
SELECT environment_id, key, value_type, enabled, default_value, context_id, description, created_at, updated_at, created_by, updated_by, deleted_by
FROM flags
WHERE environment_id = $1
ORDER BY key;

-- name: ListRulesForFlag :many
SELECT * FROM rules
WHERE environment_id = $1 AND flag_key = $2
ORDER BY position;

-- name: ListAllRules :many
SELECT * FROM rules
WHERE environment_id = $1
ORDER BY flag_key, position;

-- name: UpsertFlag :exec
INSERT INTO flags (environment_id, key, value_type, enabled, default_value, context_id, description, created_by, updated_by, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
ON CONFLICT (environment_id, key) DO UPDATE SET
    value_type    = EXCLUDED.value_type,
    enabled       = EXCLUDED.enabled,
    default_value = EXCLUDED.default_value,
    context_id    = EXCLUDED.context_id,
    description   = EXCLUDED.description,
    updated_by    = EXCLUDED.updated_by,
    updated_at    = NOW();

-- name: DeleteRulesForFlag :exec
DELETE FROM rules WHERE environment_id = $1 AND flag_key = $2;

-- name: InsertRule :exec
INSERT INTO rules (id, environment_id, flag_key, expression, rollout_percentage, rollout_bucket_by, position, value, description, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: DeleteFlag :exec
DELETE FROM flags WHERE environment_id = $1 AND key = $2;

-- name: TouchFlag :execrows
UPDATE flags
SET updated_by = $3,
    updated_at = NOW()
WHERE environment_id = $1 AND key = $2;

-- name: GetRule :one
SELECT * FROM rules
WHERE environment_id = $1 AND flag_key = $2 AND id = $3;

-- name: InsertRuleAtEnd :exec
INSERT INTO rules (id, environment_id, flag_key, expression, rollout_percentage, rollout_bucket_by, position, value, description, created_by, updated_by)
VALUES (
    $1, $2, $3, $4, $5, $6,
    COALESCE((SELECT MAX(position) + 1 FROM rules WHERE environment_id = $2 AND flag_key = $3), 0),
    $7, $8, $9, $10
);

-- name: UpdateRule :execrows
UPDATE rules
SET expression         = $4,
    rollout_percentage = $5,
    rollout_bucket_by  = $6,
    value              = $7,
    description        = $8,
    updated_by         = $9,
    updated_at         = NOW()
WHERE environment_id = $1 AND flag_key = $2 AND id = $3;

-- name: DeleteRule :execrows
DELETE FROM rules
WHERE environment_id = $1 AND flag_key = $2 AND id = $3;

-- name: SetRulePosition :execrows
UPDATE rules
SET position = $4,
    updated_by = $5,
    updated_at = NOW()
WHERE environment_id = $1 AND flag_key = $2 AND id = $3;

-- name: InsertAuditLog :one
INSERT INTO audit_logs (id, environment_id, resource_type, resource_id, action, version, snapshot, actor_id, actor_label, summary)
VALUES (
    $1, $2, $3, $4, $5,
    COALESCE((SELECT MAX(version) FROM audit_logs WHERE environment_id = $2 AND resource_type = $3 AND resource_id = $4), 0) + 1,
    $6, $7, $8, $9
)
RETURNING id, environment_id, resource_type, resource_id, action, version, snapshot, actor_id, actor_label, summary, created_at;

-- name: ListFlagAuditLog :many
SELECT id, environment_id, resource_type, resource_id, action, version, snapshot, actor_id, actor_label, summary, created_at
FROM audit_logs
WHERE environment_id = $1 AND resource_type = 'flag' AND resource_id = $2
ORDER BY version DESC;

-- name: ListContexts :many
SELECT id, name, description, fields, created_at, updated_at, created_by, deleted_by
FROM contexts
ORDER BY name;

-- name: ListContextFlags :many
SELECT f.environment_id, e.key AS environment_key, f.key, f.value_type, f.enabled,
       f.default_value, f.context_id, f.description, f.created_at, f.updated_at,
       f.created_by, f.updated_by, f.deleted_by
FROM flags f
JOIN environments e ON e.id = f.environment_id
WHERE f.context_id IS NOT NULL
ORDER BY f.context_id, e.key, f.key;

-- name: ListContextRules :many
SELECT r.*
FROM rules r
JOIN flags f ON f.environment_id = r.environment_id AND f.key = r.flag_key
WHERE f.context_id IS NOT NULL
ORDER BY r.environment_id, r.flag_key, r.position;

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
INSERT INTO api_keys (id, name, description, prefix, secret_hash, environment_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, description, prefix, environment_id, created_at, updated_at, last_used_at, revoked_at, created_by, deleted_by;

-- name: ListAPIKeys :many
SELECT id, name, description, prefix, environment_id, created_at, updated_at, last_used_at, revoked_at, created_by, deleted_by
FROM api_keys
ORDER BY created_at DESC;

-- name: GetActiveAPIKeyByHash :one
SELECT id, name, description, prefix, environment_id, created_at, updated_at, last_used_at, revoked_at, created_by, deleted_by
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

-- name: UpsertFlagUsageBucket :exec
INSERT INTO flag_usage_buckets (
    environment_id,
    flag_key,
    bucket_start,
    value_type,
    value_key,
    value,
    reason,
    matched_rule_id,
    api_key_id,
    source,
    count,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
ON CONFLICT (
    environment_id,
    flag_key,
    bucket_start,
    value_key,
    reason,
    matched_rule_id,
    api_key_id,
    source
) DO UPDATE SET
    count = flag_usage_buckets.count + EXCLUDED.count,
    value = EXCLUDED.value,
    value_type = EXCLUDED.value_type,
    updated_at = NOW();

-- name: InsertFlagEvaluationEvent :exec
INSERT INTO flag_evaluation_events (
    id,
    environment_id,
    flag_key,
    observed_at,
    value_type,
    value,
    reason,
    matched_rule_id,
    api_key_id,
    source,
    latency_ms,
    context
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: ListFlagUsageBuckets :many
SELECT b.environment_id, b.flag_key, b.bucket_start, b.value_type, b.value, b.reason, b.matched_rule_id, b.api_key_id, COALESCE(ak.name, '')::text AS api_key_name, b.source, b.count
FROM flag_usage_buckets b
LEFT JOIN api_keys ak ON ak.id::text = b.api_key_id
WHERE b.environment_id = $1
  AND b.flag_key = $2
  AND b.bucket_start >= $3
ORDER BY b.bucket_start DESC, b.reason, b.matched_rule_id, b.value_key;

-- name: ListEnvironmentUsageBuckets :many
SELECT b.environment_id, b.flag_key, b.bucket_start, b.value_type, b.value, b.reason, b.matched_rule_id, b.api_key_id, COALESCE(ak.name, '')::text AS api_key_name, b.source, b.count
FROM flag_usage_buckets b
LEFT JOIN api_keys ak ON ak.id::text = b.api_key_id
WHERE b.environment_id = $1
  AND b.bucket_start >= $2
ORDER BY b.bucket_start DESC, b.flag_key, b.reason, b.matched_rule_id, b.value_key;

-- name: ListFlagUsageLatencyBuckets :many
SELECT environment_id,
       flag_key,
       source,
       date_trunc('hour', observed_at)::timestamptz AS bucket_start,
       count(*)::bigint AS count,
       avg(latency_ms)::double precision AS avg_latency_ms,
       percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms)::double precision AS p95_latency_ms
FROM flag_evaluation_events
WHERE environment_id = $1
  AND flag_key = $2
  AND observed_at >= $3
GROUP BY environment_id, flag_key, source, bucket_start
ORDER BY bucket_start DESC, flag_key, source;

-- name: ListEnvironmentUsageLatencyBuckets :many
SELECT environment_id,
       flag_key,
       source,
       date_trunc('hour', observed_at)::timestamptz AS bucket_start,
       count(*)::bigint AS count,
       avg(latency_ms)::double precision AS avg_latency_ms,
       percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms)::double precision AS p95_latency_ms
FROM flag_evaluation_events
WHERE environment_id = $1
  AND observed_at >= $2
GROUP BY environment_id, flag_key, source, bucket_start
ORDER BY bucket_start DESC, flag_key, source;

-- name: ListFlagEvaluationEvents :many
SELECT id, environment_id, flag_key, observed_at, value_type, value, reason, matched_rule_id, api_key_id, source, latency_ms, context
FROM flag_evaluation_events
WHERE environment_id = $1
  AND flag_key = $2
ORDER BY observed_at DESC
LIMIT $3;
