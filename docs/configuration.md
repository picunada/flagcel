# Configuration

All server config is provided through environment variables.

| Variable | Default | Description |
| --- | --- | --- |
| `DATABASE_URL` | _required_ | Postgres connection string |
| `PORT` | `8080` | HTTP listen port |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT` | `json` | `json` or `text` |
| `MIGRATE_ON_STARTUP` | `true` | Apply pending migrations on boot |
| `HTTP_READ_TIMEOUT` | `5s` | Request read timeout |
| `HTTP_WRITE_TIMEOUT` | `5s` | Response write timeout |
| `HTTP_IDLE_TIMEOUT` | `10s` | Keep-alive idle timeout |
| `HTTP_SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown deadline |
| `AUTH_OIDC_ISSUER_URL` | _empty_ | OIDC issuer URL. When empty, local password auth is used |
| `AUTH_OIDC_CLIENT_ID` | _empty_ | OIDC client ID |
| `AUTH_OIDC_CLIENT_SECRET` | _empty_ | OIDC client secret |
| `AUTH_OIDC_REDIRECT_URL` | _empty_ | OIDC callback URL, for example `https://flagcel.example.com/auth/callback` |
| `AUTH_ADMIN_EMAILS` | _empty_ | Comma-separated allowlist for admin SSO users |
| `AUTH_BOOTSTRAP_ADMIN_EMAIL` | _empty_ | Local admin email used when OIDC is not configured |
| `AUTH_BOOTSTRAP_ADMIN_PASSWORD` | _empty_ | Local admin password used when OIDC is not configured |
| `AUTH_BOOTSTRAP_ADMIN_NAME` | `Admin` | Local admin display name |
| `AUTH_SESSION_SECRET` | _empty_ | At least 32 bytes; used to hash sessions and API keys |
| `AUTH_COOKIE_SECURE` | `false` | Set secure cookies; use `true` behind HTTPS |
| `AUTH_SESSION_TTL` | `24h` | Admin session lifetime |
| `DEBUG_ADDR` | _empty_ | Optional pprof debug server address, for example `:16000` |

See [Authentication](auth.md) for auth-specific behavior.
