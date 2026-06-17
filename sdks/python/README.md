# Flagcel OpenFeature Provider for Python

[![PyPI](https://img.shields.io/pypi/v/flagcel-openfeature.svg)](https://pypi.org/project/flagcel-openfeature/)

OpenFeature provider for Flagcel with local CEL evaluation. The provider polls
Flagcel for evaluation definitions, compiles rules with `cel-expr-python`, and
evaluates flags in-process.

```sh
python -m pip install flagcel-openfeature
```

```python
from openfeature import api
from openfeature.evaluation_context import EvaluationContext
from flagcel_openfeature import FlagcelProvider

provider = FlagcelProvider(
    endpoint="http://localhost:8080/api/v1",
    api_key="fc_example_secret",
)
api.set_provider(provider)

client = api.get_client("checkout-service")
enabled = client.get_boolean_value(
    "new-checkout",
    False,
    EvaluationContext(
        targeting_key="u_123",
        attributes={"user": {"id": "u_123", "country": "US"}},
    ),
)
```

## Behavior

- Requires Python 3.11+.
- Fetches `GET /eval/definitions` relative to the configured endpoint.
- Sends `Authorization: Bearer <api_key>` when an API key is configured; the API key selects the evaluation environment.
- Polls definitions every 30 seconds by default.
- Uses `ETag` and `If-None-Match` to avoid reloading unchanged definitions.
- Evaluates flags locally with CEL expressions from the latest definitions.
- Returns OpenFeature defaults with provider-not-ready details if no definitions have loaded.
- Keeps using last-known definitions after later refresh failures.
- Maps Flagcel reason, variant, and value type into OpenFeature resolution details.

Configure the polling interval in seconds:

```python
provider = FlagcelProvider(
    endpoint="http://localhost:8080/api/v1",
    api_key="fc_example_secret",
    poll_interval=10.0,
)
```
