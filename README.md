# Flagcel(l)

Self-hosted feature flag service with [CEL](https://github.com/google/cel-spec)-based targeting rules.

> **Status: experimental.** Early in development — the API and storage schema may change without notice. Pin a commit if you depend on it.

## Why

Most feature flag services either lock you into a SaaS or ship a DSL you have to learn. Flagcel runs in your own infrastructure as a single Go binary backed by Postgres, and it uses CEL — a small, sandboxed expression language already used by Kubernetes and Envoy — for targeting rules. If you can write `user.country == "US" && request.path.startsWith("/checkout")`, you can write Flagcel rules.

## Roadmap

- **Client SDKs (Go, JS/TS)** — typed clients that wrap the HTTP API and handle local evaluation caching.
- **Helm chart** — single-command install on Kubernetes with sensible production defaults.
- **Web UI** — admin dashboard for managing flags and rules without curl.

## Quickstart

### Run the service

```sh
docker compose up
```

Brings up Postgres + the service with hot reload on port `8080`. API docs are served at <http://localhost:8080/docs>.

### Create and read a flag

```sh
# Create a flag with one rule (10% rollout for US users)
curl -X POST http://localhost:8080/api/v1/flags \
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
curl http://localhost:8080/api/v1/flags/new-checkout
```

## Configuration

All config is via environment variables.

| Variable                 | Default    | Description                         |
| ------------------------ | ---------- | ----------------------------------- |
| `DATABASE_URL`           | _required_ | Postgres connection string          |
| `PORT`                   | `8080`     | HTTP listen port                    |
| `LOG_LEVEL`              | `info`     | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT`             | `json`     | `json` or `text`                    |
| `HTTP_READ_TIMEOUT`      | `5s`       | Request read timeout                |
| `HTTP_WRITE_TIMEOUT`     | `5s`       | Response write timeout              |
| `HTTP_IDLE_TIMEOUT`      | `10s`      | Keep-alive idle timeout             |
| `HTTP_SHUTDOWN_TIMEOUT`  | `15s`      | Graceful shutdown deadline          |

## API

The full OpenAPI spec lives at [`internal/api/http/docs/openapi.yaml`](internal/api/http/docs/openapi.yaml) and is served live:

- `GET /openapi.yaml` — raw spec
- `GET /docs` — Swagger UI

Endpoint overview (all under `/api/v1`):

```
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
```

## Development

Requires Go 1.26+ and Postgres 17 (the Docker setup handles both).

```sh
# Hot-reload dev environment (air + postgres)
docker compose up

# Or run directly against your own Postgres
DATABASE_URL=postgres://localhost/flagcel?sslmode=disable go run ./cmd/server
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
