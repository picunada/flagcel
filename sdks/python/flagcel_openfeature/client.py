from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Protocol
from urllib.error import HTTPError, URLError
from urllib.parse import urlsplit
from urllib.request import Request, urlopen


class URLopener(Protocol):
    def __call__(self, request: Request, timeout: float | None = None) -> Any:
        ...


class DefinitionsClientError(RuntimeError):
    def __init__(self, message: str, status: int | None = None) -> None:
        super().__init__(message)
        self.status = status


@dataclass(frozen=True)
class FetchResult:
    definitions: dict[str, Any] | None = None
    etag: str = ""
    unchanged: bool = False


class DefinitionsClient:
    def __init__(
        self,
        endpoint: str,
        api_key: str = "",
        opener: URLopener | None = None,
        timeout: float | None = None,
    ) -> None:
        endpoint = endpoint.strip()
        parsed = urlsplit(endpoint)
        if not endpoint:
            raise ValueError("flagcel: endpoint is required")
        if not parsed.scheme or not parsed.netloc:
            raise ValueError("flagcel: endpoint must be an absolute URL")

        self._endpoint = endpoint.rstrip("/")
        self._api_key = api_key
        self._opener = opener or urlopen
        self._timeout = timeout

    def fetch_definitions(self, etag: str = "") -> FetchResult:
        request = Request(
            f"{self._endpoint}/eval/definitions",
            method="GET",
            headers={"Accept": "application/json"},
        )
        if self._api_key:
            request.add_header("Authorization", f"Bearer {self._api_key}")
        if etag:
            request.add_header("If-None-Match", etag)

        try:
            response = self._opener(request, timeout=self._timeout)
            try:
                status = response.status if hasattr(response, "status") else response.getcode()
                return self._decode_response(response, status, etag)
            finally:
                close = getattr(response, "close", None)
                if callable(close):
                    close()
        except HTTPError as exc:
            if exc.code == 304:
                return FetchResult(etag=_response_etag(exc, etag), unchanged=True)
            body_text = exc.read(512).decode("utf-8", errors="replace").strip()
            raise DefinitionsClientError(_error_message(body_text, exc.code), exc.code) from exc
        except URLError as exc:
            raise DefinitionsClientError(f"flagcel: fetch definitions: {exc.reason}") from exc

    def _decode_response(self, response: Any, status: int, previous_etag: str) -> FetchResult:
        if status == 304:
            return FetchResult(etag=_response_etag(response, previous_etag), unchanged=True)

        body_text = response.read().decode("utf-8", errors="replace")
        if status < 200 or status >= 300:
            raise DefinitionsClientError(_error_message(body_text, status), status)

        envelope = json.loads(body_text)
        return FetchResult(
            definitions=envelope["data"],
            etag=_response_etag(response, previous_etag),
        )


def _response_etag(response: Any, previous: str) -> str:
    headers = getattr(response, "headers", None)
    if headers is not None:
        etag = headers.get("ETag")
        if etag:
            return etag
    return previous


def _error_message(body: str, status: int) -> str:
    fallback = f"flagcel: fetch definitions: status {status}"
    if not body.strip():
        return fallback
    try:
        error = json.loads(body).get("error", {})
        return error.get("message") or error.get("code") or f"{fallback}: {body.strip()}"
    except json.JSONDecodeError:
        return f"{fallback}: {body.strip()}"
