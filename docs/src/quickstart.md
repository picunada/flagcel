<div class="fc-breadcrumb"><span>Get Started</span><span>/</span><strong>Quickstart</strong></div>

# Quickstart

<p class="fc-lead">Get a Flagcel control plane running and evaluate your first CEL-targeted flag in under five minutes. You'll need Docker and a Postgres database.</p>

<div class="fc-tags">
  <span>≈ 5 min</span>
  <span>Docker</span>
  <span>Postgres 14+</span>
</div>

## Install

The production binary ships everything - dashboard, API, and embedded migrations - in a single image.

<div class="fc-process-block" aria-hidden="true"></div>

```bash
# pull and run the control plane
$ docker run --rm -p 8080:8080 \
    -e DATABASE_URL="postgres://user:pass@db.example.com:5432/flagcel?sslmode=require" \
    -e AUTH_SESSION_SECRET="replace-with-at-least-32-random-bytes" \
    -e AUTH_BOOTSTRAP_ADMIN_EMAIL="admin@example.com" \
    -e AUTH_BOOTSTRAP_ADMIN_PASSWORD="change-me" \
    ghcr.io/picunada/flagcel:latest

# control plane up on :8080 - migrations run on boot
→ dashboard  http://localhost:8080
→ api        http://localhost:8080/api/v1
→ docs       http://localhost:8080/docs
```

## Run the control plane

Point `DATABASE_URL` at your database. On boot, Flagcel applies any pending migrations, then serves the dashboard and versioned API on the same port.

::: tip Local development
Clone the repository and run `docker compose up` to start Flagcel with Postgres and a seeded local admin account - no external database required. Sign in at `http://localhost:8080` with `admin@localhost` / `secret`.
:::

## Create your first flag

Open the dashboard, choose **New flag**, and create a boolean flag with these values:

- Key: `checkout-v2`
- Default value: `false`
- Rule value: `true`
- Rule expression:

```text [CEL]
user.email.endsWith("@flagcel.io") ||
  (geo.region == "EU" && user.id % 10 == 0)
```

Targeting uses the [Common Expression Language](concepts.md) - the same expression language used by Kubernetes and Envoy. This rule enables the flag for internal users and ten percent of EU traffic when `user.id` is numeric.

## Evaluate from your app

Create an evaluation key from the dashboard's **API keys** page. Keys are scoped per environment, so production traffic never sees staging rules.

Install the server-side OpenFeature provider:

```bash
npm install @openfeature/server-sdk @flagcel/openfeature-server
```

Then configure it with the API URL and evaluation key:

```ts [server.ts]
import { OpenFeature } from "@openfeature/server-sdk";
import { FlagcelProvider } from "@flagcel/openfeature-server";

await OpenFeature.setProviderAndWait(new FlagcelProvider({
  endpoint: "http://localhost:8080/api/v1",
  apiKey: process.env.FLAGCEL_API_KEY,
}));

const client = OpenFeature.getClient("checkout-service");

const enabled = await client.getBooleanValue("checkout-v2", false, {
  targetingKey: String(user.id),
  user: { id: user.id, email: user.email },
  geo: { region: geo.region },
});
```

## Next steps

<div class="fc-nextgrid">
  <a class="fc-ncard" href="dashboard">
    <div class="nk">Operate</div>
    <div class="nh">Dashboard workflow</div>
    <div class="nd">Create contexts, attach flags, write rules, and mint API keys in the UI.</div>
  </a>
  <a class="fc-ncard" href="concepts">
    <div class="nk">Concepts</div>
    <div class="nh">CEL targeting</div>
    <div class="nd">Understand contexts, rules, evaluation payloads, and rollout bucketing.</div>
  </a>
  <a class="fc-ncard" href="auth">
    <div class="nk">Operate</div>
    <div class="nh">Authentication</div>
    <div class="nd">Prepare local auth or OIDC before exposing a deployment.</div>
  </a>
  <a class="fc-ncard" href="sdks/">
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
