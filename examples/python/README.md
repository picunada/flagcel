# Flagcel Python SDK Example

Run this example against a local Flagcel server:

```sh
uv sync
uv run python main.py
```

For a JSON flag like `user-variant` with rules such as
`user.role == "example"`, run:

```sh
export FLAGCEL_API_KEY=fc_your_api_key
FLAGCEL_FLAG_KEY=user-variant USER_ROLE=example uv run python main.py
```

Fish shell:

```fish
set -x FLAGCEL_API_KEY fc_your_api_key
set -x FLAGCEL_FLAG_KEY user-variant
set -x USER_ROLE example
uv run python main.py
```

## Load Test

The load test exercises the Python SDK's local evaluation path. It performs the
provider's definition fetch during startup, then measures OpenFeature provider
overhead plus local CEL evaluation latency.

```fish
set -x FLAGCEL_API_KEY fc_your_api_key
set -x FLAGCEL_FLAG_KEY user-variant
set -x USER_ROLE example
set -x LOAD_CONCURRENCY 16
set -x LOAD_DURATION_SECONDS 30
uv run python load_test.py
```

Useful knobs:

| Variable | Default | Description |
| --- | --- | --- |
| `LOAD_PROCESSES` | `1` | OS processes to run in parallel for CPU scaling |
| `LOAD_CONCURRENCY` | `8` | Worker threads evaluating flags |
| `LOAD_DURATION_SECONDS` | `30` | Timed run duration |
| `LOAD_WARMUP_REQUESTS` | `100` | Evaluations before measuring |
| `LOAD_UNIQUE_TARGETS` | `true` | Vary `user.id` and `targetingKey` per evaluation |
| `LOAD_PROFILE_STEPS` | `false` | Print context, OpenFeature, and result timing breakdowns |
| `LOAD_SLOW_THRESHOLD_US` | `1000` | Count and capture slow evaluations at or above this threshold; set `0` to disable |
| `LOAD_SLOW_EVENT_LIMIT` | `20` | Slowest events retained per worker |
| `EXPECTED_VARIANT` | `USER_ROLE` | Counts a mismatch when `value.variant` differs |

Set `LOAD_UNIQUE_TARGETS=false` to reuse one evaluation context per worker.

Use processes, not only threads, when measuring CPU-parallel Python throughput:

```fish
set -x LOAD_PROCESSES 8
set -x LOAD_CONCURRENCY 1
uv run python load_test.py
```

To profile where time is going:

```fish
set -x LOAD_PROFILE_STEPS true
set -x LOAD_SLOW_THRESHOLD_US 1000
uv run python load_test.py
```

To investigate scheduler/CPU-saturation tails, compare process counts:

```fish
set -x LOAD_PROFILE_STEPS true
set -x LOAD_SLOW_THRESHOLD_US 1000
set -x LOAD_CONCURRENCY 1
set -x LOAD_PROCESSES 4
uv run python load_test.py
```

The local SDK dependency is declared in `pyproject.toml` and resolved from
`../../sdks/python`.

For IDEs using `ty`, the repo root has a `ty.toml` that points analysis at this
example's `.venv` and the local SDK source tree. Run `uv sync` first so
`openfeature-sdk` and `cel-expr-python` are installed in `examples/python/.venv`.
