# Flagcel

Self-hosted feature flag service with [CEL](https://github.com/google/cel-spec)-based targeting rules.

> **Status: experimental.** Early in development — the API and storage schema may change without notice. Pin a commit if you depend on it.

## Why

Most feature flag services either lock you into a SaaS or ship a DSL you have to learn. Flagcel runs in your own infrastructure as a single Go binary backed by Postgres, and it uses CEL — a small, sandboxed expression language already used by Kubernetes and Envoy — for targeting rules. If you can write `user.country == "US" && request.path.startsWith("/checkout")`, you can write Flagcel rules.

## Roadmap

- **Client SDKs (Go, JS/TS)** — typed clients that wrap the HTTP API and handle local evaluation caching.
- **Helm chart** — single-command install on Kubernetes with sensible production defaults.

## Quickstart

### Run the service

```sh
docker compose up
```

Brings up Postgres + the service with hot reload on port `8080`. API docs are served at <http://localhost:8080/docs>.

### Create and read a flag

The Docker quickstart bootstraps a local admin account:
`admin@localhost` / `flagcel-dev-password`.

```sh
# Sign in and keep the admin session cookie
curl -c cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@localhost","password":"flagcel-dev-password"}'

# Create a flag with one rule (10% rollout for US users)
curl -X POST http://localhost:8080/api/v1/flags \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "key": "new-checkout",
    "enabled": true,
    "default_value": false,
    "rules": [
      {
        "expression": "user.country == \"US\"",
        "rollout": { "percentage": 10, "bucket_by": "user.id" }
      }
    ]
  }'

# Read it back
curl -b cookies.txt http://localhost:8080/api/v1/flags/new-checkout
```

## Configuration

All config is via environment variables.

| Variable                 | Default    | Description                         |
| ------------------------ | ---------- | ----------------------------------- |
| `DATABASE_URL`           | _required_ | Postgres connection string          |
| `PORT`                   | `8080`     | HTTP listen port                    |
| `LOG_LEVEL`              | `info`     | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT`             | `json`     | `json` or `text`                    |
| `MIGRATE_ON_STARTUP`     | `true`     | Apply pending migrations on boot    |
| `HTTP_READ_TIMEOUT`      | `5s`       | Request read timeout                |
| `HTTP_WRITE_TIMEOUT`     | `5s`       | Response write timeout              |
| `HTTP_IDLE_TIMEOUT`      | `10s`      | Keep-alive idle timeout             |
| `HTTP_SHUTDOWN_TIMEOUT`  | `15s`      | Graceful shutdown deadline          |
| `AUTH_OIDC_ISSUER_URL`   | _empty_    | OIDC issuer URL. When empty, local password auth is used |
| `AUTH_OIDC_CLIENT_ID`    | _empty_    | OIDC client ID                      |
| `AUTH_OIDC_CLIENT_SECRET`| _empty_    | OIDC client secret                  |
| `AUTH_OIDC_REDIRECT_URL` | _empty_    | OIDC callback URL, e.g. `https://flagcel.example.com/auth/callback` |
| `AUTH_ADMIN_EMAILS`      | _empty_    | Comma-separated allowlist for admin SSO users |
| `AUTH_BOOTSTRAP_ADMIN_EMAIL` | _empty_ | Local admin email used when OIDC is not configured |
| `AUTH_BOOTSTRAP_ADMIN_PASSWORD` | _empty_ | Local admin password used when OIDC is not configured |
| `AUTH_BOOTSTRAP_ADMIN_NAME` | `Admin` | Local admin display name |
| `AUTH_SESSION_SECRET`    | _empty_    | At least 32 bytes; used to hash sessions and API keys |
| `AUTH_COOKIE_SECURE`     | `false`    | Set secure cookies; use `true` behind HTTPS |
| `AUTH_SESSION_TTL`       | `24h`      | Admin session lifetime              |

## Auth

Flagcel always protects the dashboard and management API. If
`AUTH_OIDC_ISSUER_URL`, `AUTH_OIDC_CLIENT_ID`, `AUTH_OIDC_CLIENT_SECRET`, and
`AUTH_OIDC_REDIRECT_URL` are set, admins sign in through generic OIDC SSO and
their verified email must appear in `AUTH_ADMIN_EMAILS`.

When OIDC is not configured, Flagcel uses local email/password auth. On startup
it creates or updates the admin user from `AUTH_BOOTSTRAP_ADMIN_EMAIL`,
`AUTH_BOOTSTRAP_ADMIN_PASSWORD`, and `AUTH_BOOTSTRAP_ADMIN_NAME`.

Evaluation clients should use bearer API keys created from the dashboard's
`keys` page:

```sh
curl -X POST http://localhost:8080/api/v1/eval/new-checkout \
  -H "Authorization: Bearer fc_example_secret" \
  -H "Content-Type: application/json" \
  -d '{"context":{"user":{"id":"u_123","country":"US"}}}'
```

API keys and sessions are stored as HMAC-SHA-256 hashes; raw API key tokens are
shown only once when created.

## API

The full OpenAPI spec lives at [`internal/api/http/docs/openapi.yaml`](internal/api/http/docs/openapi.yaml) and is served live:

- `GET /openapi.yaml` — raw spec
- `GET /docs` — Swagger UI

Endpoint overview (all under `/api/v1`):

```
GET    /auth/me
POST   /auth/login
POST   /auth/logout

GET    /flags
POST   /flags
GET    /flags/{key}
DELETE /flags/{key}

GET    /flags/{key}/rules
POST   /flags/{key}/rules
POST   /flags/{key}/rules/reorder
GET    /flags/{key}/rules/{id}
PUT    /flags/{key}/rules/{id}
DELETE /flags/{key}/rules/{id}

GET    /contexts
POST   /contexts
GET    /contexts/{id}
PUT    /contexts/{id}
DELETE /contexts/{id}

POST   /eval
POST   /eval/{key}

GET    /api-keys
POST   /api-keys
DELETE /api-keys/{id}
```

## Web UI

A SvelteKit admin dashboard lives in [`web/`](web/) and is embedded into the Go binary at build time. Once built, it is served at `/` alongside the API.

```sh
# One-time
make web-install

# Dev: two processes
make docker-up   # backend on :8080 (with Postgres)
make web-dev     # frontend on :5173, proxies /api -> :8080

# Prod: single binary with embedded UI
make build       # pnpm build + go build -> bin/flagcel
```

When the binary is built without running `pnpm build` first, the UI route serves a placeholder page pointing at `/docs`.

## Migrations

Schema changes live in [`internal/store/postgres/migrations/`](internal/store/postgres/migrations/) and are managed with [goose](https://github.com/pressly/goose). Migrations are embedded into the binary at build time.

By default the server applies pending migrations on startup. For production deployments where you want to run migrations out-of-band, set `MIGRATE_ON_STARTUP=false` and use the `migrate` subcommand:

```sh
flagcel migrate up       # apply all pending
flagcel migrate down     # roll back the most recent
flagcel migrate status   # show applied / pending
flagcel migrate version  # print current version
```

The same targets are exposed via `make migrate-up`, `make migrate-status`, etc.

To add a new migration, create `internal/store/postgres/migrations/0000N_name.sql`:

```sql
-- +goose Up
ALTER TABLE flags ADD COLUMN description TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE flags DROP COLUMN description;
```

## Development

Requires Go 1.26+ and Postgres 17 (the Docker setup handles both).

```sh
# Hot-reload dev environment (air + postgres)
docker compose up

# Or run directly against your own Postgres
DATABASE_URL=postgres://localhost/flagcel?sslmode=disable \
AUTH_BOOTSTRAP_ADMIN_EMAIL=admin@localhost \
AUTH_BOOTSTRAP_ADMIN_PASSWORD=flagcel-dev-password \
AUTH_SESSION_SECRET=flagcel-dev-session-secret-change-me \
go run ./cmd/server
```

The `Dockerfile.dev` uses [air](https://github.com/air-verse/air) to rebuild on file change.

## Contributing

Issues and PRs are welcome. Before opening a PR, run:

```sh
go build ./...
go test ./...
```

## License

Apache 2.0 — see [LICENSE](LICENSE).
