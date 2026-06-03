# Flagcel Go SDK

OpenFeature provider for Flagcel with local, native CEL evaluation. The provider polls Flagcel for `/eval/definitions`, compiles definitions with `evalcore`, and evaluates flags in-process.

## Install

```sh
go get github.com/picunada/flagcel/sdks/go
go get github.com/open-feature/go-sdk/openfeature
```

When using an unreleased local checkout, add replaces for both nested modules:

```sh
go mod edit -replace github.com/picunada/flagcel/sdks/go=/path/to/flagcel/sdks/go
go mod edit -replace github.com/picunada/flagcel/evalcore=/path/to/flagcel/evalcore
go mod tidy
```

## Basic Usage

Create an evaluation API key in the Flagcel dashboard, then configure the provider with the API base URL. For the local server, the base URL is `http://localhost:8080/api/v1`.

```go
package main

import (
	"context"
	"log"
	"time"

	flagcel "github.com/picunada/flagcel/sdks/go"
	"github.com/open-feature/go-sdk/openfeature"
)

func main() {
	provider, err := flagcel.NewProvider(
		"http://localhost:8080/api/v1",
		"fc_your_api_key",
		flagcel.WithPollInterval(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Shutdown()

	if err := openfeature.SetProviderAndWait(provider); err != nil {
		log.Fatal(err)
	}

	client := openfeature.NewClient("checkout-service")

	evalCtx := openfeature.NewTargetlessEvaluationContext(map[string]any{
		"user": map[string]any{
			"id":      "u_123",
			"country": "US",
			"plan":    "pro",
		},
		"request": map[string]any{
			"path": "/checkout",
		},
	})

	enabled, err := client.BooleanValue(
		context.Background(),
		"new-checkout",
		false,
		evalCtx,
	)
	if err != nil {
		// OpenFeature returns the supplied default when evaluation fails.
		log.Printf("flag evaluation used default: %v", err)
	}

	if enabled {
		// show the new checkout
	}
}
```

## Evaluation Context

Flagcel rules run against the attributes you pass in the OpenFeature evaluation context. Keep the same shape as your CEL rules:

```go
evalCtx := openfeature.NewTargetlessEvaluationContext(map[string]any{
	"user": map[string]any{
		"id":      "u_123",
		"country": "US",
	},
	"request": map[string]any{
		"path": "/checkout",
	},
})
```

That context supports rules such as:

```cel
user.country == "US" && request.path.startsWith("/checkout")
```

For rollout bucketing, pass the value referenced by the rule's `bucket_by`, for example `user.id`.

### OpenFeature Targeting Context

The provider supports OpenFeature targeting through evaluation context. You can set context at the global, client, transaction, or invocation level; the OpenFeature Go SDK merges those contexts before calling the Flagcel provider.

```go
openfeature.SetEvaluationContext(openfeature.NewTargetlessEvaluationContext(
	map[string]any{
		"region": "us-east-1-iah-1a",
	},
))

client := openfeature.NewClient("my-app")
client.SetEvaluationContext(openfeature.NewTargetlessEvaluationContext(
	map[string]any{
		"version": "1.4.6",
	},
))

evalCtx := openfeature.NewEvaluationContext(
	"user-123",
	map[string]any{
		"company": "Initech",
	},
)

enabled, err := client.BooleanValue(
	context.Background(),
	"boolFlag",
	false,
	evalCtx,
)
```

Flagcel receives the merged targeting data as CEL variables:

```go
map[string]any{
	"region":       "us-east-1-iah-1a",
	"version":      "1.4.6",
	"company":      "Initech",
	"targetingKey": "user-123",
}
```

That supports rules such as:

```cel
region == "us-east-1-iah-1a" &&
version == "1.4.6" &&
company == "Initech" &&
targetingKey == "user-123"
```

If the same attribute exists in multiple contexts, OpenFeature precedence is invocation over client, transaction, then global. The `targetingKey` is exposed as a top-level `targetingKey` field; it is not automatically copied to `user.id`. Use `bucket_by: "targetingKey"` for rollout bucketing by OpenFeature targeting key, or include the same value in your own context shape as `user.id`.

When a flag has a context schema, include any top-level targeting fields used by CEL rules, such as `region`, `version`, `company`, or `targetingKey`, in that schema.

## Typed Flags

Use the matching OpenFeature method for the Flagcel flag type:

| Flagcel type | OpenFeature method |
| --- | --- |
| `boolean` | `BooleanValue` / `BooleanValueDetails` |
| `string` | `StringValue` / `StringValueDetails` |
| `number` | `FloatValue`, `IntValue`, or details variants |
| `json` | `ObjectValue` / `ObjectValueDetails` |

If the requested OpenFeature type does not match the Flagcel flag type, the provider returns the caller's default value with a `TYPE_MISMATCH` error.

## Polling And Fail-Open Behavior

On `Init`, the provider performs one best-effort fetch and starts a background poller. Each poll:

- sends `Authorization: Bearer <api key>` when an API key is configured
- sends `If-None-Match` with the last ETag
- keeps the current in-memory evaluator when the server returns `304 Not Modified`
- keeps the last-known evaluator when a fetch fails

If the first fetch fails before any definitions have been loaded, evaluations return the OpenFeature default value with a provider-not-ready error. After a successful fetch, later fetch failures fail open to the last-known definitions.

## Options

```go
provider, err := flagcel.NewProvider(
	"https://flagcel.example.com/api/v1",
	apiKey,
	flagcel.WithPollInterval(10*time.Second),
	flagcel.WithHTTPClient(customHTTPClient),
)
```

| Option | Description |
| --- | --- |
| `WithPollInterval` | Changes the background definitions polling interval. Defaults to `30s`. |
| `WithHTTPClient` | Sets the HTTP client used for polling. Useful for custom timeouts, transports, or tests. |

## Resolution Details

Use OpenFeature details methods when you need the reason, variant, or error code:

```go
details, err := client.BooleanValueDetails(
	context.Background(),
	"new-checkout",
	false,
	evalCtx,
)
if err != nil {
	log.Printf("evaluation error: %s", err)
}
log.Printf("value=%t reason=%s variant=%s", details.Value, details.Reason, details.Variant)
```

Flagcel reasons are mapped to OpenFeature reasons:

| Flagcel reason | OpenFeature reason |
| --- | --- |
| `matched_rule` | `TARGETING_MATCH` |
| `default_no_match` | `DEFAULT` |
| `disabled` | `DISABLED` |
| `not_found` | `ERROR` with `FLAG_NOT_FOUND` |
| `cel_error` | `ERROR` with `GENERAL` |

The raw Flagcel reason is also available in flag metadata as `flagcelReason`.
