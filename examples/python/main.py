from __future__ import annotations

import os

from flagcel_openfeature import FlagcelProvider
from openfeature import api
from openfeature.evaluation_context import EvaluationContext


def main() -> None:
    endpoint = env("FLAGCEL_ENDPOINT", "http://localhost:8080/api/v1")
    api_key = os.getenv("FLAGCEL_API_KEY", "")
    flag_key = env("FLAGCEL_FLAG_KEY", "new-checkout")
    targeting_key = env("TARGETING_KEY", "u_123")
    user_role = env("USER_ROLE", "example")

    provider = FlagcelProvider(
        endpoint=endpoint,
        api_key=api_key,
    )
    provider.initialize()
    api.set_provider(provider)

    try:
        client = api.get_client("flagcel-local-example")
        details = client.get_boolean_details(
            flag_key,
            False,
            EvaluationContext(
                targeting_key=targeting_key,
                attributes={
                    "user": {
                        "id": targeting_key,
                        "role": user_role,
                        "country": env("USER_COUNTRY", "US"),
                    },
                    "request": {
                        "path": env("REQUEST_PATH", "/checkout"),
                    },
                },
            ),
        )
        print(
            f"flag={flag_key} value={details.value} "
            f"reason={details.reason} variant={details.variant or ''}"
        )
    finally:
        provider.shutdown()


def env(key: str, fallback: str) -> str:
    return os.getenv(key) or fallback


if __name__ == "__main__":
    main()
