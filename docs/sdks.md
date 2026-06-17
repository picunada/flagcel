# SDKs

Flagcel SDKs are OpenFeature providers. Go and Python evaluate locally from polled definitions; JS/TS resolves flags through the Flagcel server-side evaluation API.

Runnable local examples for all three SDKs live in [`../examples`](../examples).

## Go

[![Go Reference](https://pkg.go.dev/badge/github.com/picunada/flagcel/sdks/go.svg)](https://pkg.go.dev/github.com/picunada/flagcel/sdks/go)

The Go SDK is an [OpenFeature](https://openfeature.dev/) provider in [`../sdks/go`](../sdks/go). It polls `GET /api/v1/eval/definitions` with an evaluation API key, compiles definitions with native `cel-go` through `evalcore`, and evaluates flags locally.

```go
provider, err := flagcel.NewProvider("http://localhost:8080/api/v1", apiKey)
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

See [`../sdks/go/README.md`](../sdks/go/README.md) for installation, OpenFeature targeting context, polling, fail-open behavior, and typed evaluation details.

## JS/TS

[![npm](https://img.shields.io/npm/v/@flagcel/openfeature-server.svg)](https://www.npmjs.com/package/@flagcel/openfeature-server)

The JS/TS SDK is an OpenFeature server provider in [`../sdks/js`](../sdks/js). It packages `@flagcel/openfeature-server` and resolves flags by calling the Flagcel evaluation API from Node server runtimes.

```ts
import { OpenFeature } from "@openfeature/server-sdk";
import { FlagcelProvider } from "@flagcel/openfeature-server";

await OpenFeature.setProviderAndWait(new FlagcelProvider({
	endpoint: "http://localhost:8080/api/v1",
	apiKey,
}));

const client = OpenFeature.getClient("checkout-service");
const enabled = await client.getBooleanValue("new-checkout", false, {
	targetingKey: "u_123",
	user: { id: "u_123", country: "US" },
});
```

See [`../sdks/js/README.md`](../sdks/js/README.md) for installation, runtime requirements, targeting context, and failure behavior.

## Python

[![PyPI](https://img.shields.io/pypi/v/flagcel-openfeature.svg)](https://pypi.org/project/flagcel-openfeature/)

The Python SDK is an OpenFeature provider in [`../sdks/python`](../sdks/python). It packages `flagcel-openfeature`, polls `GET /api/v1/eval/definitions` with an evaluation API key, compiles definitions with `cel-expr-python`, and evaluates flags locally.

```python
from openfeature import api
from openfeature.evaluation_context import EvaluationContext
from flagcel_openfeature import FlagcelProvider

api.set_provider(FlagcelProvider(
	endpoint="http://localhost:8080/api/v1",
	api_key=api_key,
))

client = api.get_client("checkout-service")
enabled = client.get_boolean_value("new-checkout", False, EvaluationContext(
	targeting_key="u_123",
	attributes={"user": {"id": "u_123", "country": "US"}},
))
```

See [`../sdks/python/README.md`](../sdks/python/README.md) for installation, targeting context, polling, fail-open behavior, and typed evaluation details.
