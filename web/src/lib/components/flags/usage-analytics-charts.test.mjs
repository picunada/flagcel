import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";

const source = await readFile(
    new URL("./usage-analytics-charts.svelte", import.meta.url),
    "utf8",
);
const chartTooltipSource = await readFile(
    new URL("../ui/chart/chart-tooltip.svelte", import.meta.url),
    "utf8",
);

assert.equal(
    source.includes("{@render SeriesTable"),
    false,
    "usage analytics charts should not render footer breakdown rows",
);

assert.equal(
    source.includes("{#snippet SeriesTable"),
    false,
    "usage analytics charts should not keep the footer breakdown snippet",
);

assert.equal(
    source.includes("AreaChart"),
    true,
    "dashboard evaluation trend should render an area chart",
);

assert.equal(
    source.includes("by flag") && source.includes("by source") && source.includes("by api key"),
    true,
    "dashboard evaluation trend should expose breakdown selector labels",
);

assert.equal(
    source.includes('seriesLayout="overlap"') && source.includes("line: true"),
    true,
    "dashboard evaluation trend should draw overlapping area series with visible boundary lines",
);

assert.equal(
    source.includes("yDomain={trendArea.yDomain}") && !source.includes("yNice\n                                    axis=\"x\""),
    true,
    "dashboard evaluation trend should use padded y-domain without automatic nicening",
);

assert.equal(
    source.includes("{#snippet ChartTooltip()}") &&
        source.includes("<Chart.Tooltip />") &&
        source.includes("tooltip={ChartTooltip}"),
    true,
    "usage analytics charts should render the shared chart tooltip component",
);

assert.equal(
    source.includes("chartTooltipProps") ||
        source.includes("tooltip: chartTooltipProps"),
    false,
    "usage analytics charts should not pass tooltip style props into LayerChart",
);

assert.equal(
    chartTooltipSource.includes("bg-surface-subtle") &&
        chartTooltipSource.includes("backdrop-blur-xl") &&
        chartTooltipSource.includes("border-border") &&
        chartTooltipSource.includes("text-foreground-soft"),
    true,
    "shared chart tooltip should use the app tooltip glass surface classes",
);

assert.equal(
    chartTooltipSource.includes("motion-tooltip") ||
        chartTooltipSource.includes("backdrop-filter: blur"),
    false,
    "shared chart tooltip should not carry inline backdrop-filter hacks or bits-ui motion classes",
);
