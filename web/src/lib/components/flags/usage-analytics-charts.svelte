<script lang="ts">
    import type { FlagUsage } from "$lib/api";
    import Card from "$lib/components/ui/card.svelte";
    import * as Chart from "$lib/components/ui/chart";
    import {
        breakdownSeries,
        evaluationTrendSeries,
        latencySeries,
        latencySourceAreaSeries,
        trendBreakdownAreaSeries,
        usageAxisFormat,
        usageAxisTickLabels,
        usageTotal,
        usageYDomain,
        type TrendBreakdownKey,
        type UsageRangeHours,
    } from "$lib/usage-analytics";
    import { curveMonotoneX } from "d3-shape";
    import { AreaChart } from "layerchart";
    import { Activity, Gauge, ListFilter } from "lucide-svelte";
    import type { Snippet } from "svelte";

    type Props = {
        usage: FlagUsage;
        range?: UsageRangeHours;
        variant?: "dashboard" | "flag";
    };

    const chartConfig = {
        evaluations: {
            label: "evaluations",
            color: "var(--color-app-accent)",
        },
        average: {
            label: "avg latency",
            unit: "ms",
            color: "var(--color-foreground-soft)",
        },
        p95: {
            label: "p95 latency",
            unit: "ms",
            color: "var(--color-destructive)",
        },
    } satisfies Chart.ChartConfig;

    type SegmentedOption = { value: string; label: string };

    const trendBreakdownOptions: SegmentedOption[] = [
        { value: "flag_key", label: "by flag" },
        { value: "source", label: "by source" },
        { value: "api_key_id", label: "by api key" },
    ];

    const latencyModeOptions: SegmentedOption[] = [
        { value: "overall", label: "overall" },
        { value: "source", label: "by source" },
    ];

    const trendAreaColors = [
        "var(--color-app-accent)",
        "#38bdf8",
        "#f59e0b",
        "#a78bfa",
        "#ff6b6b",
    ];

    let { usage, range = 24, variant = "flag" }: Props = $props();
    let trendBreakdown = $state<TrendBreakdownKey>("flag_key");
    let latencyMode = $state<"overall" | "source">("overall");

    function setTrendBreakdown(value: string) {
        trendBreakdown = value as TrendBreakdownKey;
    }

    function setLatencyMode(value: string) {
        latencyMode = value as "overall" | "source";
    }

    const total = $derived(usageTotal(usage));
    const trend = $derived(evaluationTrendSeries(usage, range));
    const latency = $derived(latencySeries(usage, range));
    const sourceCount = $derived(breakdownSeries(usage, "source").length);
    const trendArea = $derived(
        trendBreakdownAreaSeries(usage, trendBreakdown, range),
    );
    const trendAreaSeries = $derived(
        trendArea.series.map((series, index) => ({
            ...series,
            color: trendAreaColors[index % trendAreaColors.length],
        })),
    );
    const xAxisProps = $derived({
        ticks: usageAxisTickLabels(trend, range),
        format: usageAxisFormat(range),
    });
    const trendYDomain = $derived(usageYDomain(trend.map((point) => point.count)));
    const latencyYDomain = $derived(
        usageYDomain(latency.map((point) => point.p95_latency_ms)),
    );
    const latencySource = $derived(latencySourceAreaSeries(usage, range));
    const latencySourceSeries = $derived(
        latencySource.series.map((series, index) => ({
            ...series,
            color: trendAreaColors[index % trendAreaColors.length],
        })),
    );
    // Dynamic per-source entries so the tooltip shows source names with ms units.
    const config = $derived({
        ...chartConfig,
        ...Object.fromEntries(
            latencySource.series.map((series) => [
                series.key,
                { label: series.label, unit: "ms" },
            ]),
        ),
    } satisfies Chart.ChartConfig);
    const trendHasData = $derived(
        variant === "dashboard"
            ? trendArea.series.length > 0
            : usage.buckets.length > 0,
    );
    const latencyHasData = $derived(usage.latency_buckets.length > 0);
    const lastLatency = $derived(
        [...latency].reverse().find((point) => point.count > 0),
    );
    const p95Label = $derived(
        latencyHasData && lastLatency
            ? `${lastLatency.p95_latency_ms} ms`
            : "n/a",
    );
</script>

