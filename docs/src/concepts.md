# Concepts

Flagcel evaluates feature flags from a small set of concepts: environments,
flags, rules, CEL expressions, rollouts, and evaluation clients.

## Environments

An environment represents an isolated set of flag definitions, such as
`production` or `staging`. Evaluation API keys are scoped to exactly one
environment. SDKs and `/eval` callers do not pass the environment in every
request; the bearer token selects it.

The local Docker setup creates a default `production` environment.

## Feature Flags

A flag is identified by a stable key, such as `new-checkout`. Each flag has:

- a type, currently used by the API and SDKs to select typed evaluation
- an enabled state
- a default value
- zero or more ordered rules

When a flag is disabled or no rule matches, evaluation returns the default
value.

## Rules

Rules are evaluated in order. A rule contains:

- a CEL expression that decides whether the rule matches the request context
- an optional percentage rollout
- the value returned when the rule matches and passes rollout

The first matching rule wins. Keep high-priority or narrow rules above broader
rules.

## CEL Targeting

Flagcel uses CEL, the Common Expression Language, for targeting. Expressions
read fields from the evaluation context sent by the SDK or API client.

```cel
user.country == "US" && request.path.startsWith("/checkout")
```

The context shape is application-defined. If a rule reads `user.country`, the
client must send a compatible `user` object:

```json
{
  "context": {
    "user": {
      "id": "u_123",
      "country": "US"
    },
    "request": {
      "path": "/checkout"
    }
  }
}
```

## Percentage Rollouts

Rules can include deterministic percentage rollouts. A rollout uses a percentage
and a `bucket_by` field path, such as `user.id` or `targetingKey`.

```json
{
  "percentage": 10,
  "bucket_by": "user.id"
}
```

Use a stable identifier for `bucket_by` so the same user remains in the same
bucket across requests.

## Evaluation Flow

Application code should evaluate flags through an OpenFeature provider or the
server-side evaluation API.

- Go and Python providers poll `GET /api/v1/eval/definitions` and evaluate
  locally.
- The JavaScript/TypeScript provider calls the Flagcel evaluation API from Node
  server runtimes.
- Every evaluation request uses an environment-scoped API key.

See [SDKs](sdks.md) for provider examples and [API](api.md) for endpoint
details.
