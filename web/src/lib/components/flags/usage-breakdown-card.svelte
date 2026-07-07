<script lang="ts">
    import Card from "$lib/components/ui/card.svelte";

    type Item = {
        name: string;
        count: number;
    };

    type Props = {
        title: string;
        items: Item[];
        total: number;
    };

    let { title, items, total }: Props = $props();

    function percent(count: number) {
        if (total === 0) return "0%";
        return `${Math.round((count / total) * 100)}%`;
    }
</script>

<Card class="p-4">
    <p class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
        {title}
    </p>
    {#if items.length === 0}
        <p class="mt-3 text-sm text-foreground-softer">none</p>
    {:else}
        <div class="mt-3 space-y-2">
            {#each items.slice(0, 3) as item}
                <div class="grid grid-cols-[1fr_auto] items-center gap-3 text-xs">
                    <span class="truncate font-mono text-foreground">{item.name}</span>
                    <span class="text-muted-foreground">{percent(item.count)}</span>
                </div>
            {/each}
        </div>
    {/if}
</Card>
