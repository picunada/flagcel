import { describe, expect, it } from "vitest";
import source from "./usage-analytics-charts.svelte?raw";
import chartTooltipSource from "../ui/chart/chart-tooltip.svelte?raw";

describe("usage analytics charts source", () => {
	it("keeps the expected chart markup contracts", () => {
		expect(source.includes("{@render SeriesTable")).toBe(false);
		expect(source.includes("{#snippet SeriesTable")).toBe(false);
		expect(source.includes("AreaChart")).toBe(true);
		expect(
			source.includes("by flag") &&
				source.includes("by source") &&
				source.includes("by api key"),
		).toBe(true);
		expect(
			source.includes('seriesLayout="overlap"') && source.includes("line: true"),
		).toBe(true);
		expect(
			source.includes("yDomain={trendArea.yDomain}") &&
				!source.includes("yNice\n                                    axis=\"x\""),
		).toBe(true);
		expect(
			source.includes("{#snippet ChartTooltip()}") &&
				source.includes("<Chart.Tooltip />") &&
				source.includes("tooltip={ChartTooltip}"),
		).toBe(true);
		expect(
			source.includes("chartTooltipProps") ||
				source.includes("tooltip: chartTooltipProps"),
		).toBe(false);
		expect(
			chartTooltipSource.includes("bg-surface-subtle") &&
				chartTooltipSource.includes("backdrop-blur-xl") &&
				chartTooltipSource.includes("border-border") &&
				chartTooltipSource.includes("text-foreground-soft"),
		).toBe(true);
		expect(
			chartTooltipSource.includes("motion-tooltip") ||
				chartTooltipSource.includes("backdrop-filter: blur"),
		).toBe(false);
	});
});
