from __future__ import annotations

import json
import math
import threading
from collections.abc import Mapping, Sequence
from dataclasses import dataclass
from typing import Any

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import FlagResolutionDetails, Reason
from openfeature.hook import Hook
from openfeature.provider import AbstractProvider, Metadata

from .client import DefinitionsClient, DefinitionsClientError, URLopener
from .evaluator import (
    VALUE_TYPE_BOOLEAN,
    VALUE_TYPE_JSON,
    VALUE_TYPE_NUMBER,
    VALUE_TYPE_STRING,
    EvaluationResult,
    Evaluator,
)


PROVIDER_NAME = "flagcel"
DEFAULT_POLL_INTERVAL = 30.0


@dataclass(frozen=True)
class FlagcelProviderOptions:
    endpoint: str
    api_key: str = ""
    http_client: URLopener | None = None
    timeout: float | None = None
    poll_interval: float = DEFAULT_POLL_INTERVAL


class FlagcelProvider(AbstractProvider):
    def __init__(
        self,
        endpoint: str | FlagcelProviderOptions,
        api_key: str = "",
        http_client: URLopener | None = None,
        timeout: float | None = None,
        poll_interval: float = DEFAULT_POLL_INTERVAL,
    ) -> None:
        super().__init__()
        if isinstance(endpoint, FlagcelProviderOptions):
            options = endpoint
        else:
            options = FlagcelProviderOptions(
                endpoint=endpoint,
                api_key=api_key,
                http_client=http_client,
                timeout=timeout,
                poll_interval=poll_interval,
            )

        self._client = DefinitionsClient(
            options.endpoint,
            options.api_key,
            opener=options.http_client,
            timeout=options.timeout,
        )
        self._cache = _DefinitionsCache()
        self._poll_interval = options.poll_interval if options.poll_interval > 0 else DEFAULT_POLL_INTERVAL
        self._lifecycle_lock = threading.Lock()
        self._stop_event: threading.Event | None = None
        self._poll_thread: threading.Thread | None = None

    def get_metadata(self) -> Metadata:
        return Metadata(name=PROVIDER_NAME)

    def get_provider_hooks(self) -> list[Hook]:
        return []

    def initialize(self, evaluation_context: EvaluationContext | None = None) -> None:
        del evaluation_context
        with self._lifecycle_lock:
            if self._poll_thread is not None:
                return
            stop_event = threading.Event()
            self._stop_event = stop_event
            self._poll_thread = threading.Thread(
                target=self._poll,
                args=(stop_event,),
                name="flagcel-definitions-poll",
                daemon=True,
            )

        try:
            self._refresh()
        except Exception:
            pass

        self._poll_thread.start()

    def shutdown(self) -> None:
        with self._lifecycle_lock:
            stop_event = self._stop_event
            poll_thread = self._poll_thread
            self._stop_event = None
            self._poll_thread = None

        if stop_event is None or poll_thread is None:
            return
        stop_event.set()
        poll_thread.join()

    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[bool]:
        return self._resolve(flag_key, default_value, evaluation_context, VALUE_TYPE_BOOLEAN, bool)

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[str]:
        return self._resolve(flag_key, default_value, evaluation_context, VALUE_TYPE_STRING, str)

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[int]:
        detail = self._evaluate(flag_key, default_value, evaluation_context, VALUE_TYPE_NUMBER)
        if detail.error_code is not None:
            return detail
        value = detail.value
        if isinstance(value, bool) or not isinstance(value, (int, float)):
            return _type_mismatch(default_value, flag_key, "integer", type(value).__name__)
        if isinstance(value, float):
            if not math.isfinite(value) or math.trunc(value) != value:
                return _type_mismatch(default_value, flag_key, "integer", "number")
            value = int(value)
        return FlagResolutionDetails(
            value=int(value),
            reason=detail.reason,
            variant=detail.variant,
            flag_metadata=detail.flag_metadata,
        )

    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[float]:
        detail = self._evaluate(flag_key, default_value, evaluation_context, VALUE_TYPE_NUMBER)
        if detail.error_code is not None:
            return detail
        value = detail.value
        if isinstance(value, bool) or not isinstance(value, (int, float)):
            return _type_mismatch(default_value, flag_key, "number", type(value).__name__)
        return FlagResolutionDetails(
            value=float(value),
            reason=detail.reason,
            variant=detail.variant,
            flag_metadata=detail.flag_metadata,
        )

    def resolve_object_details(
        self,
        flag_key: str,
        default_value: Sequence[Any] | Mapping[str, Any],
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[Sequence[Any] | Mapping[str, Any]]:
        detail = self._evaluate(flag_key, default_value, evaluation_context, VALUE_TYPE_JSON)
        if detail.error_code is not None:
            return detail
        if not isinstance(detail.value, (dict, list)):
            return _type_mismatch(default_value, flag_key, "object", type(detail.value).__name__)
        return FlagResolutionDetails(
            value=detail.value,
            reason=detail.reason,
            variant=detail.variant,
            flag_metadata=detail.flag_metadata,
        )

    def _resolve(
        self,
        flag_key: str,
        default_value: Any,
        evaluation_context: EvaluationContext | None,
        expected_type: str,
        py_type: type,
    ) -> FlagResolutionDetails[Any]:
        detail = self._evaluate(flag_key, default_value, evaluation_context, expected_type)
        if detail.error_code is not None:
            return detail
        if not isinstance(detail.value, py_type):
            return _type_mismatch(default_value, flag_key, expected_type, type(detail.value).__name__)
        return detail

    def _evaluate(
        self,
        flag_key: str,
        default_value: Any,
        evaluation_context: EvaluationContext | None,
        expected_type: str,
    ) -> FlagResolutionDetails[Any]:
        try:
            context = _context_to_data(evaluation_context)
        except TypeError as exc:
            return _error_details(default_value, ErrorCode.INVALID_CONTEXT, str(exc))

        evaluator, last_error, ready = self._cache.snapshot()
        if evaluator is None:
            if ready and last_error is not None:
                return _error_details(default_value, ErrorCode.GENERAL, str(last_error))
            message = str(last_error) if last_error is not None else "flagcel: definitions not loaded"
            return _error_details(default_value, ErrorCode.PROVIDER_NOT_READY, message)

        result = evaluator.evaluate(flag_key, context)
        if result.error:
            return _result_error_details(default_value, result)
        if result.value_type != expected_type:
            return _type_mismatch(default_value, flag_key, expected_type, type(result.value).__name__)

        return FlagResolutionDetails(
            value=result.value,
            reason=_map_reason(result.reason),
            variant=result.variant,
            flag_metadata={
                "flagcelReason": result.reason,
                "valueType": result.value_type,
            },
        )

    def _poll(self, stop_event: threading.Event) -> None:
        while not stop_event.wait(self._poll_interval):
            try:
                self._refresh()
            except Exception:
                pass

    def _refresh(self) -> None:
        current_etag = self._cache.etag_value()
        try:
            result = self._client.fetch_definitions(current_etag)
            if result.unchanged:
                self._cache.mark_unchanged(result.etag)
                return
            if result.definitions is None:
                raise DefinitionsClientError("flagcel: definitions response missing data")
            self._cache.store(result.definitions, result.etag)
        except Exception as exc:
            self._cache.mark_error(exc)
            raise


class _DefinitionsCache:
    def __init__(self) -> None:
        self._lock = threading.RLock()
        self._evaluator: Evaluator | None = None
        self._etag = ""
        self._last_error: Exception | None = None
        self._ready = False

    def snapshot(self) -> tuple[Evaluator | None, Exception | None, bool]:
        with self._lock:
            return self._evaluator, self._last_error, self._ready

    def etag_value(self) -> str:
        with self._lock:
            return self._etag

    def mark_unchanged(self, etag: str) -> None:
        with self._lock:
            if etag:
                self._etag = etag
            self._last_error = None
            self._ready = True

    def store(self, definitions: dict[str, Any], etag: str) -> None:
        try:
            evaluator = Evaluator.load(definitions)
        except Exception as exc:
            error = ValueError(f"flagcel: compile definitions: {exc}")
            self.mark_error(error)
            raise error from exc

        with self._lock:
            self._evaluator = evaluator
            self._etag = etag
            self._last_error = None
            self._ready = True

    def mark_error(self, error: Exception) -> None:
        with self._lock:
            self._last_error = error


def _context_to_data(evaluation_context: EvaluationContext | None) -> dict[str, Any]:
    if evaluation_context is None:
        return {}
    context = dict(evaluation_context.attributes)
    if evaluation_context.targeting_key is not None:
        context["targetingKey"] = evaluation_context.targeting_key
    return json.loads(json.dumps(context, separators=(",", ":")))


def _error_details(default_value: Any, error_code: ErrorCode, message: str) -> FlagResolutionDetails[Any]:
    return FlagResolutionDetails(
        value=default_value,
        reason=Reason.ERROR,
        error_code=error_code,
        error_message=message,
    )


def _result_error_details(default_value: Any, result: EvaluationResult) -> FlagResolutionDetails[Any]:
    error_code = ErrorCode.FLAG_NOT_FOUND if result.reason == "not_found" else ErrorCode.GENERAL
    return _error_details(default_value, error_code, result.error)


def _type_mismatch(default_value: Any, flag_key: str, expected: str, got: str) -> FlagResolutionDetails[Any]:
    return _error_details(
        default_value,
        ErrorCode.TYPE_MISMATCH,
        f"{flag_key}: expected {expected}, got {got}",
    )


def _map_reason(reason: str) -> Reason | str:
    if reason == "matched_rule":
        return Reason.TARGETING_MATCH
    if reason == "default_no_match":
        return Reason.DEFAULT
    if reason == "disabled":
        return Reason.DISABLED
    if reason in {"not_found", "cel_error", "error"}:
        return Reason.ERROR
    return reason
