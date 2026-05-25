-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id           UUID        PRIMARY KEY,
    oidc_subject TEXT        NOT NULL UNIQUE,
    email        TEXT        NOT NULL UNIQUE,
    name         TEXT        NOT NULL DEFAULT '',
    password_hash TEXT       NOT NULL DEFAULT '',
    admin        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id         UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash
    ON sessions(token_hash);

CREATE INDEX IF NOT EXISTS idx_sessions_expires_at
    ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS api_keys (
    id          UUID        PRIMARY KEY,
    name        TEXT        NOT NULL,
    prefix      TEXT        NOT NULL UNIQUE,
    secret_hash TEXT        NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    revoked_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_api_keys_secret_hash
    ON api_keys(secret_hash);

-- +goose Down
DROP INDEX IF EXISTS idx_api_keys_secret_hash;
DROP TABLE IF EXISTS api_keys;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
