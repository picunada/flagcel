from __future__ import annotations

import multiprocessing as mp
import os
import statistics
import threading
import time
from collections import Counter
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass, field
from typing import Any

from flagcel_openfeature import FlagcelProvider
from openfeature import api
from openfeature.evaluation_context import EvaluationContext


@dataclass
class WorkerResult:
    worker_id: int = -1
    count: int = 0
    errors: int = 0
    mismatches: int = 0
    slow_count: int = 0
    slow_events: list[SlowEvent] = field(default_factory=list)
    latencies_ms: list[float] = field(default_factory=list)
    context_ms: list[float] = field(default_factory=list)
    openfeature_ms: list[float] = field(default_factory=list)
    result_ms: list[float] = field(default_factory=list)
    values: Counter[str] = field(default_factory=Counter)
    reasons: Counter[str] = field(default_factory=Counter)


@dataclass
class SlowEvent:
    worker_id: int
    sequence: int
    latency_ms: float
    context_ms: float | None
    openfeature_ms: float
    result_ms: float | None


@dataclass(frozen=True)
class LoadConfig:
    endpoint: str
    api_key: str
    flag_key: str
    targeting_key: str
    user_role: str
    expected_variant: str | None
    concurrency: int
    duration_seconds: float
    warmup_requests: int
    unique_targets: bool
    profile_steps: bool
    slow_threshold_ms: float
    slow_event_limit: int


_worker_local = threading.local()


def main() -> None:
    concurrency = env_int("LOAD_CONCURRENCY", 8)
    config = LoadConfig(
        endpoint=env("FLAGCEL_ENDPOINT", "http://localhost:8080/api/v1"),
        api_key=os.getenv(
            "FLAGCEL_API_KEY", "fc_vkrhmuPR_0q7XfHuvQ4E2ugfnQLiMNnJ35CJRGlzDgWSa9_FIyQw"
        ),
        flag_key=env("FLAGCEL_FLAG_KEY", "user-variant"),
        targeting_key=env("TARGETING_KEY", "u_123"),
        user_role=env("USER_ROLE", "example"),
        expected_variant=os.getenv("EXPECTED_VARIANT", env("USER_ROLE", "example")),
        concurrency=concurrency,
        duration_seconds=env_float("LOAD_DURATION_SECONDS", 30.0),
        warmup_requests=env_int("LOAD_WARMUP_REQUESTS", 100),
        unique_targets=env_bool("LOAD_UNIQUE_TARGETS", True),
        profile_steps=env_bool("LOAD_PROFILE_STEPS", False),
        slow_threshold_ms=env_float("LOAD_SLOW_THRESHOLD_US", 1_000.0) / 1_000.0,
        slow_event_limit=env_int("LOAD_SLOW_EVENT_LIMIT", 20),
    )
    processes = env_int("LOAD_PROCESSES", 1)

    if processes <= 1:
        results, elapsed = run_process(config, process_id=0, verbose=True)
        report(
            results, elapsed, processes=1, slow_threshold_ms=config.slow_threshold_ms
        )
        return

    results, elapsed = run_processes(config, processes)
    report(
        results,
        elapsed,
        processes=processes,
        slow_threshold_ms=config.slow_threshold_ms,
    )


def run_processes(
    config: LoadConfig, processes: int
) -> tuple[list[WorkerResult], float]:
    output: mp.Queue[Any] = mp.Queue()
    workers = [
        mp.Process(
            target=run_process_entry,
            args=(config, process_id, output),
            name=f"flagcel-load-{process_id}",
        )
        for process_id in range(processes)
    ]

    for worker in workers:
        worker.start()

    results: list[WorkerResult] = []
    elapsed_by_process: list[float] = []
    errors: list[str] = []
    for _ in workers:
        process_id, elapsed, process_results, error = output.get()
        if error is not None:
            errors.append(f"process {process_id}: {error}")
        else:
            elapsed_by_process.append(elapsed)
            results.extend(process_results)

    for worker in workers:
        worker.join()

    if errors:
        raise RuntimeError("; ".join(errors))

    elapsed = max(elapsed_by_process) if elapsed_by_process else 0.0
    return results, elapsed


def run_process_entry(
    config: LoadConfig,
    process_id: int,
    output: mp.Queue[Any],
) -> None:
    try:
        results, elapsed = run_process(config, process_id=process_id, verbose=False)
        output.put((process_id, elapsed, results, None))
    except Exception as exc:
        output.put((process_id, 0.0, [], f"{type(exc).__name__}: {exc}"))


