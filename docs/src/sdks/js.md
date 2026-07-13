<div class="fc-breadcrumb"><span>SDKs</span><span>/</span><strong>JS/TS</strong></div>

# JS/TS

[![npm](https://img.shields.io/npm/v/@flagcel/openfeature-server.svg)](https://www.npmjs.com/package/@flagcel/openfeature-server)

The JS/TS SDK is an OpenFeature server provider in
[`sdks/js`](https://github.com/picunada/flagcel/tree/main/sdks/js). It packages
`@flagcel/openfeature-server` and resolves flags by calling the Flagcel
evaluation API from Node server runtimes.

## Install

```sh
npm install @flagcel/openfeature-server @openfeature/server-sdk
```

## Usage

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

## Behavior

- Targets Node.js server runtimes with a global `fetch`
- Sends `POST /eval/{flagKey}` with `{ "context": ... }`
- Uses `Authorization: Bearer <apiKey>`; the key selects the environment
- Returns the OpenFeature default value on HTTP or network failures

Keep evaluation context aligned with your
[context schemas](../concepts.md#contexts) and CEL rules. For rollout bucketing
by OpenFeature targeting key, set `bucket_by` to `targetingKey`.

See the
[JS/TS SDK README](https://github.com/picunada/flagcel/tree/main/sdks/js#readme)
for runtime requirements, targeting notes, and failure behavior.
