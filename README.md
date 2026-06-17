# Flagcel

Self-hosted feature flags with [CEL](https://github.com/google/cel-spec)-based targeting rules.

> **Status: experimental.** Flagcel is early in development. The API, storage schema, and SDK behavior may change without notice. Pin a commit if you depend on it.

## What It Is

Flagcel is a feature flag service designed to run in your own infrastructure. It ships as a Go service backed by Postgres, includes an admin dashboard, exposes an HTTP API, and supports OpenFeature SDK providers.

Targeting rules are written in CEL, the same expression language used by projects like Kubernetes and Envoy. A rule can use application context directly:

```cel
user.country == "US" && request.path.startsWith("/checkout")
```

## Why

Most feature flag systems either depend on a hosted SaaS control plane or introduce a custom rule language. Flagcel keeps the control plane self-hosted and uses a small, sandboxed expression language that is already widely used in infrastructure software.

## Project Shape

- **Server:** Go HTTP service with Postgres persistence.
- **Targeting:** CEL expressions plus deterministic percentage rollouts.
- **Dashboard:** SvelteKit admin UI embedded into the Go binary for production builds.
- **SDKs:** OpenFeature providers for Go, JS/TS, and Python.
- **Evaluation core:** Shared Go evaluator used by the server and native Go SDK.
- **API docs:** OpenAPI spec served by the running service.

## Quickstart

```sh
docker compose up
```

This starts Postgres and the Flagcel service on <http://localhost:8080>. The local Docker setup bootstraps an admin account:

```text
admin@localhost / secret
```

See [Quickstart](docs/src/quickstart.md) for creating a flag, evaluating it, and using the local API.

## Documentation

- [Documentation index](docs/src/index.md)
- [Quickstart](docs/src/quickstart.md)
- [Concepts](docs/src/concepts.md)
- [Configuration](docs/src/configuration.md)
- [Authentication](docs/src/auth.md)
- [API](docs/src/api.md)
- [SDKs](docs/src/sdks.md)
- [Web UI](docs/src/web-ui.md)
- [Migrations](docs/src/migrations.md)
- [Development](docs/src/development.md)
- [Examples](examples/README.md)

## Repository Layout

- [`cmd/server`](cmd/server) - server entrypoint and migration command.
- [`internal`](internal) - API handlers, services, config, storage, and engine integration.
- [`evalcore`](evalcore) - shared flag evaluation library.
- [`web`](web) - SvelteKit dashboard.
- [`sdks`](sdks) - OpenFeature providers.
- [`examples`](examples) - runnable SDK examples.
- [`docs/src`](docs/src) - project documentation.

## Roadmap

- Stabilize the published SDKs toward a 1.0 release.
- Add a Helm chart for Kubernetes installs.
- Expand production deployment guidance.

## Contributing

Issues and PRs are welcome. Before opening a PR, run:

```sh
go build ./...
go test ./...
```

Development setup details live in [Development](docs/src/development.md).

## License

Apache 2.0 - see [LICENSE](LICENSE).
