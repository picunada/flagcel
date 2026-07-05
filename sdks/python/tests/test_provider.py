from __future__ import annotations

import json
import threading
from email.message import Message
from typing import Any
from urllib.error import HTTPError, URLError

import pytest
from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import Reason

from flagcel_openfeature import FlagcelProvider


def test_provider_fetches_definitions_and_evaluates_locally() -> None:
    calls: list[tuple[str, dict[str, str], bytes | None]] = []

    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del timeout
        calls.append((request.full_url, dict(request.header_items()), request.data))
        return definitions_response(
            definitions_fixture(
                "checkout-copy",
                "string",
                "control",
                [
                    {
                        "id": "pro",
                        "expression": 'user.tier == "pro"',
                        "rollout": {"percentage": 100},
                        "value": "pro-copy",
                    }
                ],
            ),
            etag='"v1"',
        )

    provider = FlagcelProvider(
        endpoint="https://flagcel.test/api/v1",
        api_key="secret",
        http_client=opener,
        poll_interval=3600,
    )
    provider.initialize()
    try:
        detail = provider.resolve_string_details(
            "checkout-copy",
            "fallback",
            EvaluationContext(attributes={"user": {"tier": "pro"}}),
        )
    finally:
        provider.shutdown()

    assert detail.value == "pro-copy"
    assert detail.reason == Reason.TARGETING_MATCH
    assert detail.variant == "pro"
    assert detail.flag_metadata == {
        "flagcelReason": "matched_rule",
        "valueType": "string",
    }
    definition_calls = [call for call in calls if call[0].endswith("/eval/definitions")]
    assert definition_calls == [
        (
            "https://flagcel.test/api/v1/eval/definitions",
            {"Accept": "application/json", "Authorization": "Bearer secret"},
            None,
        )
    ]


def test_provider_honors_etag_and_keeps_last_known_definitions() -> None:
    requests = 0

    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        nonlocal requests
        del timeout
        requests += 1
        if requests == 1:
            return definitions_response(
                definitions_fixture("enabled", "boolean", True, []),
                etag='"v1"',
            )
        if requests == 2:
            assert dict(request.header_items())["If-none-match"] == '"v1"'
            return FakeResponse(304, b"", {"ETag": '"v1"'})
        raise HTTPError(
            request.full_url,
            503,
            "Service Unavailable",
            {},
            FakeBody(b"temporarily unavailable"),
        )

    provider = FlagcelProvider("https://flagcel.test/api/v1", http_client=opener, poll_interval=3600)
    provider.initialize()
    try:
        provider._refresh()
        with pytest.raises(Exception):
            provider._refresh()

        detail = provider.resolve_boolean_details("enabled", False, EvaluationContext())
    finally:
        provider.shutdown()

    assert detail.value is True
    assert detail.error_code is None
    assert requests == 3


def test_provider_returns_default_when_initial_fetch_fails() -> None:
    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del request, timeout
        raise URLError("connection refused")

    provider = FlagcelProvider("https://flagcel.test/api/v1", http_client=opener, poll_interval=3600)
    provider.initialize()
    try:
        detail = provider.resolve_boolean_details("enabled", False, EvaluationContext())
    finally:
        provider.shutdown()

    assert detail.value is False
    assert detail.reason == Reason.ERROR
    assert detail.error_code == ErrorCode.PROVIDER_NOT_READY
    assert "connection refused" in str(detail.error_message)


def test_provider_resolves_string_number_boolean_and_object_values() -> None:
    provider = initialized_provider(
        {
            "flags": [
                flag_definition("title", "string", "hello", []),
                flag_definition("count", "number", 42, []),
                flag_definition("enabled", "boolean", True, []),
                flag_definition("payload", "json", {"plan": "pro"}, []),
            ]
        }
    )
    try:
        title = provider.resolve_string_details("title", "", EvaluationContext())
        count = provider.resolve_float_details("count", 0.0, EvaluationContext())
        enabled = provider.resolve_boolean_details("enabled", False, EvaluationContext())
        payload = provider.resolve_object_details("payload", {}, EvaluationContext())
    finally:
        provider.shutdown()

    assert title.value == "hello"
    assert title.reason == Reason.DEFAULT
    assert count.value == 42.0
    assert count.reason == Reason.DEFAULT
    assert enabled.value is True
    assert payload.value == {"plan": "pro"}


