# Flagcel SDK Examples

Runnable examples for the local Flagcel SDKs live here:

- [`go`](go/) uses the Go OpenFeature provider.
- [`js`](js/) uses the Node server OpenFeature provider.
- [`python`](python/) uses the Python OpenFeature provider.

Each example reads the same environment variables:

| Variable | Default | Description |
| --- | --- | --- |
| `FLAGCEL_ENDPOINT` | `http://localhost:8080/api/v1` | Flagcel API base URL |
| `FLAGCEL_API_KEY` | _empty_ | Evaluation API key token |
| `FLAGCEL_FLAG_KEY` | `new-checkout` | Boolean flag key to evaluate |
| `TARGETING_KEY` | `u_123` | OpenFeature targeting key |
| `USER_ROLE` | `example` | `user.role` value passed to CEL |
| `USER_COUNTRY` | `US` | `user.country` value passed to CEL |
| `REQUEST_PATH` | `/checkout` | `request.path` value passed to CEL |

## Local Setup

Start Flagcel:

```sh
docker compose up
```

Create a boolean flag and an evaluation API key in the dashboard at
<http://localhost:8080>. In the default Docker setup, sign in as
`admin@localhost` with password `secret`.

For the examples below, create a boolean flag named `new-checkout` with:

- default value: `false`
- rule expression: `user.country == "US"`
- rule rollout: `100%`, bucketed by `user.id`
- rule value: `true`

Then create an API key from the `keys` page and export the raw token:

```sh
export FLAGCEL_API_KEY=fc_your_api_key
```

You can also set up the same flag and key with curl:

```sh
curl -c /tmp/flagcel-cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@localhost","password":"secret"}'

curl -b /tmp/flagcel-cookies.txt -X POST http://localhost:8080/api/v1/flags \
  -H "Content-Type: application/json" \
  -d '{
    "key": "new-checkout",
    "description": "SDK example flag",
    "type": "boolean",
    "enabled": true,
    "default_value": false,
    "rules": [
      {
        "description": "US checkout users",
        "expression": "user.country == \"US\"",
        "rollout": { "percentage": 100, "bucket_by": "user.id" },
        "value": true
      }
    ]
  }'

export FLAGCEL_API_KEY=$(
  curl -s -b /tmp/flagcel-cookies.txt -X POST http://localhost:8080/api/v1/api-keys \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"local SDK examples\",\"description\":\"Generated for local development\",\"environment_id\":\"$(curl -s -b /tmp/flagcel-cookies.txt http://localhost:8080/api/v1/environments | jq -r '.data[] | select(.key == \"production\") | .id')\"}" \
    | jq -r '.data.token'
)
```

## Go

```sh
cd examples/go
go run .
```

## JS/TS

From the repo root, build the local JS SDK package once, then run the example:

```sh
pnpm --dir sdks/js install
pnpm --dir sdks/js build
pnpm --dir examples/js install
pnpm --dir examples/js start
```

## Python

Install the local Python SDK dependency from `pyproject.toml`, then run the example:

```sh
cd examples/python
uv sync
uv run python main.py
```

For a local Python SDK load test:

```sh
cd examples/python
LOAD_CONCURRENCY=16 uv run python load_test.py
```
