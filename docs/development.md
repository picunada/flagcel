# Development

Flagcel requires Go 1.26+ and Postgres 17. The Docker setup handles Postgres and hot reload for local development.

## Run Locally

```sh
# Hot-reload dev environment with air and Postgres
docker compose up

# Or run directly against your own Postgres
DATABASE_URL=postgres://localhost/flagcel?sslmode=disable \
AUTH_BOOTSTRAP_ADMIN_EMAIL=admin@localhost \
AUTH_BOOTSTRAP_ADMIN_PASSWORD=flagcel-dev-password \
AUTH_SESSION_SECRET=flagcel-dev-session-secret-change-me \
go run ./cmd/server
```

The `Dockerfile.dev` uses [air](https://github.com/air-verse/air) to rebuild on file change.

## Common Commands

```sh
make web-install
make web-dev
make build
make migrate-status
```

Run `make help` for the full command list.

## Checks

Before opening a PR, run:

```sh
go build ./...
go test ./...
```