def run_process(
    config: LoadConfig,
    process_id: int,
    verbose: bool,
) -> tuple[list[WorkerResult], float]:
    provider = FlagcelProvider(
        endpoint=config.endpoint,
        api_key=config.api_key,
    )
    provider.initialize()
    api.set_provider(provider)

    try:
        client = api.get_client(f"flagcel-python-load-test-{process_id}")
        base_context = context_for(config.targeting_key, config.user_role)

        first = client.get_object_details(config.flag_key, {}, base_context)
        if first.error_code is not None:
            raise RuntimeError(
                f"initial evaluation failed: {first.error_code} {first.error_message}"
            )

        if verbose:
            print(
                "loaded "
                f"flag={config.flag_key} endpoint={config.endpoint} "
                f"value={first.value} reason={first.reason} variant={first.variant}"
            )

        for _ in range(config.warmup_requests):
            client.get_object_value(config.flag_key, {}, base_context)

        started = time.perf_counter()
        deadline = started + config.duration_seconds

        with ThreadPoolExecutor(max_workers=config.concurrency) as executor:
            futures = [
                executor.submit(
                    run_worker,
                    client,
                    config.flag_key,
                    config.targeting_key,
                    config.user_role,
                    config.expected_variant,
                    config.unique_targets,
                    config.profile_steps,
                    config.slow_threshold_ms,
                    config.slow_event_limit,
                    deadline,
                    process_id * config.concurrency + worker_id,
                )
                for worker_id in range(config.concurrency)
            ]
            results = [future.result() for future in futures]

        elapsed = time.perf_counter() - started
        return results, elapsed
    finally:
        provider.shutdown()


def run_worker(
    client: Any,
    flag_key: str,
    targeting_key: str,
    user_role: str,
    expected_variant: str | None,
    unique_targets: bool,
    profile_steps: bool,
    slow_threshold_ms: float,
    slow_event_limit: int,
    deadline: float,
    worker_id: int,
) -> WorkerResult:
    result = WorkerResult(worker_id=worker_id)
    _worker_local.result = result
    _worker_local.profile_steps = profile_steps
    sequence = 0
    shared_context = None if unique_targets else context_for(targeting_key, user_role)
    measure_steps = profile_steps or slow_threshold_ms > 0

    try:
        while time.perf_counter() < deadline:
            loop_started = time.perf_counter_ns()
            if measure_steps:
                context_started = time.perf_counter_ns()
            current_target = (
                f"{targeting_key}-{worker_id}-{sequence}"
                if unique_targets
                else targeting_key
            )
            eval_context = shared_context or context_for(current_target, user_role)
            context_elapsed_ms = None
            result_elapsed_ms = None
            if measure_steps:
                context_elapsed_ms = (
                    time.perf_counter_ns() - context_started
                ) / 1_000_000
            if profile_steps:
                result.context_ms.append(context_elapsed_ms or 0.0)

            eval_started = time.perf_counter_ns()
            details = client.get_object_details(flag_key, {}, eval_context)
            eval_elapsed_ms = (time.perf_counter_ns() - eval_started) / 1_000_000
            if profile_steps:
                result.openfeature_ms.append(eval_elapsed_ms)

            if measure_steps:
                result_started = time.perf_counter_ns()
            result.count += 1
            result.reasons[str(details.reason)] += 1

            if details.error_code is not None:
                result.errors += 1
            value = details.value
            result.values[repr(value)] += 1

            if expected_variant:
                actual_variant = (
                    value.get("variant") if isinstance(value, dict) else None
                )
                if actual_variant != expected_variant:
                    result.mismatches += 1

            if measure_steps:
                result_elapsed_ms = (
                    time.perf_counter_ns() - result_started
                ) / 1_000_000
            if profile_steps:
                result.result_ms.append(result_elapsed_ms or 0.0)
            latency_ms = (time.perf_counter_ns() - loop_started) / 1_000_000
            result.latencies_ms.append(latency_ms)
            if slow_threshold_ms > 0 and latency_ms >= slow_threshold_ms:
                result.slow_count += 1
                record_slow_event(
                    result,
                    SlowEvent(
                        worker_id=worker_id,
                        sequence=sequence,
                        latency_ms=latency_ms,
                        context_ms=context_elapsed_ms,
                        openfeature_ms=eval_elapsed_ms,
                        result_ms=result_elapsed_ms,
                    ),
                    slow_event_limit,
                )
            sequence += 1
    finally:
        _worker_local.result = None
        _worker_local.profile_steps = False

    return result


def record_slow_event(result: WorkerResult, event: SlowEvent, limit: int) -> None:
    if limit <= 0:
        return
    result.slow_events.append(event)
    if len(result.slow_events) > limit:
        result.slow_events.sort(key=lambda item: item.latency_ms, reverse=True)
        del result.slow_events[limit:]


def context_for(targeting_key: str, user_role: str) -> EvaluationContext:
    return EvaluationContext(
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
    )


