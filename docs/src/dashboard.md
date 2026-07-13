<div class="fc-breadcrumb"><span>Operate</span><span>/</span><strong>Dashboard</strong></div>

# Dashboard

<p class="fc-lead">Use the embedded Flagcel dashboard to define contexts, ship targeting rules, and create environment-scoped evaluation keys without calling the admin API by hand.</p>

<div class="fc-tags">
  <span>Web UI</span>
  <span>Contexts</span>
  <span>Flags</span>
  <span>API keys</span>
</div>

The dashboard is served from `/` on the Flagcel binary. Local Docker Compose seeds
`admin@localhost` / `secret`. For build and local frontend development details, see
[Web UI](web-ui.md). For the underlying model, see [Concepts](concepts.md).

## Orientation

The sidebar is the main map of the product:

| Nav item | What it manages |
| --- | --- |
| **Flags** | Flags and rules for the selected environment |
| **Contexts** | Reusable evaluation schemas, shared across environments |
| **Environments** | Isolated flag sets such as `production` or `staging` |
| **API keys** | Bearer tokens for SDK and `/eval` traffic |
| **API docs** | OpenAPI / Swagger UI at `/docs` |

Use the environment selector near the top of the sidebar before editing flags or
keys. Contexts are global; flags, rules, and API keys are environment-scoped.

Recommended first workflow:

1. Create a context
2. Create a flag and attach that context
3. Add CEL rules
4. Try the evaluation playground
5. Create an API key for your app

## Create a context

Open **Contexts**, then **new context**.

1. Give the context a unique name such as `web-user`.
2. Optionally paste a representative payload under **paste payload JSON**.
3. Review the **inferred fields** list. Nested objects become dotted paths like
   `user.plan`.
4. Click **create context**.

You can also create an empty context and edit the schema later as one field per
line:

```text
user.id        string
user.plan      string
request.path   string
targetingKey   string
```

Supported types are `string`, `int`, `double`, `bool`, `timestamp`, `list`, and
`map`.

Guardrails you will hit in the UI:

- Schema edits that would break attached flags are rejected.
- A context still referenced by any flag cannot be deleted.
- Fields used by existing rules show warnings such as
  `user.plan is referenced by N rules; restore it before saving`.

See [Contexts](concepts.md#contexts) for how schemas differ from the runtime
evaluation payload.

## Create and attach a flag

Open **Flags**, select the target environment, then **new flag**.

1. Set a stable **key** such as `checkout-v2`.
2. Optionally choose a **context**.
3. Pick the flag **type**, **enabled** state, and **default value**.
4. Click **create flag**.

Rules are added after creation. On the flag detail page you can still change the
attached context from the flag settings panel. Attaching a context unlocks typed
CEL validation and autocomplete while editing rules.

Without a context, evaluation still works, but Flagcel cannot verify that rule
paths exist or match declared types.

## Add targeting rules

On a flag detail page, rules are listed under **targeting · evaluated
top-to-bottom**.

1. Click **add rule**.
2. Write a CEL expression such as
   `user.plan == "pro" && request.path.startsWith("/checkout")`.
3. Optionally configure a percentage rollout and `bucket_by` path.
4. Set the rule value and save.

Rules run in order. The first match wins, so keep narrow or high-priority rules
above broader ones. Use the move controls to reorder without rewriting
expressions.

If the flag has a context attached, invalid paths and type mismatches surface
while editing instead of only at evaluation time.

## Try the evaluation playground

Open a flag and use **[ evaluation playground ]**.

1. Paste or edit a JSON evaluation payload.
2. Click **evaluate**.
3. Inspect the returned value, matched rule, and trace details.

When a context is attached, the playground can start from a sample payload for
that schema. This is the fastest way to verify a rule before wiring an SDK.

The playground uses the admin session and the selected environment. Production
apps should evaluate with environment-scoped API keys instead.

## Create an API key

Open **API keys**, then **new key**.

1. Name the key, for example `checkout-service`.
2. Choose the environment the key can evaluate.
3. Click **create**.
4. Copy the raw token immediately. Flagcel shows it once.

Store the token as an application secret and pass it to an OpenFeature provider
or `/eval` caller. Revoked keys stop evaluating immediately. Key prefixes are
safe to share; raw tokens are not.

See [Authentication](auth.md) and [SDKs](sdks/) for how evaluation keys fit
into local auth, OIDC, and provider setup.

## Manage environments

Open **Environments** to create additional isolated flag sets.

- Each environment has its own flags, rules, and API keys.
- Contexts remain shared so the same payload schema can be reused in staging and
  production.
- An environment with flags, rules, or API keys cannot be deleted until those
  dependents are removed.

Use separate keys per environment so staging traffic never evaluates production
rules.

## Practical tips

- Define the context before writing complex CEL. Autocomplete and validation are
  much clearer with a schema attached.
- Keep `bucket_by` on a stable identifier such as `user.id` or `targetingKey`.
- Prefer one context per client shape, not one context per flag, when several
  flags share the same payload.
- Use the playground to confirm match order before rolling a percentage out.

## Related docs

- [Concepts](concepts.md) for environments, contexts, rules, and rollouts
- [Quickstart](quickstart.md) for the shortest path to a first evaluation
- [Web UI](web-ui.md) for local frontend development and embedding
- [API](api.md) for the admin and evaluation endpoints behind the dashboard
