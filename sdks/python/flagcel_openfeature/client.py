from __future__ import annotations

import json
from typing import Any, Protocol
from urllib.error import HTTPError, URLError
from urllib.parse import quote, urlsplit
from urllib.request import Request, urlopen


class URLopener(Protocol):
    def __call__(self, request: Request, timeout: float | None = None) -> Any:
        ...


class EvalClientError(RuntimeError):
    def __init__(self, message: str, status: int | None = None) -> None:
        super().__init__(message)
        self.status = status


class EvalClient:
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

    def evaluate_json(self, flag_key: str, context_json: str) -> dict[str, Any]:
        body = json.dumps(
            {"context": json.loads(context_json)},
            separators=(",", ":"),
        ).encode("utf-8")
        request = Request(
            self._eval_url(flag_key),
            data=body,
            method="POST",
            headers={
                "Accept": "application/json",
                "Content-Type": "application/json",
            },
        )
        if self._api_key:
            request.add_header("Authorization", f"Bearer {self._api_key}")

        try:
            response = self._opener(request, timeout=self._timeout)
            try:
                status = response.status if hasattr(response, "status") else response.getcode()
                return self._decode_response(response, status)
            finally:
                close = getattr(response, "close", None)
                if callable(close):
                    close()
        except HTTPError as exc:
            body_text = exc.read(512).decode("utf-8", errors="replace").strip()
            raise EvalClientError(_error_message(body_text, exc.code), exc.code) from exc
        except URLError as exc:
            raise EvalClientError(f"flagcel: evaluate flag: {exc.reason}") from exc

    def _decode_response(self, response: Any, status: int) -> dict[str, Any]:
        body_text = response.read().decode("utf-8", errors="replace")
        if status < 200 or status >= 300:
            raise EvalClientError(_error_message(body_text, status), status)

        envelope = json.loads(body_text)
        return envelope["data"]

    def _eval_url(self, flag_key: str) -> str:
        return f"{self._endpoint}/eval/{quote(flag_key, safe='')}"


def _error_message(body: str, status: int) -> str:
    fallback = f"flagcel: evaluate flag: status {status}"
    if not body.strip():
        return fallback
    try:
        error = json.loads(body).get("error", {})
        return error.get("message") or error.get("code") or f"{fallback}: {body.strip()}"
    except json.JSONDecodeError:
        return f"{fallback}: {body.strip()}"
