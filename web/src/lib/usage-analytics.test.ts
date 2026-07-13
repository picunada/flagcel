import { describe, expect, it } from "vitest";
import type { FlagUsage } from "$lib/api";
import {
	breakdownSeries,
	evaluationTrendSeries,
	latencySeries,
	latencySourceAreaSeries,
	trendBreakdownAreaSeries,
	usageAxisFormat,
	usageAxisTickLabels,
} from "./usage-analytics";

const now = new Date("2026-07-04T12:30:00Z");

const usage: FlagUsage = {
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

describe("usage analytics", () => {
	it("zero-fills 24h hourly trend buckets", () => {
		const trend = evaluationTrendSeries(usage, 24, now);
		expect(trend).toHaveLength(24);
		expect(trend[0].bucket_start).toBe("2026-07-03T13:00:00Z");
		expect(trend.at(-1)?.bucket_start).toBe("2026-07-04T12:00:00Z");
		expect(trend.every((point) => point.count >= 0)).toBe(true);
		expect(trend.filter((point) => point.count > 0)).toEqual([
			{ bucket_start: "2026-07-04T10:00:00Z", label: "10:00", count: 10 },
			{ bucket_start: "2026-07-04T11:00:00Z", label: "11:00", count: 4 },
			{ bucket_start: "2026-07-04T12:00:00Z", label: "12:00", count: 1 },
		]);
	});

	it("uses 6h buckets for 7d trends", () => {
		const weekTrend = evaluationTrendSeries(usage, 168, now);
		expect(weekTrend).toHaveLength(28);
		expect(weekTrend.filter((point) => point.count > 0)).toEqual([
			{
				bucket_start: "2026-07-04T06:00:00Z",
				label: "Jul 4, 06:00",
				count: 14,
			},
			{
				bucket_start: "2026-07-04T12:00:00Z",
				label: "Jul 4, 12:00",
				count: 1,
			},
		]);
	});

	it("uses daily buckets for 30d trends", () => {
		const monthTrend = evaluationTrendSeries(usage, 720, now);
		expect(monthTrend).toHaveLength(30);
		expect(monthTrend.filter((point) => point.count > 0)).toEqual([
			{ bucket_start: "2026-07-04T00:00:00Z", label: "Jul 4", count: 15 },
		]);
	});

	it("formats 7d axis ticks as day-only labels", () => {
		const weekTrend = evaluationTrendSeries(usage, 168, now);
		const weekTicks = usageAxisTickLabels(weekTrend, 168);
		expect(weekTicks).toHaveLength(7);
		expect(weekTicks[0]).toBe("Jun 28, 00:00");
		expect(usageAxisFormat(168)?.(weekTicks.at(-1)!)).toBe("Jul 4");
		expect(usageAxisFormat(24)).toBeUndefined();
		expect(
			usageAxisTickLabels(evaluationTrendSeries(usage, 24, now), 24),
		).toHaveLength(8);
	});

	it("builds reason breakdown series", () => {
		expect(breakdownSeries(usage, "reason")).toEqual([
			{ name: "matched_rule", count: 13, percent: 87 },
			{ name: "default_no_match", count: 2, percent: 13 },
		]);
	});

	it("aggregates latency over the selected range", () => {
		const latency = latencySeries(usage, 24, now);
		expect(latency).toHaveLength(24);
		expect(latency.filter((point) => point.count > 0)).toEqual([
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
		]);
	});

	it("breaks latency down by source", () => {
		const latencyBySource = latencySourceAreaSeries(usage, 24, 5, now);
		expect(latencyBySource.series).toEqual([
			{ key: "latency_0", label: "go-sdk" },
			{ key: "latency_1", label: "js-sdk" },
		]);
		expect(latencyBySource.yDomain).toEqual([0, 25]);
		expect(latencyBySource.rows).toHaveLength(24);
		expect(
			latencyBySource.rows.filter(
				(row) => Number(row.latency_0) > 0 || Number(row.latency_1) > 0,
			),
		).toEqual([
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
		]);
	});

	it("builds stacked trend breakdowns by flag", () => {
		const breakdown = trendBreakdownAreaSeries(usage, "flag_key", 24, 2, now);
		expect(breakdown.series).toEqual([
			{ key: "series_0", label: "search" },
			{ key: "series_1", label: "checkout" },
		]);
		expect(breakdown.yDomain).toEqual([0, 10]);
		expect(breakdown.rows).toHaveLength(24);
		expect(breakdown.rows.filter((row) => row.count > 0)).toEqual([
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
		]);
	});

	it("labels source and api-key breakdown series", () => {
		expect(
			trendBreakdownAreaSeries(usage, "source", 24, 3, now).series.map(
				(series) => series.label,
			),
		).toEqual(["js-sdk", "go-sdk", "unknown"]);
		expect(
			trendBreakdownAreaSeries(usage, "api_key_id", 24, 2, now).series.map(
				(series) => series.label,
			),
		).toEqual(["search key", "unknown"]);
	});
});
