<div class="fc-breadcrumb"><span>Guide</span><span>/</span><strong>Quickstart</strong></div>

# Quickstart

<p class="fc-lead">Start Flagcel locally, create a boolean flag, and evaluate it through the HTTP API.</p>

<div class="fc-tags">
  <span>≈ 5 min</span>
  <span>Docker</span>
  <span>Postgres 14+</span>
</div>

## Prerequisites

- Docker and Docker Compose.
- `curl`.
- `jq` is optional, but useful for API work.

## Run The Service

<div class="fc-process-block" aria-hidden="true"></div>

```sh
docker compose up
```

This starts Postgres and the Flagcel service with hot reload on port `8080`.
The dashboard is available at `http://localhost:8080`, and live API docs are
served at `http://localhost:8080/docs`.

The Docker quickstart bootstraps a local admin account:

```text
admin@localhost / secret
```

::: tip Local development
Run `flagcel dev` to spin up an ephemeral Postgres and seed a demo project.
:::

## Create A Flag

```sh
# Sign in and keep the admin session cookie
curl -c cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@localhost","password":"secret"}'

# Create a boolean flag with one rule
curl -X POST http://localhost:8080/api/v1/flags \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "key": "new-checkout",
    "description": "Checkout rollout",
    "type": "boolean",
    "enabled": true,
    "default_value": false,
    "rules": [
      {
        "description": "10% rollout for US users",
        "expression": "user.country == \"US\"",
        "rollout": { "percentage": 10, "bucket_by": "user.id" },
        "value": true
      }
    ]
  }'

# Read it back
curl -b cookies.txt http://localhost:8080/api/v1/flags/new-checkout
```

## Evaluate A Flag

Evaluation clients should use bearer API keys created from the dashboard's
`keys` page. Each key is scoped to one environment, and that key selects the
environment for `/eval` requests.

```sh
curl -X POST http://localhost:8080/api/v1/eval/new-checkout \
  -H "Authorization: Bearer fc_example_secret" \
  -H "Content-Type: application/json" \
  -d '{"context":{"user":{"id":"u_123","country":"US"}}}'
```

For runnable SDK examples, see the
[examples directory](https://github.com/picunada/flagcel/tree/main/examples).

## Next Steps

<div class="fc-nextgrid">
  <a class="fc-ncard" href="concepts">
    <div class="nk">Concepts</div>
    <div class="nh">CEL targeting</div>
    <div class="nd">Understand rules, evaluation context, and rollout bucketing.</div>
  </a>
  <a class="fc-ncard" href="auth">
    <div class="nk">Operate</div>
    <div class="nh">Authentication</div>
    <div class="nd">Prepare local auth or OIDC before exposing a deployment.</div>
  </a>
  <a class="fc-ncard" href="sdks">
    <div class="nk">SDKs</div>
    <div class="nh">Evaluate from your app</div>
    <div class="nd">Use OpenFeature providers from Go, TypeScript, and Python.</div>
  </a>
  <a class="fc-ncard" href="api">
    <div class="nk">Reference</div>
    <div class="nh">HTTP API</div>
    <div class="nd">Review endpoints for flags, environments, contexts, and eval.</div>
  </a>
</div>
