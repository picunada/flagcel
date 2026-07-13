<div class="fc-breadcrumb"><span>Get Started</span><span>/</span><strong>Concepts</strong></div>

# Concepts

Flagcel evaluates feature flags from a small set of concepts: environments,
flags, contexts, rules, CEL expressions, rollouts, and evaluation clients.

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
- an optional context schema
- zero or more ordered rules

When a flag is disabled or no rule matches, evaluation returns the default
value.

## Contexts

A context is a reusable schema that describes the evaluation payload your
clients send. It is not the runtime request itself - it declares the fields and
types that rules may read.

Each context has:

- a unique name, such as `web-user`
- an optional description
- a list of fields with a dotted path and type

Supported field types are `string`, `int`, `double`, `bool`, `timestamp`,
`list`, and `map`. Nested JSON becomes dotted paths, so this payload:

```json
{
  "user": { "id": "usr_991", "plan": "pro", "beta": true },
  "device": { "os": "ios", "version": 17 }
}
```

maps to fields like:

```text
user.id        string
user.plan      string
user.beta      bool
device.os      string
device.version int
```

Contexts are global across environments. Attach one to a flag with `context_id`
when you create or edit the flag. Once attached, Flagcel uses the schema to:

- validate CEL expressions and `bucket_by` paths against known fields
- power autocomplete and payload checks in the dashboard
- package the schema with evaluation definitions for typed local evaluation

Flags can omit a context. Without one, rules are still evaluated, but Flagcel
cannot check that referenced paths exist or match a declared type.

Schema changes that would break attached flags are rejected. You also cannot
delete a context while any flag still references it. Manage contexts in the
dashboard under **Contexts**, or through the `/contexts` admin API.

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

The runtime evaluation context is application-defined JSON. If a rule reads
`user.country`, the client must send a compatible `user` object:

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

When the flag has an attached context schema, keep that payload aligned with
the declared fields and types. The schema documents the contract; the SDK or
`/eval` request still supplies the live values.

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

See [SDKs](sdks/) for provider examples and [API](api.md) for endpoint
details.
