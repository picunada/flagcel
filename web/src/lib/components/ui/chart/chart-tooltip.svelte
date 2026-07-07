<script lang="ts">
	import { cn, type WithElementRef, type WithoutChildren } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";
	import { getPayloadConfigFromPayload, useChart, type TooltipPayload } from "./chart-utils.js";
	import { getChartContext, Tooltip as TooltipPrimitive } from "layerchart";

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: WithoutChildren<WithElementRef<HTMLAttributes<HTMLDivElement>>> = $props();

	const chart = useChart();
	const chartCtx = getChartContext();

	// Firefox cannot render a backdrop-filter nested inside another backdrop-filter
	// (the glass-panel card), so hoist the layerchart tooltip root to <body> — same
	// escape hatch bits-ui tooltips use via their portal. A fixed wrapper tracks the
	// chart's viewport position so layerchart's container-relative top/left keep working.
	function portalTooltip(node: HTMLElement) {
		const root = node.parentElement;
		const anchor = root?.offsetParent as HTMLElement | null;
		if (!root || !anchor) return;

		const wrapper = document.createElement("div");
		wrapper.style.position = "fixed";
		wrapper.style.top = "0";
		wrapper.style.left = "0";
		wrapper.style.zIndex = "50";
		document.body.appendChild(wrapper);
		wrapper.appendChild(root);

		let raf = 0;
		const sync = () => {
			const rect = anchor.getBoundingClientRect();
			wrapper.style.transform = `translate(${rect.left}px, ${rect.top}px)`;
			raf = requestAnimationFrame(sync);
		};
		sync();

		return {
			destroy() {
				cancelAnimationFrame(raf);
				wrapper.remove();
			},
		};
	}

	// Filter to series with defined values (important for item-based charts like Pie/Arc
	// where only the hovered item has a value)
	const visibleSeries = $derived(
		chartCtx.tooltip.series.filter((s: TooltipPayload) => s.value !== undefined)
	);

	// The x-axis value for the hovered datum (e.g. a date or bucket label)
	const label = $derived.by(() => {
		if (!visibleSeries.length) return null;
		const tooltipData = chartCtx.tooltip.data;
		return tooltipData != null ? chartCtx.x(tooltipData) : null;
	});
</script>

<TooltipPrimitive.Root variant="none">
	<div
		bind:this={ref}
		use:portalTooltip
		class={cn(
			"border-border bg-surface-subtle text-foreground-soft z-50 grid w-fit min-w-[9rem] max-w-xs items-start gap-1.5 rounded-md border px-3 py-1.5 text-xs backdrop-blur-xl",
			className
		)}
		{...restProps}
	>
		{#if label != null}
			<div class="font-medium">{label}</div>
		{/if}
		<div class="grid gap-1.5">
			{#each visibleSeries as item, i (item.key + i)}
				{@const key = `${item.key || item.label || "value"}`}
				{@const itemConfig = getPayloadConfigFromPayload(
					chart.config,
					item,
					key,
					chartCtx.tooltip.data
				)}
				{@const indicatorColor =
					// The tooltip is portaled to <body>, so chart-scoped generated vars
					// (--color-<key>) don't resolve here; the config holds the raw value.
					itemConfig?.color || item.config?.color || item.color}
				<div class="flex w-full flex-wrap items-center gap-2">
					<div
						style="background-color: {indicatorColor};"
						class="size-2.5 shrink-0 rounded-[2px]"
					></div>
					<div class="flex flex-1 shrink-0 items-center justify-between gap-3 leading-none">
						<span class="text-muted-foreground">{itemConfig?.label || item.label}</span>
						{#if item.value !== undefined}
							<span class="text-foreground font-mono font-medium tabular-nums">
								{item.value.toLocaleString()}{itemConfig?.unit
									? ` ${itemConfig.unit}`
									: ""}
							</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	</div>
</TooltipPrimitive.Root>
