<div class="fc-breadcrumb"><span>SDKs</span><span>/</span><strong>Overview</strong></div>

# SDKs

Flagcel SDKs are OpenFeature providers. Use an environment-scoped evaluation API
key from the dashboard, then evaluate flags from your service.

| SDK | Package | Evaluation model |
| --- | --- | --- |
| [Go](go.md) | `github.com/picunada/flagcel/sdks/go` | Polls definitions, evaluates locally with `cel-go` |
| [JS/TS](js.md) | `@flagcel/openfeature-server` | Calls Flagcel evaluation API from Node |
| [Python](python.md) | `flagcel-openfeature` | Polls definitions, evaluates locally with `cel-expr-python` |

Runnable local examples for all three SDKs live in the
[examples directory](https://github.com/picunada/flagcel/tree/main/examples).

## Choose a provider

- Prefer **Go** or **Python** when you want in-process CEL evaluation and lower
  per-request latency after the definitions cache warms.
- Prefer **JS/TS** for Node server runtimes that should resolve through the
  Flagcel evaluation API.

All providers accept OpenFeature evaluation context. Keep that payload aligned
with your [context schemas](../concepts.md#contexts) and CEL rules.

## Next steps

<div class="fc-nextgrid">
  <a class="fc-ncard" href="go">
    <div class="nk">Go</div>
    <div class="nh">Local CEL evaluation</div>
    <div class="nd">Poll definitions and evaluate flags in-process with OpenFeature.</div>
  </a>
  <a class="fc-ncard" href="js">
    <div class="nk">JS/TS</div>
    <div class="nh">Server-side provider</div>
    <div class="nd">Resolve flags from Node through the Flagcel evaluation API.</div>
  </a>
  <a class="fc-ncard" href="python">
    <div class="nk">Python</div>
    <div class="nh">Local CEL evaluation</div>
    <div class="nd">Poll definitions and evaluate flags in-process with OpenFeature.</div>
  </a>
  <a class="fc-ncard" href="../dashboard">
    <div class="nk">Operate</div>
    <div class="nh">Create an API key</div>
    <div class="nd">Mint an environment-scoped evaluation key in the dashboard.</div>
  </a>
</div>
