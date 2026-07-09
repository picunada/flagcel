<script lang="ts">
	import type { AuditEntry, FlagUsage } from '$lib/api';
	import Card from '$lib/components/ui/card.svelte';
	import { evaluationTrendSeries, usageTotal, type UsageRangeHours } from '$lib/usage-analytics';

	type HistoryItem = {
		entry: AuditEntry;
		changes: string[];
	};

	type SparkBar = {
		key: string;
		height: string;
		active: boolean;
		count: number;
	};

	type Props = {
		historyView: HistoryItem[];
		usage: FlagUsage;
		range?: UsageRangeHours;
		formatTimestamp: (iso: string) => string;
	};

	let { historyView, usage, range = 24, formatTimestamp }: Props = $props();

	const total = $derived(usageTotal(usage));
	const trend = $derived(evaluationTrendSeries(usage, range));
	const sparkBars = $derived.by<SparkBar[]>(() => {
		const points = trend.slice(-18);
		const max = Math.max(1, ...points.map((point) => point.count));
		return points.map((point) => ({
			key: point.bucket_start,
			height: `${Math.max(4, Math.round((point.count / max) * 32))}px`,
			active: point.count > 0,
			count: point.count
		}));
	});

	const latestEvent = $derived(usage.events[0] ?? null);
</script>

{#if total === 0 && historyView.length === 0}
	<Card class="motion-panel p-8 text-center">
		<p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">[ no activity ]</p>
		<p class="mt-3 text-sm text-foreground-softer">
			Usage and flag changes will appear here.
		</p>
	</Card>
{:else}
	<Card class="motion-panel overflow-hidden p-0">
		<div class="border-b border-border px-4 py-4 sm:px-5">
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div>
					<p class="text-[0.65rem] uppercase tracking-[0.16em] text-muted-foreground">
						usage · {range === 24 ? '24h' : range === 168 ? '7d' : '30d'}
					</p>
					<div class="mt-2 flex items-baseline gap-2">
						<span class="font-mono text-2xl font-semibold text-foreground">{total}</span>
						<span class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
							evaluations
						</span>
					</div>
					{#if latestEvent}
						<p class="mt-1 truncate text-xs text-muted-foreground">
							latest · {formatTimestamp(latestEvent.observed_at)} · {latestEvent.source || 'unknown source'}
						</p>
					{:else}
						<p class="mt-1 text-xs text-muted-foreground">no recent individual evaluations</p>
					{/if}
				</div>

				<div class="flex h-12 items-end gap-1" aria-label="evaluation trend sparkline">
					{#each sparkBars as bar (bar.key)}
						<div
							title={`${bar.count} evaluations`}
							class={bar.active ? 'w-1.5 rounded-sm bg-foreground' : 'w-1.5 rounded-sm bg-surface-selected'}
							style:height={bar.height}
						></div>
					{/each}
				</div>
			</div>
		</div>

		<div class="px-4 py-4 sm:px-5">
			<p class="mb-2 text-[0.65rem] uppercase tracking-[0.16em] text-muted-foreground">
				activity · {historyView.length} {historyView.length === 1 ? 'version' : 'versions'}
			</p>

			{#if historyView.length === 0}
				<p class="py-4 text-sm text-foreground-softer">
					Changes to this flag and its rules will appear here.
				</p>
			{:else}
				<div class="divide-y divide-border">
					{#each historyView as { entry, changes } (entry.version)}
						<div class="flex gap-3 py-3">
							<span
								class="w-9 shrink-0 pt-0.5 font-mono text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
							>
								v{entry.version}
							</span>
							<div class="min-w-0 flex-1">
								<div class="flex flex-wrap items-baseline gap-x-2 gap-y-1">
									<span class="font-mono text-xs uppercase tracking-[0.12em] text-foreground">
										{entry.action}
									</span>
									<span class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
										{entry.actor_label ?? 'system'} · {formatTimestamp(entry.created_at)}
									</span>
								</div>
								<div class="mt-1 space-y-1">
									{#each changes.slice(0, 3) as change}
										<p class="truncate font-mono text-xs leading-relaxed text-foreground-soft">
											{change}
										</p>
									{/each}
									{#if changes.length > 3}
										<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
											+{changes.length - 3} more
										</p>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</Card>
{/if}