def report(
    results: list[WorkerResult],
    elapsed: float,
    processes: int,
    slow_threshold_ms: float,
) -> None:
    total = sum(result.count for result in results)
    errors = sum(result.errors for result in results)
    mismatches = sum(result.mismatches for result in results)
    latencies = [latency for result in results for latency in result.latencies_ms]
    values = Counter[str]()
    reasons = Counter[str]()

    for result in results:
        values.update(result.values)
        reasons.update(result.reasons)

    print()
    print(f"processes={processes}")
    print(f"threads_per_process={len(results) // processes if processes else 0}")
    print(f"requests={total}")
    print(f"elapsed_seconds={elapsed:.3f}")
    print(f"throughput_per_second={total / elapsed if elapsed else 0.0:.2f}")
    print(f"errors={errors}")
    print(f"mismatches={mismatches}")
    print(f"reasons={dict(reasons)}")
    print(f"top_values={dict(values.most_common(5))}")

    if not latencies:
        return

    print_metric("latency_ms", latencies)
    print_step_metrics(results)
    print_slow_events(results, slow_threshold_ms)


def print_step_metrics(results: list[WorkerResult]) -> None:
    steps = {
        "context_create_ms": [
            value for result in results for value in result.context_ms
        ],
        "openfeature_eval_ms": [
            value for result in results for value in result.openfeature_ms
        ],
        "result_collect_ms": [
            value for result in results for value in result.result_ms
        ],
    }
    if not any(steps.values()):
        return

    print()
    print("step_latency_ms:")
    for name, values in steps.items():
        print_metric(name, values)

    print()
    print(
        "notes: openfeature_eval_ms includes the OpenFeature client call, provider "
        "context normalization, local CEL evaluation, and result mapping."
    )


def print_slow_events(results: list[WorkerResult], threshold_ms: float) -> None:
    if threshold_ms <= 0:
        return

    slow_count = sum(result.slow_count for result in results)
    slow_by_worker = {
        result.worker_id: result.slow_count
        for result in results
        if result.slow_count > 0
    }
    slow_events = [event for result in results for event in result.slow_events]
    slow_events.sort(key=lambda event: event.latency_ms, reverse=True)

    print()
    print("slow_events:")
    print(f"threshold_us={threshold_ms * 1_000:.3f}")
    print(f"count={slow_count}")
    print(f"workers={slow_by_worker}")
    if not slow_events:
        print("top=[]")
        return

    print("top:")
    for event in slow_events[:10]:
        print(
            "  "
            f"worker={event.worker_id} "
            f"seq={event.sequence} "
            f"latency_us={to_us(event.latency_ms)} "
            f"context_us={to_optional_us(event.context_ms)} "
            f"openfeature_us={to_us(event.openfeature_ms)} "
            f"result_us={to_optional_us(event.result_ms)}"
        )


def to_us(value_ms: float) -> str:
    return f"{value_ms * 1_000:.3f}"


def to_optional_us(value_ms: float | None) -> str:
    if value_ms is None:
        return "na"
    return to_us(value_ms)


def print_metric(name: str, values: list[float]) -> None:
    if not values:
        print(f"{name} count=0")
        return

    values.sort()
    unit, scale = metric_unit(values)
    print(
        f"{name} "
        f"count={len(values)} "
        f"unit={unit} "
        f"min={values[0] * scale:.3f} "
        f"avg={statistics.fmean(values) * scale:.3f} "
        f"p50={percentile(values, 50) * scale:.3f} "
        f"p95={percentile(values, 95) * scale:.3f} "
        f"p99={percentile(values, 99) * scale:.3f} "
        f"max={values[-1] * scale:.3f}"
    )


def metric_unit(values_ms: list[float]) -> tuple[str, float]:
    if percentile(values_ms, 99) < 1.0:
        return "us", 1_000.0
    return "ms", 1.0


def percentile(sorted_values: list[float], percentile_value: float) -> float:
    if len(sorted_values) == 1:
        return sorted_values[0]
    index = (len(sorted_values) - 1) * (percentile_value / 100)
    lower = int(index)
    upper = min(lower + 1, len(sorted_values) - 1)
    weight = index - lower
    return sorted_values[lower] * (1 - weight) + sorted_values[upper] * weight


def env(key: str, fallback: str) -> str:
    return os.getenv(key) or fallback


def env_int(key: str, fallback: int) -> int:
    value = os.getenv(key)
    if not value:
        return fallback
    return int(value)


def env_float(key: str, fallback: float) -> float:
    value = os.getenv(key)
    if not value:
        return fallback
    return float(value)


def env_bool(key: str, fallback: bool) -> bool:
    value = os.getenv(key)
    if not value:
        return fallback
    return value.lower() in {"1", "true", "yes", "on"}


if __name__ == "__main__":
    main()
