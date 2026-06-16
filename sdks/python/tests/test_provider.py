from __future__ import annotations

import json
import threading
from email.message import Message
from typing import Any
from urllib.error import HTTPError, URLError

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import Reason

from flagcel_openfeature import FlagcelProvider


def test_provider_sends_authenticated_server_side_evaluation_request() -> None:
    calls: list[tuple[str, dict[str, str], dict[str, Any]]] = []

    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del timeout
        body = json.loads(request.data.decode("utf-8"))
        calls.append((request.full_url, dict(request.header_items()), body))
        return json_response(
            {
                "key": "enabled",
                "value": True,
                "value_type": "boolean",
                "reason": "matched_rule",
                "variant": "targeted",
            }
        )

    provider = FlagcelProvider(
        endpoint="https://flagcel.test/api/v1",
        api_key="secret",
        http_client=opener,
    )
    provider.initialize()

    detail = provider.resolve_boolean_details(
        "enabled",
        False,
        EvaluationContext(targeting_key="user-123"),
    )

    assert detail.value is True
    assert detail.reason == Reason.TARGETING_MATCH
    assert detail.variant == "targeted"
    assert calls == [
        (
            "https://flagcel.test/api/v1/eval/enabled",
            {
                "Accept": "application/json",
                "Content-type": "application/json",
                "Authorization": "Bearer secret",
            },
            {"context": {"targetingKey": "user-123"}},
        )
    ]


def test_provider_resolves_string_number_and_object_values() -> None:
    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del timeout
        key = request.full_url.rsplit("/", 1)[-1]
        if key == "title":
            return json_response({"key": key, "value": "hello", "value_type": "string", "reason": "default_no_match"})
        if key == "count":
            return json_response({"key": key, "value": 42, "value_type": "number", "reason": "default_no_match"})
        return json_response({"key": key, "value": {"plan": "pro"}, "value_type": "json", "reason": "default_no_match"})

    provider = FlagcelProvider("https://flagcel.test/api/v1", http_client=opener)

    title = provider.resolve_string_details("title", "", EvaluationContext())
    count = provider.resolve_float_details("count", 0.0, EvaluationContext())
    payload = provider.resolve_object_details("payload", {}, EvaluationContext())

    assert title.value == "hello"
    assert title.reason == Reason.DEFAULT
    assert count.value == 42.0
    assert count.reason == Reason.DEFAULT
    assert payload.value == {"plan": "pro"}
    assert payload.reason == Reason.DEFAULT


def test_provider_returns_default_on_type_mismatch() -> None:
    provider = initialized_provider({"key": "title", "value": "hello", "value_type": "string"})

    detail = provider.resolve_boolean_details("title", True, EvaluationContext())

    assert detail.value is True
    assert detail.reason == Reason.ERROR
    assert detail.error_code == ErrorCode.TYPE_MISMATCH


def test_integer_resolution_rejects_non_integral_float() -> None:
    provider = initialized_provider({"key": "ratio", "value": 1.5, "value_type": "number"})

    detail = provider.resolve_integer_details("ratio", 0, EvaluationContext())

    assert detail.value == 0
    assert detail.error_code == ErrorCode.TYPE_MISMATCH


def test_provider_returns_default_on_not_found() -> None:
    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del timeout
        raise HTTPError(
            request.full_url,
            404,
            "Not Found",
            {},
            FakeBody(json.dumps({"error": {"code": "FLAG_NOT_FOUND", "message": "Flag not found"}}).encode("utf-8")),
        )

    provider = FlagcelProvider("https://flagcel.test/api/v1", http_client=opener)

    detail = provider.resolve_boolean_details("missing", True, EvaluationContext())

    assert detail.value is True
    assert detail.reason == Reason.ERROR
    assert detail.error_code == ErrorCode.FLAG_NOT_FOUND
    assert detail.error_message == "Flag not found"


def test_provider_returns_default_on_network_failure() -> None:
    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del request, timeout
        raise URLError("connection refused")

    provider = FlagcelProvider("https://flagcel.test/api/v1", http_client=opener)

    detail = provider.resolve_boolean_details("enabled", False, EvaluationContext())

    assert detail.value is False
    assert detail.reason == Reason.ERROR
    assert detail.error_code == ErrorCode.GENERAL
    assert "connection refused" in str(detail.error_message)


def initialized_provider(result: dict[str, Any]) -> FlagcelProvider:
    def opener(request: Any, timeout: float | None = None) -> FakeResponse:
        del request, timeout
        return json_response(result)

    return FlagcelProvider("https://flagcel.test/api/v1", http_client=opener)


def json_response(result: dict[str, Any]) -> "FakeResponse":
    return FakeResponse(
        200,
        json.dumps({"message": "success", "data": result}).encode("utf-8"),
        {"Content-Type": "application/json"},
    )


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
