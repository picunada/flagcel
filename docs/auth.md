# Authentication

Flagcel always protects the dashboard and management API.

## Admin Sign-In

If `AUTH_OIDC_ISSUER_URL`, `AUTH_OIDC_CLIENT_ID`, `AUTH_OIDC_CLIENT_SECRET`, and `AUTH_OIDC_REDIRECT_URL` are set, admins sign in through generic OIDC SSO. Their verified email must appear in `AUTH_ADMIN_EMAILS`.

When OIDC is not configured, Flagcel uses local email/password auth. On startup it creates or updates the admin user from `AUTH_BOOTSTRAP_ADMIN_EMAIL`, `AUTH_BOOTSTRAP_ADMIN_PASSWORD`, and `AUTH_BOOTSTRAP_ADMIN_NAME`.

## Evaluation API Keys

Evaluation clients should use bearer API keys created from the dashboard's `keys` page. Each key is scoped to one environment, and that key selects the environment for `/eval` requests:

```sh
curl -X POST http://localhost:8080/api/v1/eval/new-checkout \
  -H "Authorization: Bearer fc_example_secret" \
  -H "Content-Type: application/json" \
  -d '{"context":{"user":{"id":"u_123","country":"US"}}}'
```

API keys and sessions are stored as HMAC-SHA-256 hashes. Raw API key tokens are shown only once when created.

See [Configuration](configuration.md) for auth environment variables.
