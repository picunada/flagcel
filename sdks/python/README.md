# Flagcel OpenFeature Provider for Python

OpenFeature provider for Flagcel with server-side evaluation. The provider
resolves flags by calling the Flagcel evaluation API.

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

- Sends `POST /eval/{flag_key}` relative to the configured endpoint.
- Sends JSON bodies containing `{ "context": ... }`.
- Sends `Authorization: Bearer <api_key>` when an API key is configured.
- Returns OpenFeature defaults with error details on HTTP or network failures.
- Maps Flagcel reason, variant, and value type into OpenFeature resolution details.
