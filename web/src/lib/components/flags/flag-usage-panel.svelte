<script lang="ts">
    import type { FlagUsage } from "$lib/api";
    import { formatFlagValue } from "$lib/values";
    import Card from "$lib/components/ui/card.svelte";
    import Badge from "$lib/components/ui/badge.svelte";
    import UsageAnalyticsCharts from "$lib/components/flags/usage-analytics-charts.svelte";
    import { usageTotal, type UsageRangeHours } from "$lib/usage-analytics";

    type Props = {
        usage: FlagUsage;
        range?: UsageRangeHours;
        formatTimestamp: (iso: string) => string;
    };

    let { usage, range = 24, formatTimestamp }: Props = $props();

    const total = $derived(usageTotal(usage));
</script>

{#if total === 0 && usage.events.length === 0}
    <Card class="motion-panel p-8 text-center">
        <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
            [ no usage ]
        </p>
        <p class="mt-3 text-sm text-foreground-softer">
            Evaluations will appear here after SDKs or API clients use this flag.
        </p>
    </Card>
{:else}
    <div class="space-y-4">
        <UsageAnalyticsCharts {usage} {range} variant="flag" />

        <Card class="overflow-hidden">
            <div class="border-b border-border px-5 py-4">
                <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
                    recent evaluations
                </p>
            </div>
            {#if usage.events.length === 0}
                <p class="p-5 text-sm text-foreground-softer">
                    Recent individual evaluations will appear here.
                </p>
            {:else}
                <div class="divide-y divide-border">
                    {#each usage.events as event (event.id)}
                        <div class="grid gap-3 px-5 py-4 md:grid-cols-[1fr_auto]">
                            <div class="min-w-0">
                                <div class="flex flex-wrap items-center gap-2">
                                    <Badge variant="muted">{event.reason}</Badge>
                                    {#if event.matched_rule_id}
                                        <span class="font-mono text-xs text-muted-foreground">
                                            {event.matched_rule_id}
                                        </span>
                                    {/if}
                                    <span class="font-mono text-xs text-foreground">
                                        {formatFlagValue(event.value)}
                                    </span>
                                </div>
                                <p class="mt-2 truncate text-xs text-muted-foreground">
                                    {event.source || "unknown source"}
                                    {#if event.api_key_id}
                                        · {event.api_key_id}
                                    {/if}
                                </p>
                            </div>
                            <div class="text-left md:text-right">
                                <p class="text-xs uppercase tracking-[0.12em] text-muted-foreground">
                                    {formatTimestamp(event.observed_at)}
                                </p>
                                <p class="mt-1 font-mono text-xs text-foreground-soft">
                                    {event.latency_ms.toFixed(2)} ms
                                </p>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </Card>
    </div>
{/if}
