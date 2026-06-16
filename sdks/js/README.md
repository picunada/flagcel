# Flagcel OpenFeature Server Provider

`@flagcel/openfeature-server` is a Node server provider for
`@openfeature/server-sdk`. It resolves flags by calling the Flagcel server-side
evaluation API.

## Install

```sh
npm install @flagcel/openfeature-server @openfeature/server-sdk
```

## Usage

```ts
import { OpenFeature } from "@openfeature/server-sdk";
import { FlagcelProvider } from "@flagcel/openfeature-server";

const provider = new FlagcelProvider({
  endpoint: "http://localhost:8080/api/v1",
  apiKey: "fc_your_api_key",
});

await OpenFeature.setProviderAndWait(provider);

const client = OpenFeature.getClient("checkout-service");
const enabled = await client.getBooleanValue("new-checkout", false, {
  targetingKey: "u_123",
  user: {
    id: "u_123",
    country: "US",
  },
});
```

## Targeting

OpenFeature evaluation context is sent to Flagcel's CEL evaluator. Keep the
context shape aligned with your rules:

```cel
user.country == "US" && targetingKey == "u_123"
```

For rollout bucketing by OpenFeature targeting key, set `bucket_by` to
`targetingKey`. If your rules use context schemas, include the top-level fields
used by CEL rules, such as `targetingKey`, `user`, or `request`.

## Runtime

The package targets Node.js server runtimes with a global `fetch`
implementation. Each evaluation sends `POST /eval/{flagKey}` to the configured
API endpoint with a JSON body containing `{ "context": ... }`.

Each request sends `Authorization: Bearer <apiKey>` when an API key is
configured. HTTP and network failures return the OpenFeature default value with
an error resolution.