{#if total === 0}
    <Card class="motion-panel p-8 text-center">
        <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
            [ no usage ]
        </p>
        <p class="mt-3 text-sm text-foreground-softer">
            Evaluations will appear here after SDKs or API clients use flags.
        </p>
    </Card>
{:else}
    <div class="space-y-4">
        <div class="grid gap-3 lg:grid-cols-3">
            {@render MetricCard(
                "evaluations",
                total.toLocaleString(),
                Activity,
            )}
            {@render MetricCard(
                "active sources",
                sourceCount.toLocaleString(),
                ListFilter,
            )}
            {@render MetricCard("p95 latency", p95Label, Gauge)}
        </div>

        <div class="grid gap-4 xl:grid-cols-2">
            {@render ChartCard(
                "evaluation trend",
                trendHasData,
                TrendChart,
                variant === "dashboard" ? TrendBreakdownPicker : undefined,
            )}
            {@render ChartCard(
                "latency",
                latencyHasData,
                LatencyChart,
                LatencyModePicker,
            )}
        </div>
    </div>
{/if}

{#snippet ChartTooltip()}
    <Chart.Tooltip />
{/snippet}

{#snippet MetricCard(title: string, value: string, icon: typeof Activity)}
    {@const Icon = icon}
    <Card class="p-4">
        <div class="flex items-center justify-between gap-3">
            <p
                class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
            >
                {title}
            </p>
            <Icon class="size-3.5 text-muted-foreground" />
        </div>
        <p class="mt-2 font-mono text-2xl">{value}</p>
    </Card>
{/snippet}

{#snippet ChartCard(
    title: string,
    hasData: boolean,
    chart: Snippet,
    controls?: Snippet,
)}
    <Card class="p-4">
        <div class="flex min-h-9 flex-wrap items-center justify-between gap-3">
            <p
                class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
            >
                {title}
            </p>
            {#if controls}
                {@render controls()}
            {/if}
        </div>
        {#if hasData}
            <div class="mt-3">
                <Chart.Container {config} class="h-64 aspect-auto">
                    {@render chart()}
                </Chart.Container>
            </div>
        {:else}
            <p class="mt-6 text-sm text-foreground-softer">none</p>
        {/if}
    </Card>
{/snippet}

{#snippet SegmentedControl(
    label: string,
    options: SegmentedOption[],
    selected: string,
    onselect: (value: string) => void,
)}
    <div
        class="ios-corners-sm inline-flex h-9 shrink-0 border border-border p-0.5"
        aria-label={label}
    >
        {#each options as option (option.value)}
            <button
                type="button"
                aria-pressed={selected === option.value}
                class={[
                    "ios-corners-xs h-full px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors",
                    selected === option.value
                        ? "bg-surface-selected text-foreground"
                        : "text-muted-foreground hover:text-foreground",
                ]}
                onclick={() => onselect(option.value)}
            >
                {option.label}
            </button>
        {/each}
    </div>
{/snippet}

{#snippet TrendBreakdownPicker()}
    {@render SegmentedControl(
        "Trend breakdown",
        trendBreakdownOptions,
        trendBreakdown,
        setTrendBreakdown,
    )}
{/snippet}

{#snippet LatencyModePicker()}
    {@render SegmentedControl(
        "Latency breakdown",
        latencyModeOptions,
        latencyMode,
        setLatencyMode,
    )}
{/snippet}

{#snippet TrendChart()}
    {#if variant === "dashboard"}
        <AreaChart
            data={trendArea.rows}
            x="label"
            yDomain={trendArea.yDomain}
            axis="x"
            grid
            series={trendAreaSeries}
            seriesLayout="overlap"
            tooltip={ChartTooltip}
            props={{
                area: { curve: curveMonotoneX, line: true },
                xAxis: xAxisProps,
            }}
        />
    {:else}
        <AreaChart
            data={trend}
            x="label"
            yDomain={trendYDomain}
            axis="x"
            grid
            tooltip={ChartTooltip}
            props={{
                area: { curve: curveMonotoneX, line: true },
                xAxis: xAxisProps,
            }}
            series={[
                {
                    key: "evaluations",
                    label: "evaluations",
                    value: "count",
                    color: "var(--color-evaluations)",
                },
            ]}
        />
    {/if}
{/snippet}

{#snippet LatencyChart()}
    {#if latencyMode === "source"}
        <AreaChart
            data={latencySource.rows}
            x="label"
            yDomain={latencySource.yDomain}
            axis="x"
            grid
            tooltip={ChartTooltip}
            seriesLayout="overlap"
            props={{
                area: { curve: curveMonotoneX, line: true },
                xAxis: xAxisProps,
            }}
            series={latencySourceSeries}
        />
    {:else}
        <AreaChart
            data={latency}
            x="label"
            yDomain={latencyYDomain}
            axis="x"
            grid
            tooltip={ChartTooltip}
            seriesLayout="overlap"
            props={{
                area: { curve: curveMonotoneX, line: true },
                xAxis: xAxisProps,
            }}
            series={[
                {
                    key: "average",
                    label: "avg",
                    value: "avg_latency_ms",
                    color: "var(--color-average)",
                },
                {
                    key: "p95",
                    label: "p95",
                    value: "p95_latency_ms",
                    color: "var(--color-p95)",
                },
            ]}
        />
    {/if}
{/snippet}
