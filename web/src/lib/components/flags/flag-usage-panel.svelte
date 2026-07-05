<script lang="ts">
    import type { FlagUsage } from "$lib/api";
    import { formatFlagValue } from "$lib/values";
    import Card from "$lib/components/ui/card.svelte";
    import Badge from "$lib/components/ui/badge.svelte";
    import UsageBreakdownCard from "$lib/components/flags/usage-breakdown-card.svelte";

    type Props = {
        usage: FlagUsage;
        formatTimestamp: (iso: string) => string;
    };

    let { usage, formatTimestamp }: Props = $props();

    const total = $derived(usage.buckets.reduce((sum, bucket) => sum + bucket.count, 0));
    const byValue = $derived(groupCounts(usage.buckets, (bucket) => formatFlagValue(bucket.value)));
    const byReason = $derived(groupCounts(usage.buckets, (bucket) => bucket.reason));
    const byRule = $derived(
        groupCounts(usage.buckets, (bucket) => bucket.matched_rule_id ?? "default")
    );

    function groupCounts<T>(items: T[], label: (item: T) => string) {
        const counts = new Map<string, number>();
        for (const item of items) {
            const key = label(item) || "unknown";
            counts.set(key, (counts.get(key) ?? 0) + (item as { count: number }).count);
        }
        return [...counts.entries()]
            .map(([name, count]) => ({ name, count }))
            .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name));
    }

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
        <div class="grid gap-3 lg:grid-cols-4">
            <Card class="p-4">
                <p class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
                    evaluations
                </p>
                <p class="mt-2 font-mono text-2xl">{total}</p>
            </Card>
            <UsageBreakdownCard title="values" items={byValue} {total} />
            <UsageBreakdownCard title="reasons" items={byReason} {total} />
            <UsageBreakdownCard title="rules" items={byRule} {total} />
        </div>

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