def test_provider_returns_default_on_type_mismatch() -> None:
    provider = initialized_provider({"flags": [flag_definition("title", "string", "hello", [])]})
    try:
        detail = provider.resolve_boolean_details("title", True, EvaluationContext())
    finally:
        provider.shutdown()

    assert detail.value is True
    assert detail.reason == Reason.ERROR
    assert detail.error_code == ErrorCode.TYPE_MISMATCH


def test_integer_resolution_rejects_non_integral_float() -> None:
    provider = initialized_provider({"flags": [flag_definition("ratio", "number", 1.5, [])]})
    try:
        detail = provider.resolve_integer_details("ratio", 0, EvaluationContext())
    finally:
        provider.shutdown()

    assert detail.value == 0
    assert detail.error_code == ErrorCode.TYPE_MISMATCH


def test_provider_returns_default_on_not_found() -> None:
    provider = initialized_provider({"flags": [flag_definition("enabled", "boolean", True, [])]})
    try:
        detail = provider.resolve_boolean_details("missing", True, EvaluationContext())
    finally:
        provider.shutdown()

    assert detail.value is True
    assert detail.reason == Reason.ERROR
    assert detail.error_code == ErrorCode.FLAG_NOT_FOUND
    assert detail.error_message == "flag not found"


def test_disabled_flag_returns_default_with_disabled_reason() -> None:
    definitions = {
        "flags": [
            {
                **flag_definition("enabled", "boolean", True, []),
                "enabled": False,
            }
        ]
    }
    provider = initialized_provider(definitions)
    try:
        detail = provider.resolve_boolean_details("enabled", False, EvaluationContext())
    finally:
        provider.shutdown()

    assert detail.value is True
    assert detail.reason == Reason.DISABLED


def test_rollout_bucket_excludes_context_from_rule_value() -> None:
    provider = initialized_provider(
        definitions_fixture(
            "experiment",
            "string",
            "control",
            [
                {
                    "id": "variant",
                    "expression": 'user.tier == "pro"',
                    "rollout": {"percentage": 0, "bucket_by": "user.id"},
                    "value": "variant",
                }
            ],
        )
    )
    try:
        detail = provider.resolve_string_details(
            "experiment",
            "fallback",
            EvaluationContext(attributes={"user": {"tier": "pro", "id": "u_123"}}),
        )
    finally:
        provider.shutdown()

    assert detail.value == "control"
    assert detail.reason == Reason.TARGETING_MATCH
    assert detail.variant == "variant"


def initialized_provider(definitions: dict[str, Any]) -> FlagcelProvider:
    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del request, timeout
        return definitions_response(definitions)

    provider = FlagcelProvider("https://flagcel.test/api/v1", http_client=opener, poll_interval=3600)
    provider.initialize()
    return provider


def definitions_response(definitions: dict[str, Any], etag: str = '"v1"') -> "FakeResponse":
    return FakeResponse(
        200,
        json.dumps({"message": "success", "data": definitions}).encode("utf-8"),
        {"Content-Type": "application/json", "ETag": etag},
    )


def definitions_fixture(
    key: str,
    value_type: str,
    default_value: Any,
    rules: list[dict[str, Any]],
) -> dict[str, Any]:
    context_id = "user-context"
    return {
        "flags": [
            flag_definition(
                key,
                value_type,
                default_value,
                rules,
                context_id=context_id,
            )
        ],
        "contexts": [
            {
                "id": context_id,
                "name": "User",
                "fields": [
                    {"path": "user.tier", "type": "string"},
                    {"path": "user.id", "type": "string"},
                ],
            }
        ],
    }


def flag_definition(
    key: str,
    value_type: str,
    default_value: Any,
    rules: list[dict[str, Any]],
    context_id: str | None = None,
) -> dict[str, Any]:
    definition = {
        "key": key,
        "type": value_type,
        "enabled": True,
        "default_value": default_value,
        "rules": rules,
    }
    if context_id is not None:
        definition["context_id"] = context_id
    return definition


class FakeResponse:
    def __init__(self, status: int, body: bytes, headers: dict[str, str]) -> None:
        self.status = status
        self._body = body
        self.headers = Message()
        for key, value in headers.items():
            self.headers[key] = value

    def getcode(self) -> int:
        return self.status

    def read(self, limit: int | None = None) -> bytes:
        if limit is None:
            return self._body
        return self._body[:limit]

    def close(self) -> None:
        pass


class FakeBody:
    def __init__(self, body: bytes) -> None:
        self._body = body
        self._lock = threading.Lock()

    def read(self, limit: int | None = None) -> bytes:
        with self._lock:
            if limit is None:
                return self._body
            return self._body[:limit]

    def close(self) -> None:
        pass
