from __future__ import annotations

import hashlib
import math
import re
from dataclasses import dataclass
from typing import Any

from cel_expr_python import cel


VALUE_TYPE_BOOLEAN = "boolean"
VALUE_TYPE_STRING = "string"
VALUE_TYPE_NUMBER = "number"
VALUE_TYPE_JSON = "json"


@dataclass(frozen=True)
class Rollout:
    percentage: int = 0
    bucket_by: str = ""


@dataclass(frozen=True)
class CompiledRule:
    id: str
    source: str
    program: Any
    rollout: Rollout
    value: Any


@dataclass(frozen=True)
class CompiledFlag:
    key: str
    value_type: str
    enabled: bool
    default_value: Any
    rules: tuple[CompiledRule, ...]


@dataclass(frozen=True)
class EvaluationResult:
    key: str
    value: Any
    value_type: str
    reason: str
    variant: str | None = None
    error: str = ""


class Evaluator:
    def __init__(self, flags: dict[str, CompiledFlag]) -> None:
        self._flags = flags

    @classmethod
    def load(cls, definitions: dict[str, Any]) -> Evaluator:
        contexts = {
            str(context.get("id", "")): context
            for context in definitions.get("contexts", [])
            if context.get("id")
        }

        flags: dict[str, CompiledFlag] = {}
        for index, definition in enumerate(definitions.get("flags", [])):
            key = str(definition.get("key", ""))
            try:
                schema = definition.get("context_schema")
                context_id = definition.get("context_id")
                if schema is None and context_id:
                    schema = contexts.get(str(context_id))
                flags[key] = _compile_flag(definition, schema)
            except Exception as exc:
                raise ValueError(f'flags[{index}] "{key}": {exc}') from exc
        return cls(flags)

    def evaluate(self, key: str, data: dict[str, Any]) -> EvaluationResult:
        flag = self._flags.get(key)
        if flag is None:
            return EvaluationResult(
                key=key,
                value=False,
                value_type=VALUE_TYPE_BOOLEAN,
                reason="not_found",
                error="flag not found",
            )
        return _evaluate_flag(flag, data)


def _compile_flag(definition: dict[str, Any], schema: dict[str, Any] | None) -> CompiledFlag:
    key = str(definition.get("key", ""))
    value_type = str(definition.get("type") or VALUE_TYPE_BOOLEAN)
    default_value = definition.get("default_value")
    if default_value is None and value_type == VALUE_TYPE_BOOLEAN:
        default_value = False

    env = cel.NewEnv(variables=_variables_for_schema(schema))
    rules: list[CompiledRule] = []
    for index, rule in enumerate(definition.get("rules") or []):
        source = str(rule.get("expression") or "")
        program = None
        if source:
            try:
                program = env.compile(source)
            except Exception as exc:
                raise ValueError(f"rule {index}: compile: {exc}") from exc
            if program.return_type() != cel.Type.BOOL:
                raise ValueError(
                    f"rule {index}: expression must return bool type, got {program.return_type()}"
                )

        rollout = rule.get("rollout") or {}
        value = rule.get("value")
        if value is None and value_type == VALUE_TYPE_BOOLEAN:
            value = True

        rules.append(
            CompiledRule(
                id=str(rule.get("id") or ""),
                source=source,
                program=program,
                rollout=Rollout(
                    percentage=_int_value(rollout.get("percentage"), 0),
                    bucket_by=str(rollout.get("bucket_by") or ""),
                ),
                value=value,
            )
        )

    return CompiledFlag(
        key=key,
        value_type=value_type,
        enabled=bool(definition.get("enabled")),
        default_value=default_value,
        rules=tuple(rules),
    )


def _evaluate_flag(flag: CompiledFlag, data: dict[str, Any]) -> EvaluationResult:
    value = _flag_value(flag.value_type, flag.default_value)
    if not flag.enabled:
        return EvaluationResult(
            key=flag.key,
            value=value,
            value_type=flag.value_type,
            reason="disabled",
        )

    first_error = ""
    for rule in flag.rules:
        matches, error = _evaluate_expression(rule.program, data)
        if error:
            if not first_error:
                first_error = error
            continue
        if not matches:
            continue

        if _bucket(flag.key, data, rule.rollout):
            value = _flag_value(flag.value_type, rule.value)
        return EvaluationResult(
            key=flag.key,
            value=value,
            value_type=flag.value_type,
            reason="matched_rule",
            variant=rule.id,
            error=first_error,
        )

    return EvaluationResult(
        key=flag.key,
        value=value,
        value_type=flag.value_type,
        reason="default_no_match",
        error=first_error,
    )


def _evaluate_expression(program: Any, data: dict[str, Any]) -> tuple[bool, str]:
    if program is None:
        return True, ""
    try:
        result = program.eval(data=data)
    except Exception as exc:
        return False, f"eval: {exc}"
    if result.type() != cel.Type.BOOL:
        return False, f"expression returned non-bool: {result.type()}"
    return bool(result.value()), ""


def _flag_value(value_type: str, value: Any) -> Any:
    if value is None and value_type == VALUE_TYPE_BOOLEAN:
        return False
    return value


def _bucket(flag_key: str, data: dict[str, Any], rollout: Rollout) -> bool:
    bucket_attr = rollout.bucket_by or "id"
    bucket_value, found = _lookup_path(data, bucket_attr)
    if not found:
        return rollout.percentage >= 100

    digest = hashlib.sha1(f"{flag_key}:{bucket_value}".encode("utf-8")).digest()
    bucket_number = int.from_bytes(digest[:4], "big") % 100
    if rollout.percentage >= 100:
        return True
    if rollout.percentage <= 0:
        return False
    return bucket_number < rollout.percentage


def _lookup_path(data: dict[str, Any], path: str) -> tuple[Any, bool]:
    if path in data:
        return data[path], True

    current: Any = data
    for part in path.split("."):
        if isinstance(current, dict) and part in current:
            current = current[part]
            continue
        return None, False
    return current, True


def _variables_for_schema(schema: dict[str, Any] | None) -> dict[str, Any]:
    if not schema:
        return {}

    roots: dict[str, Any] = {}
    for field in schema.get("fields", []):
        path = str(field.get("path") or "")
        root = path.split(".", 1)[0]
        if not _is_cel_identifier(root):
            continue
        if "." in path:
            roots[root] = cel.Type.MAP
        elif root not in roots:
            roots[root] = _cel_type_for_context_type(str(field.get("type") or ""))
    return {key: roots[key] for key in sorted(roots)}


def _cel_type_for_context_type(context_type: str) -> Any:
    if context_type in {"string", "timestamp"}:
        return cel.Type.STRING
    if context_type == "int":
        return cel.Type.INT
    if context_type == "double":
        return cel.Type.DOUBLE
    if context_type == "bool":
        return cel.Type.BOOL
    if context_type == "list":
        return cel.Type.LIST
    if context_type == "map":
        return cel.Type.MAP
    return cel.Type.DYN


def _int_value(value: Any, fallback: int) -> int:
    if isinstance(value, bool):
        return fallback
    if isinstance(value, int):
        return value
    if isinstance(value, float) and math.trunc(value) == value:
        return int(value)
    return fallback


_CEL_IDENTIFIER_RE = re.compile(r"^[^\W\d]\w*$", re.UNICODE)


def _is_cel_identifier(value: str) -> bool:
    return bool(_CEL_IDENTIFIER_RE.match(value))
