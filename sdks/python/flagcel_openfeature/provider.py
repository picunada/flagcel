from __future__ import annotations

import json
import math
from collections.abc import Mapping, Sequence
from dataclasses import dataclass
from typing import Any

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import FlagResolutionDetails, Reason
from openfeature.hook import Hook
from openfeature.provider import AbstractProvider, Metadata

from .client import EvalClient, EvalClientError, URLopener


PROVIDER_NAME = "flagcel"


@dataclass(frozen=True)
class FlagcelProviderOptions:
    endpoint: str
    api_key: str = ""
    http_client: URLopener | None = None
    timeout: float | None = None


class FlagcelProvider(AbstractProvider):
    def __init__(
        self,
        endpoint: str | FlagcelProviderOptions,
        api_key: str = "",
        http_client: URLopener | None = None,
        timeout: float | None = None,
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
            )

        self._client = EvalClient(
            options.endpoint,
            options.api_key,
            opener=options.http_client,
            timeout=options.timeout,
        )

    def get_metadata(self) -> Metadata:
        return Metadata(name=PROVIDER_NAME)

    def get_provider_hooks(self) -> list[Hook]:
        return []

    def initialize(self, evaluation_context: EvaluationContext | None = None) -> None:
        del evaluation_context

    def shutdown(self) -> None:
        return

    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[bool]:
        return self._resolve(flag_key, default_value, evaluation_context, "boolean", bool)

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[str]:
        return self._resolve(flag_key, default_value, evaluation_context, "string", str)

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: EvaluationContext | None = None,
    ) -> FlagResolutionDetails[int]:
        detail = self._evaluate(flag_key, default_value, evaluation_context)
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
        detail = self._evaluate(flag_key, default_value, evaluation_context)
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
        detail = self._evaluate(flag_key, default_value, evaluation_context)
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
        detail = self._evaluate(flag_key, default_value, evaluation_context)
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
    ) -> FlagResolutionDetails[Any]:
        try:
            context_json = _context_to_json(evaluation_context)
        except TypeError as exc:
            return _error_details(default_value, ErrorCode.INVALID_CONTEXT, str(exc))

        try:
            result = self._client.evaluate_json(flag_key, context_json)
        except EvalClientError as exc:
            error_code = ErrorCode.FLAG_NOT_FOUND if exc.status == 404 else ErrorCode.GENERAL
            return _error_details(default_value, error_code, str(exc))
        except Exception as exc:
            return _error_details(default_value, ErrorCode.GENERAL, str(exc))

        if result.get("error"):
            error_code = (
                ErrorCode.FLAG_NOT_FOUND
                if result.get("reason") == "not_found"
                else ErrorCode.GENERAL
            )
            return _error_details(default_value, error_code, str(result["error"]))

        reason = str(result.get("reason", ""))
        return FlagResolutionDetails(
            value=result.get("value"),
            reason=_map_reason(reason),
            variant=result.get("variant"),
            flag_metadata={
                "flagcelReason": reason,
                "valueType": str(result.get("value_type", "")),
            },
        )


def _context_to_json(evaluation_context: EvaluationContext | None) -> str:
    if evaluation_context is None:
        return "{}"
    context = dict(evaluation_context.attributes)
    if evaluation_context.targeting_key is not None:
        context["targetingKey"] = evaluation_context.targeting_key
    return json.dumps(context, separators=(",", ":"))


def _error_details(default_value: Any, error_code: ErrorCode, message: str) -> FlagResolutionDetails[Any]:
    return FlagResolutionDetails(
        value=default_value,
        reason=Reason.ERROR,
        error_code=error_code,
        error_message=message,
    )


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
