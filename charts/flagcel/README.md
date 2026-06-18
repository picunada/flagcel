# Flagcel Helm chart

Deploy the [Flagcel](https://github.com/picunada/flagcel) server to Kubernetes.

The chart is published as an OCI artifact to GHCR.

## Install

```sh
# External Postgres (recommended for production)
helm install flagcel oci://ghcr.io/picunada/charts/flagcel \
  --set database.url="postgres://user:pass@my-db:5432/flagcel?sslmode=require" \
  --set auth.cookieSecure=true
```

For a quick trial with a bundled Postgres:

```sh
helm install flagcel oci://ghcr.io/picunada/charts/flagcel \
  --set postgresql.enabled=true \
  --set auth.bootstrapAdminEmail=admin@example.com \
  --set auth.bootstrapAdminPassword=change-me \
  --set auth.cookieSecure=false
```

## Database

Provide exactly one of:

- `database.url` — DSN stored in a chart-managed Secret.
- `database.existingSecret` (+ `database.existingSecretKey`, default `DATABASE_URL`) — reference your own Secret.
- `postgresql.enabled=true` — deploy the bundled Bitnami Postgres subchart (demo use).

## Migrations

The server applies pending migrations on boot (`config.migrateOnStartup=true`).
To manage them out of band, set `config.migrateOnStartup=false` and run:

```sh
kubectl exec deploy/flagcel -- /flagcel migrate up
```

## Key values

| Key | Default | Description |
| --- | --- | --- |
| `image.repository` | `ghcr.io/picunada/flagcel` | Server image |
| `image.tag` | `""` (chart appVersion) | Image tag override |
| `replicaCount` | `1` | Number of replicas |
| `service.port` | `8080` | Service port |
| `ingress.enabled` | `false` | Create an Ingress |
| `config.logLevel` / `config.logFormat` | `info` / `json` | Logging |
| `config.migrateOnStartup` | `true` | Auto-migrate on boot |
| `database.url` | `""` | Postgres DSN |
| `database.existingSecret` | `""` | Existing Secret with the DSN |
| `auth.cookieSecure` | `true` | Secure session cookies (needs HTTPS) |
| `auth.sessionSecret` | `""` (generated) | Session signing secret |
| `auth.bootstrapAdminEmail` / `...Password` | `""` | First-boot local admin |
| `auth.oidc.*` | `""` | OIDC issuer/client config |
| `postgresql.enabled` | `false` | Bundle a Postgres subchart |

See [`values.yaml`](values.yaml) for the full list.
