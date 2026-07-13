<div class="fc-breadcrumb"><span>Operate</span><span>/</span><strong>Deployment</strong></div>

# Deployment

Flagcel ships as a single self-contained binary: the SvelteKit dashboard, SQL
migrations, and OpenAPI spec are embedded, so the only runtime dependency is a
Postgres database. By default the server applies pending migrations on boot
(`MIGRATE_ON_STARTUP=true`).

Pick whichever distribution fits your environment. See
[Configuration](configuration.md) for the full environment variable reference.

## Container image

Multi-arch (amd64 + arm64) images are published on every release to both GHCR
and Docker Hub:

```sh
docker pull ghcr.io/picunada/flagcel:latest
# or
docker pull docker.io/picunada/flagcel:latest
```

Pin a version in production (e.g. `:0.1.0`) rather than `:latest`.

```sh
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/flagcel?sslmode=require" \
  -e AUTH_SESSION_SECRET="$(openssl rand -hex 32)" \
  -e AUTH_BOOTSTRAP_ADMIN_EMAIL="admin@example.com" \
  -e AUTH_BOOTSTRAP_ADMIN_PASSWORD="change-me" \
  ghcr.io/picunada/flagcel:latest
```

## Docker Compose

A production-style Compose file that pulls the published image and runs Postgres
lives at [`examples/deploy/docker-compose.yml`](https://github.com/picunada/flagcel/blob/main/examples/deploy/docker-compose.yml):

```sh
AUTH_SESSION_SECRET=$(openssl rand -hex 32) \
docker compose -f examples/deploy/docker-compose.yml up -d
```

## Kubernetes (Helm)

The chart is published as an OCI artifact to GHCR.

```sh
# External Postgres (recommended for production)
helm install flagcel oci://ghcr.io/picunada/charts/flagcel \
  --set database.url="postgres://user:pass@my-db:5432/flagcel?sslmode=require" \
  --set auth.cookieSecure=true
```

For a quick trial with a bundled Postgres subchart:

```sh
helm install flagcel oci://ghcr.io/picunada/charts/flagcel \
  --set postgresql.enabled=true \
  --set auth.bootstrapAdminEmail=admin@example.com \
  --set auth.bootstrapAdminPassword=change-me \
  --set auth.cookieSecure=false
```

See the [chart README](https://github.com/picunada/flagcel/tree/main/charts/flagcel)
for all values, secret handling, and migration options.

## Prebuilt binary

Each release attaches cross-compiled binaries (linux/macOS, amd64/arm64) with a
`checksums.txt` to the [GitHub Releases](https://github.com/picunada/flagcel/releases)
page.

```sh
# Download and extract the archive for your platform, then:
DATABASE_URL="postgres://user:pass@host:5432/flagcel?sslmode=require" \
AUTH_SESSION_SECRET="$(openssl rand -hex 32)" \
./flagcel
```

The same binary runs migrations explicitly:

```sh
./flagcel migrate up
```

## Production checklist

- Set `DATABASE_URL` and a stable `AUTH_SESSION_SECRET` (≥32 bytes).
- Configure either OIDC SSO or local bootstrap admin credentials
  (see [Authentication](auth.md)).
- Set `AUTH_COOKIE_SECURE=true` when serving over HTTPS.
- Pin an image/chart/binary version rather than tracking `latest`.
