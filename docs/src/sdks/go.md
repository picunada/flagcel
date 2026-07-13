<div class="fc-breadcrumb"><span>SDKs</span><span>/</span><strong>Go</strong></div>

# Go

[![Go Reference](https://pkg.go.dev/badge/github.com/picunada/flagcel/sdks/go.svg)](https://pkg.go.dev/github.com/picunada/flagcel/sdks/go)

The Go SDK is an [OpenFeature](https://openfeature.dev/) provider in
[`sdks/go`](https://github.com/picunada/flagcel/tree/main/sdks/go). It polls
`GET /api/v1/eval/definitions` with an evaluation API key, compiles definitions
with native `cel-go` through `evalcore`, and evaluates flags locally.

## Install

```sh
go get github.com/picunada/flagcel/sdks/go
go get github.com/open-feature/go-sdk/openfeature
```

## Usage

Create an environment-scoped evaluation API key in the Flagcel dashboard, then
configure the provider with the API base URL. For the local server, that is
`http://localhost:8080/api/v1`.

```go
provider, err := flagcel.NewProvider(
	"http://localhost:8080/api/v1",
	apiKey,
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
enabled, err := client.BooleanValue(ctx, "new-checkout", false, openfeature.NewTargetlessEvaluationContext(map[string]any{
	"user": map[string]any{"id": "u_123", "country": "US"},
}))
```

## Behavior

- Polls definitions on a background interval (default `30s`)
- Uses `ETag` / `If-None-Match` to skip unchanged payloads
- Fails open to last-known definitions after a successful first fetch
- Returns OpenFeature defaults if no definitions have loaded yet
- Exposes OpenFeature `targetingKey` as a top-level CEL field

Keep evaluation context aligned with your
[context schemas](../concepts.md#contexts) and rules. For rollout bucketing by
OpenFeature targeting key, set `bucket_by` to `targetingKey`.

See the
[Go SDK README](https://github.com/picunada/flagcel/tree/main/sdks/go#readme)
for typed evaluation, resolution details, options, and more examples.
