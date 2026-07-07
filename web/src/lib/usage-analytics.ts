import type { FlagUsage, FlagUsageBucket } from "$lib/api";

export type UsageRangeHours = 24 | 168 | 720;

export const usageRangeOptions: { value: UsageRangeHours; label: string }[] = [
    { value: 24, label: "24h" },
    { value: 168, label: "7d" },
    { value: 720, label: "30d" },
];

export type BreakdownKey =
    | "flag_key"
    | "reason"
    | "value"
    | "matched_rule_id"
    | "source"
    | "api_key_id";

export type TrendBreakdownKey = "flag_key" | "source" | "api_key_id";

export type CountPoint = {
    bucket_start: string;
    label: string;
    count: number;
};

export type BreakdownPoint = {
    name: string;
    count: number;
    percent: number;
};

export type LatencyPoint = {
    bucket_start: string;
    label: string;
    avg_latency_ms: number;
    p95_latency_ms: number;
    count: number;
};

export type TrendBreakdownRow = CountPoint & Record<string, string | number>;

export type TrendBreakdownSeries = {
    key: string;
    label: string;
};

export type TrendBreakdownAreaSeries = {
    series: TrendBreakdownSeries[];
    yDomain: [number, number];
    rows: TrendBreakdownRow[];
};

export function usageQuery(hours?: number) {
    return hours ? `?hours=${encodeURIComponent(String(hours))}` : "";
}

export function parseUsageRange(value: string | null | undefined): UsageRangeHours {
    const parsed = Number(value);
    return usageRangeOptions.some((option) => option.value === parsed)
        ? (parsed as UsageRangeHours)
        : 24;
}

export function normalizeUsage(usage: Partial<FlagUsage> | null | undefined): FlagUsage {
    return {
        buckets: usage?.buckets ?? [],
        latency_buckets: usage?.latency_buckets ?? [],
        events: usage?.events ?? [],
    };
}

export function usageTotal(usage: FlagUsage) {
    return usage.buckets.reduce((sum, bucket) => sum + bucket.count, 0);
}

// Chart buckets are coarser than the backend's hourly buckets on longer ranges
// so a full zero-filled range stays readable: 24h -> hourly, 7d -> 6h, 30d -> daily.
const rangeStepHours: Record<UsageRangeHours, number> = {
    24: 1,
    168: 6,
    720: 24,
};

// Axis ticks are a subset of the band domain so the full range stays readable:
// 24h -> every 3 hours, 7d -> daily (at midnight UTC), 30d -> every 5 days.
export function usageAxisTickLabels(points: CountPoint[], hours: UsageRangeHours): string[] {
    if (hours === 168) {
        return points
            .filter((point) => point.bucket_start.endsWith("T00:00:00Z"))
            .map((point) => point.label);
    }
    const step = hours === 24 ? 3 : 5;
    return points.filter((_, index) => index % step === 0).map((point) => point.label);
}

// Band labels carry the full bucket time for tooltips; axis ticks only show the
// day on multi-day ranges to avoid label collisions.
export function usageAxisFormat(hours: UsageRangeHours) {
    if (hours !== 168) return undefined;
    return (label: string) => label.split(",")[0];
}

// Pads the y-axis so the tallest point sits at 80% of the chart height.
export function usageYDomain(values: number[]): [number, number] {
    const max = Math.max(0, ...values);
    return [0, max === 0 ? 1 : max / 0.8];
}

function rangeEpochs(hours: UsageRangeHours, now: Date) {
    const stepMs = rangeStepHours[hours] * 3_600_000;
    const end = Math.floor(now.getTime() / stepMs) * stepMs;
    const count = Math.ceil(hours / rangeStepHours[hours]);
    return {
        stepMs,
        epochs: Array.from({ length: count }, (_, i) => end - (count - 1 - i) * stepMs),
    };
}

function alignEpoch(bucketStart: string, stepMs: number) {
    const time = Date.parse(bucketStart);
    if (Number.isNaN(time)) return null;
    return Math.floor(time / stepMs) * stepMs;
}

function countPoint(epoch: number, hours: UsageRangeHours, count: number): CountPoint {
    return {
        bucket_start: new Date(epoch).toISOString().replace(".000Z", "Z"),
        label: timeLabel(epoch, hours),
        count,
    };
}

export function evaluationTrendSeries(
    usage: FlagUsage,
    hours: UsageRangeHours,
    now = new Date(),
): CountPoint[] {
    const { stepMs, epochs } = rangeEpochs(hours, now);
    const counts = new Map<number, number>();
    for (const bucket of usage.buckets) {
        const epoch = alignEpoch(bucket.bucket_start, stepMs);
        if (epoch === null) continue;
        counts.set(epoch, (counts.get(epoch) ?? 0) + bucket.count);
    }
    return epochs.map((epoch) => countPoint(epoch, hours, counts.get(epoch) ?? 0));
}

