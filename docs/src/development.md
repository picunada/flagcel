# Development

Flagcel requires Go 1.26+ and Postgres 17. The Docker setup handles Postgres
and hot reload for local development.

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

## Documentation Site

The public documentation site is built with VitePress from the `docs`
directory.

```sh
cd docs
pnpm install
pnpm dev
pnpm build
pnpm preview
```

The GitHub Pages workflow deploys the generated site from
`docs/src/.vitepress/dist` to `https://picunada.github.io/flagcel/`.

To use a custom domain later, add `docs/public/CNAME` and change the VitePress
`base` setting from `/flagcel/` to `/`.

## Checks

Before opening a PR, run:

```sh
go build ./...
go test ./...
cd docs && pnpm build
```
