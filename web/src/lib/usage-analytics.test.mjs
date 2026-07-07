import assert from "node:assert/strict";

const {
    breakdownSeries,
    evaluationTrendSeries,
    latencySeries,
    latencySourceAreaSeries,
    trendBreakdownAreaSeries,
    usageAxisFormat,
    usageAxisTickLabels,
} = await import("./usage-analytics.ts");

const now = new Date("2026-07-04T12:30:00Z");

const usage = {
    buckets: [
        {
            bucket_start: "2026-07-04T11:00:00Z",
            flag_key: "checkout",
            value_type: "boolean",
            value: true,
            reason: "matched_rule",
            matched_rule_id: "rule-a",
            source: "js-sdk",
            count: 4,
        },
        {
            bucket_start: "2026-07-04T10:00:00Z",
            flag_key: "checkout",
            value_type: "boolean",
            value: false,
            reason: "default_no_match",
            source: "go-sdk",
            count: 2,
        },
        {
            bucket_start: "2026-07-04T10:00:00Z",
            flag_key: "search",
            value_type: "string",
            value: "variant-a",
            reason: "matched_rule",
            matched_rule_id: "rule-b",
            api_key_id: "key-search",
            api_key_name: "search key",
            source: "js-sdk",
            count: 8,
        },
        {
            bucket_start: "2026-07-04T12:00:00Z",
            flag_key: "billing",
            value_type: "boolean",
            value: true,
            reason: "matched_rule",
            matched_rule_id: "rule-c",
            source: "",
            count: 1,
        },
    ],
    latency_buckets: [
        {
            bucket_start: "2026-07-04T11:00:00Z",
            flag_key: "checkout",
            source: "js-sdk",
            count: 4,
            avg_latency_ms: 6,
            p95_latency_ms: 12,
        },
        {
            bucket_start: "2026-07-04T10:00:00Z",
            flag_key: "checkout",
            source: "js-sdk",
            count: 2,
            avg_latency_ms: 4,
            p95_latency_ms: 9,
        },
        {
            bucket_start: "2026-07-04T10:00:00Z",
            flag_key: "search",
            source: "go-sdk",
            count: 8,
            avg_latency_ms: 10,
            p95_latency_ms: 20,
        },
    ],
    events: [],
};

// The trend series is zero-filled across the whole selected range, so a 24h
// range always yields 24 hourly points ending at the current hour.
const trend = evaluationTrendSeries(usage, 24, now);
assert.equal(trend.length, 24);
assert.equal(trend[0].bucket_start, "2026-07-03T13:00:00Z");
assert.equal(trend.at(-1).bucket_start, "2026-07-04T12:00:00Z");
assert.ok(trend.every((point) => point.count >= 0));
assert.deepEqual(
    trend.filter((point) => point.count > 0),
    [
        { bucket_start: "2026-07-04T10:00:00Z", label: "10:00", count: 10 },
        { bucket_start: "2026-07-04T11:00:00Z", label: "11:00", count: 4 },
        { bucket_start: "2026-07-04T12:00:00Z", label: "12:00", count: 1 },
    ],
);

// Longer ranges use coarser buckets (7d -> 6h) with date-aware labels.
const weekTrend = evaluationTrendSeries(usage, 168, now);
assert.equal(weekTrend.length, 28);
assert.deepEqual(
    weekTrend.filter((point) => point.count > 0),
    [
        { bucket_start: "2026-07-04T06:00:00Z", label: "Jul 4, 06:00", count: 14 },
        { bucket_start: "2026-07-04T12:00:00Z", label: "Jul 4, 12:00", count: 1 },
    ],
);

// 30d -> daily buckets.
const monthTrend = evaluationTrendSeries(usage, 720, now);
assert.equal(monthTrend.length, 30);
assert.deepEqual(
    monthTrend.filter((point) => point.count > 0),
    [{ bucket_start: "2026-07-04T00:00:00Z", label: "Jul 4", count: 15 }],
);

// 7d axis ticks land on midnight UTC and render as day-only labels.
const weekTicks = usageAxisTickLabels(weekTrend, 168);
assert.equal(weekTicks.length, 7);
assert.equal(weekTicks[0], "Jun 28, 00:00");
assert.equal(usageAxisFormat(168)(weekTicks.at(-1)), "Jul 4");
assert.equal(usageAxisFormat(24), undefined);
assert.equal(usageAxisTickLabels(trend, 24).length, 8);

assert.deepEqual(breakdownSeries(usage, "reason"), [
    { name: "matched_rule", count: 13, percent: 87 },
    { name: "default_no_match", count: 2, percent: 13 },
]);

const latency = latencySeries(usage, 24, now);
assert.equal(latency.length, 24);
assert.deepEqual(
    latency.filter((point) => point.count > 0),
    [
        {
            bucket_start: "2026-07-04T10:00:00Z",
            label: "10:00",
            avg_latency_ms: 8.8,
            p95_latency_ms: 20,
            count: 10,
        },
        {
            bucket_start: "2026-07-04T11:00:00Z",
            label: "11:00",
            avg_latency_ms: 6,
            p95_latency_ms: 12,
            count: 4,
        },
    ],
);

// Latency broken down by source: p95 per source, zero-filled over the range.
const latencyBySource = latencySourceAreaSeries(usage, 24, 5, now);
assert.deepEqual(latencyBySource.series, [
    { key: "latency_0", label: "go-sdk" },
    { key: "latency_1", label: "js-sdk" },
]);
assert.deepEqual(latencyBySource.yDomain, [0, 25]);
assert.equal(latencyBySource.rows.length, 24);
assert.deepEqual(
    latencyBySource.rows.filter((row) => row.latency_0 > 0 || row.latency_1 > 0),
    [
        {
            bucket_start: "2026-07-04T10:00:00Z",
            label: "10:00",
            count: 0,
            latency_0: 20,
            latency_1: 9,
        },
        {
            bucket_start: "2026-07-04T11:00:00Z",
            label: "11:00",
            count: 0,
            latency_0: 0,
            latency_1: 12,
        },
    ],
);

const breakdown = trendBreakdownAreaSeries(usage, "flag_key", 24, 2, now);
assert.deepEqual(breakdown.series, [
    { key: "series_0", label: "search" },
    { key: "series_1", label: "checkout" },
]);
assert.deepEqual(breakdown.yDomain, [0, 10]);
assert.equal(breakdown.rows.length, 24);
assert.deepEqual(
    breakdown.rows.filter((row) => row.count > 0),
    [
        {
            bucket_start: "2026-07-04T10:00:00Z",
            label: "10:00",
            count: 10,
            series_0: 8,
            series_1: 2,
        },
        {
            bucket_start: "2026-07-04T11:00:00Z",
            label: "11:00",
            count: 4,
            series_0: 0,
            series_1: 4,
        },
    ],
);

assert.deepEqual(
    trendBreakdownAreaSeries(usage, "source", 24, 3, now).series.map(
        (series) => series.label,
    ),
    ["js-sdk", "go-sdk", "unknown"],
);

// API key series prefer the key's name over its id.
assert.deepEqual(
    trendBreakdownAreaSeries(usage, "api_key_id", 24, 2, now).series.map(
        (series) => series.label,
    ),
    ["search key", "unknown"],
);
