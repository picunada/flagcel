<script lang="ts">
    import type { AuditEntry } from "$lib/api";
    import { actionBadgeVariant } from "$lib/history";
    import Badge from "$lib/components/ui/badge.svelte";
    import Card from "$lib/components/ui/card.svelte";

    type HistoryItem = {
        entry: AuditEntry;
        changes: string[];
    };

    type Props = {
        historyView: HistoryItem[];
        formatTimestamp: (iso: string) => string;
    };

    let { historyView, formatTimestamp }: Props = $props();
</script>

{#if historyView.length === 0}
    <Card class="motion-panel p-8 text-center">
        <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
            [ no history ]
        </p>
        <p class="mt-3 text-sm text-foreground-softer">
            Changes to this flag and its rules will appear here.
        </p>
    </Card>
{:else}
    <div class="motion-list space-y-2">
        {#each historyView as { entry, changes } (entry.version)}
            <Card class="p-5">
                <div class="flex flex-wrap items-start justify-between gap-3">
                    <div class="flex items-center gap-3">
                        <span
                            class="font-mono text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground"
                        >
                            v{entry.version}
                        </span>
                        <Badge variant={actionBadgeVariant(entry.action)} dot
                            >{entry.action}</Badge
                        >
                    </div>
                    <div
                        class="flex flex-wrap items-center gap-x-3 gap-y-1 text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
                    >
                        <span class="font-mono text-foreground"
                            >{entry.actor_label ?? "system"}</span
                        >
                        <span>{formatTimestamp(entry.created_at)}</span>
                    </div>
                </div>
                <ul class="mt-3 space-y-1">
                    {#each changes as change}
                        <li class="flex gap-2 text-sm text-foreground-soft">
                            <span class="select-none text-muted-foreground"
                                >·</span
                            >
                            <span class="font-mono text-xs leading-relaxed"
                                >{change}</span
                            >
                        </li>
                    {/each}
                </ul>
            </Card>
        {/each}
    </div>
{/if}
