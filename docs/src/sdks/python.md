<div class="fc-breadcrumb"><span>SDKs</span><span>/</span><strong>Python</strong></div>

# Python

[![PyPI](https://img.shields.io/pypi/v/flagcel-openfeature.svg)](https://pypi.org/project/flagcel-openfeature/)

The Python SDK is an OpenFeature provider in
[`sdks/python`](https://github.com/picunada/flagcel/tree/main/sdks/python). It
packages `flagcel-openfeature`, polls `GET /api/v1/eval/definitions` with an
evaluation API key, compiles definitions with `cel-expr-python`, and evaluates
flags locally.

## Install

```sh
python -m pip install flagcel-openfeature
```

Requires Python 3.11+.

## Usage

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

## Behavior

- Polls definitions every 30 seconds by default
- Uses `ETag` / `If-None-Match` to skip unchanged payloads
- Evaluates flags locally from the latest definitions
- Keeps last-known definitions after later refresh failures
- Returns OpenFeature defaults if no definitions have loaded yet

Configure the polling interval in seconds:

```python
provider = FlagcelProvider(
	endpoint="http://localhost:8080/api/v1",
	api_key=api_key,
	poll_interval=10.0,
)
```

Keep evaluation context aligned with your
[context schemas](../concepts.md#contexts) and CEL rules.

See the
[Python SDK README](https://github.com/picunada/flagcel/tree/main/sdks/python#readme)
for installation, targeting context, polling, and fail-open details.
