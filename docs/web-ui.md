# Web UI

A SvelteKit admin dashboard lives in
[`web`](https://github.com/picunada/flagcel/tree/main/web) and is embedded into
the Go binary at build time. Once built, it is served at `/` alongside the API.

```sh
# One-time
make web-install

# Dev: two processes
make docker-up   # backend on :8080 with Postgres
make web-dev     # frontend on :5173, proxies /api -> :8080

# Prod: single binary with embedded UI
make build       # pnpm build + go build -> bin/flagcel
```

When the binary is built without running `pnpm build` first, the UI route serves
a placeholder page pointing at `/docs`.

## Local Development Flow

Use the SvelteKit dev server when changing dashboard code. It runs on
`localhost:5173` and proxies API requests to the Go service on `localhost:8080`.
For production builds, `make build` produces one Go binary with the compiled UI
embedded.
