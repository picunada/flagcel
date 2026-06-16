# Quickstart

## Run The Service

```sh
docker compose up
```

This starts Postgres and the Flagcel service with hot reload on port `8080`. The dashboard is available at <http://localhost:8080>, and API docs are served at <http://localhost:8080/docs>.

The Docker quickstart bootstraps a local admin account:

```text
admin@localhost / secret
```

## Create And Read A Flag

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

Evaluation clients should use bearer API keys created from the dashboard's `keys` page. Each key is scoped to one environment, and that key selects the environment for `/eval` requests.

```sh
curl -X POST http://localhost:8080/api/v1/eval/new-checkout \
  -H "Authorization: Bearer fc_example_secret" \
  -H "Content-Type: application/json" \
  -d '{"context":{"user":{"id":"u_123","country":"US"}}}'
```

For runnable SDK examples, see [`../examples`](../examples).