export function trendBreakdownAreaSeries(
    usage: FlagUsage,
    key: TrendBreakdownKey,
    hours: UsageRangeHours,
    limit = 5,
    now = new Date(),
): TrendBreakdownAreaSeries {
    const { stepMs, epochs } = rangeEpochs(hours, now);
    const totals = new Map<string, number>();
    const bucketCounts = new Map<number, Map<string, number>>();

    for (const bucket of usage.buckets) {
        const epoch = alignEpoch(bucket.bucket_start, stepMs);
        if (epoch === null) continue;
        const name = bucketLabel(bucket, key);
        totals.set(name, (totals.get(name) ?? 0) + bucket.count);

        const counts = bucketCounts.get(epoch) ?? new Map<string, number>();
        counts.set(name, (counts.get(name) ?? 0) + bucket.count);
        bucketCounts.set(epoch, counts);
    }

    const labels = [...totals.entries()]
        .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
        .slice(0, limit)
        .map(([name]) => name);
    const series = labels.map((label, index) => ({
        key: `series_${index}`,
        label,
    }));

    let maxSeriesValue = 0;
    const rows = epochs.map((epoch) => {
        const counts = bucketCounts.get(epoch);
        const row: TrendBreakdownRow = {
            ...countPoint(
                epoch,
                hours,
                labels.reduce((sum, name) => sum + (counts?.get(name) ?? 0), 0),
            ),
        };
        for (const [index, name] of labels.entries()) {
            const value = counts?.get(name) ?? 0;
            maxSeriesValue = Math.max(maxSeriesValue, value);
            row[`series_${index}`] = value;
        }
        return row;
    });

    return {
        series,
        yDomain: usageYDomain([maxSeriesValue]),
        rows,
    };
}

export function breakdownSeries(usage: FlagUsage, key: BreakdownKey, limit = 6): BreakdownPoint[] {
    const total = usageTotal(usage);
    const counts = new Map<string, number>();
    for (const bucket of usage.buckets) {
        const name = bucketLabel(bucket, key);
        counts.set(name, (counts.get(name) ?? 0) + bucket.count);
    }
    return [...counts.entries()]
        .map(([name, count]) => ({
            name,
            count,
            percent: total === 0 ? 0 : Math.round((count / total) * 100),
        }))
        .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name))
        .slice(0, limit);
}

export function latencySeries(
    usage: FlagUsage,
    hours: UsageRangeHours,
    now = new Date(),
): LatencyPoint[] {
    const { stepMs, epochs } = rangeEpochs(hours, now);
    const buckets = new Map<
        number,
        { count: number; weightedAvgTotal: number; p95_latency_ms: number }
    >();
    for (const bucket of usage.latency_buckets) {
        const epoch = alignEpoch(bucket.bucket_start, stepMs);
        if (epoch === null) continue;
        const current = buckets.get(epoch) ?? {
            count: 0,
            weightedAvgTotal: 0,
            p95_latency_ms: 0,
        };
        current.count += bucket.count;
        current.weightedAvgTotal += bucket.avg_latency_ms * bucket.count;
        current.p95_latency_ms = Math.max(current.p95_latency_ms, bucket.p95_latency_ms);
        buckets.set(epoch, current);
    }
    return epochs.map((epoch) => {
        const bucket = buckets.get(epoch);
        return {
            ...countPoint(epoch, hours, bucket?.count ?? 0),
            avg_latency_ms: roundMs(
                bucket && bucket.count > 0 ? bucket.weightedAvgTotal / bucket.count : 0,
            ),
            p95_latency_ms: roundMs(bucket?.p95_latency_ms ?? 0),
        };
    });
}

export function latencySourceAreaSeries(
    usage: FlagUsage,
    hours: UsageRangeHours,
    limit = 5,
    now = new Date(),
): TrendBreakdownAreaSeries {
    const { stepMs, epochs } = rangeEpochs(hours, now);
    const totals = new Map<string, number>();
    const bucketP95s = new Map<number, Map<string, number>>();

    for (const bucket of usage.latency_buckets) {
        const epoch = alignEpoch(bucket.bucket_start, stepMs);
        if (epoch === null) continue;
        const name = bucket.source || "unknown";
        totals.set(name, (totals.get(name) ?? 0) + bucket.count);

        const p95s = bucketP95s.get(epoch) ?? new Map<string, number>();
        p95s.set(name, Math.max(p95s.get(name) ?? 0, bucket.p95_latency_ms));
        bucketP95s.set(epoch, p95s);
    }

    const labels = [...totals.entries()]
        .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
        .slice(0, limit)
        .map(([name]) => name);
    const series = labels.map((label, index) => ({
        key: `latency_${index}`,
        label,
    }));

    let maxSeriesValue = 0;
    const rows = epochs.map((epoch) => {
        const p95s = bucketP95s.get(epoch);
        const row: TrendBreakdownRow = { ...countPoint(epoch, hours, 0) };
        for (const [index, name] of labels.entries()) {
            const value = roundMs(p95s?.get(name) ?? 0);
            maxSeriesValue = Math.max(maxSeriesValue, value);
            row[`latency_${index}`] = value;
        }
        return row;
    });

    return {
        series,
        yDomain: usageYDomain([maxSeriesValue]),
        rows,
    };
}

function bucketLabel(bucket: FlagUsageBucket, key: BreakdownKey) {
    if (key === "value") return formatValue(bucket.value);
    if (key === "api_key_id") return bucket.api_key_name || bucket.api_key_id || "unknown";
    return bucket[key] || "unknown";
}

function formatValue(value: unknown) {
    if (typeof value === "string") return value;
    if (typeof value === "number" || typeof value === "boolean") return String(value);
    if (value === null) return "null";
    return JSON.stringify(value);
}

const dayFormatter = new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
});

function timeLabel(epoch: number, hours: UsageRangeHours) {
    const date = new Date(epoch);
    const day = dayFormatter.format(date);
    const time = `${String(date.getUTCHours()).padStart(2, "0")}:${String(date.getUTCMinutes()).padStart(2, "0")}`;
    const stepHours = rangeStepHours[hours];
    if (stepHours >= 24) return day;
    if (stepHours > 1) return `${day}, ${time}`;
    return time;
}

// Sub-millisecond latencies are common for local evaluation, so keep two
// decimals below 10ms and one above to avoid rounding everything to 0.
function roundMs(value: number) {
    const factor = value < 10 ? 100 : 10;
    return Math.round(value * factor) / factor;
}
