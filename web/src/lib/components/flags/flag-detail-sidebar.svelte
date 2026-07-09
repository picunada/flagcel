<script lang="ts">
	import type { AuditEntry, EvalTrace, FlagUsage } from '$lib/api';
	import EvalPlayground from '$lib/components/eval-playground.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import * as Select from '$lib/components/ui/select';
	import {
		evaluationTrendSeries,
		usageRangeOptions,
		usageTotal,
		type UsageRangeHours
	} from '$lib/usage-analytics';
	import { Activity, FlaskConical } from 'lucide-svelte';

	type HistoryItem = {
		entry: AuditEntry;
		changes: string[];
	};

	type Props = {
		historyView: HistoryItem[];
		usage: FlagUsage;
		range: UsageRangeHours;
		formatTimestamp: (iso: string) => string;
		showFullUsage: () => void;
		onRangeChange: (range: UsageRangeHours) => void;
		playgroundContext: string;
		playgroundResult: EvalTrace | null;
		playgroundError: string | null;
		playgroundRunning: boolean;
		evaluatePlayground: () => void | Promise<void>;
		resetPlayground: () => void;
		markDirty: () => void;
	};

	let {
		historyView,
		usage,
		range,
		formatTimestamp,
		showFullUsage,
		onRangeChange,
		playgroundContext = $bindable(),
		playgroundResult,
		playgroundError,
		playgroundRunning,
		evaluatePlayground,
		resetPlayground,
		markDirty
	}: Props = $props();

	let tab = $state<'activity' | 'test'>('activity');

	const total = $derived(usageTotal(usage));
	// Keep the full range so 7d/30d spikes outside the most recent few buckets still render.
	const trend = $derived(evaluationTrendSeries(usage, range));
	const maxTrend = $derived(Math.max(1, ...trend.map((point) => point.count)));
	const sourceCount = $derived(
		new Set(usage.buckets.map((bucket) => bucket.source || 'unknown')).size
	);
	const p95 = $derived.by(() => {
		const buckets = usage.latency_buckets;
		if (!buckets.length) return 0;
		return Math.max(...buckets.map((bucket) => bucket.p95_latency_ms));
	});
	const rangeLabel = $derived(
		usageRangeOptions.find((option) => option.value === range)?.label ?? '24h'
	);
</script>

<Card class="sticky top-6 overflow-hidden p-0 lg:max-h-[calc(100vh-3rem)] lg:overflow-y-auto">
	<div class="grid grid-cols-2 border-b border-border/60">
		<button
			type="button"
			onclick={() => (tab = 'activity')}
			class="inline-flex cursor-pointer items-center justify-center gap-1.5 border-b-2 px-4 py-3 font-mono text-[0.65rem] uppercase tracking-[0.16em] transition-colors {tab ===
			'activity'
				? 'border-foreground text-foreground'
				: 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			<Activity class="h-3 w-3" />
			activity
		</button>
		<button
			type="button"
			onclick={() => (tab = 'test')}
			class="inline-flex cursor-pointer items-center justify-center gap-1.5 border-b-2 px-4 py-3 font-mono text-[0.65rem] uppercase tracking-[0.16em] transition-colors {tab ===
			'test'
				? 'border-foreground text-foreground'
				: 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			<FlaskConical class="h-3 w-3" />
			test
		</button>
	</div>

	{#if tab === 'activity'}
		<div class="space-y-5 p-4">
			<div class="flex items-center justify-between gap-3">
				<div class="flex items-center gap-1.5 font-mono text-[0.65rem] uppercase tracking-[0.16em] text-muted-foreground">
					<span>pulse</span>
					<span aria-hidden="true">·</span>
					<Select.Root
						type="single"
						value={String(range)}
						onValueChange={(next) => {
							const parsed = Number(next);
							if (parsed === 24 || parsed === 168 || parsed === 720) {
								onRangeChange(parsed);
							}
						}}
					>
						<Select.Trigger
							aria-label="Pulse range"
							class="ios-corners-xs h-auto w-auto min-w-0 cursor-pointer gap-1 border-0 bg-transparent px-1 py-0.5 text-[0.65rem] uppercase tracking-[0.16em] text-foreground outline-none transition-colors hover:bg-surface-hover hover:text-foreground focus-visible:ring-1 focus-visible:ring-ring"
						>
							<span>{rangeLabel}</span>
						</Select.Trigger>
						<Select.Content align="start" class="min-w-24">
							{#each usageRangeOptions as option (option.value)}
								<Select.Item value={String(option.value)} label={option.label}>
									<span class="font-mono text-xs uppercase tracking-[0.12em]">{option.label}</span>
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
				<button
					type="button"
					onclick={showFullUsage}
					class="cursor-pointer font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground transition-colors hover:text-foreground"
				>
					full usage →
				</button>
			</div>

			<div class="grid grid-cols-3 gap-3">
				<div>
					<p class="font-mono text-lg text-foreground">{total}</p>
					<p class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">evals</p>
				</div>
				<div>
					<p class="font-mono text-lg text-foreground">{sourceCount}</p>
					<p class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">sources</p>
				</div>
				<div>
					<p class="font-mono text-lg text-foreground">{p95.toFixed(2)}ms</p>
					<p class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">p95</p>
				</div>
			</div>

			<div class="flex h-10 items-end gap-1" aria-label="evaluation trend sparkline">
				{#each trend as point (point.bucket_start)}
					<div
						title={`${point.count} evaluations`}
						class={point.count > 0 ? 'flex-1 rounded-sm bg-foreground' : 'flex-1 rounded-sm bg-surface-selected'}
						style:height={`${Math.max(4, Math.round((point.count / maxTrend) * 32))}px`}
					></div>
				{/each}
			</div>

			<div>
				<p class="mb-3 font-mono text-[0.65rem] uppercase tracking-[0.16em] text-muted-foreground">
					activity · {historyView.length} {historyView.length === 1 ? 'version' : 'versions'}
				</p>
				{#if historyView.length === 0}
					<p class="py-2 text-sm text-foreground-softer">Changes will appear here.</p>
				{:else}
					<div class="divide-y divide-border/60">
						{#each historyView.slice(0, 5) as { entry, changes } (entry.version)}
							<div class="flex gap-3 py-3">
								<span class="w-7 shrink-0 font-mono text-[0.65rem] uppercase text-muted-foreground">
									v{entry.version}
								</span>
								<div class="min-w-0 flex-1">
									<p class="truncate font-mono text-xs text-foreground">
										{changes[0] ?? entry.action}
									</p>
									<p class="mt-1 truncate text-[0.65rem] uppercase tracking-[0.08em] text-muted-foreground">
										{formatTimestamp(entry.created_at)} · {entry.actor_label ?? 'system'}
									</p>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<div class="p-4">
			<EvalPlayground
				inputId="playground-context-sidebar"
				bind:contextJson={playgroundContext}
				result={playgroundResult}
				error={playgroundError}
				running={playgroundRunning}
				onevaluate={evaluatePlayground}
				onreset={resetPlayground}
				oninput={markDirty}
			/>
		</div>
	{/if}
</Card>
