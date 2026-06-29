<script lang="ts">
    import Button from "$lib/components/ui/button.svelte";
    import { Plus } from "lucide-svelte";

    type Props = {
        kind: "empty" | "filtered";
        activeFilters?: number;
        newFlagHref: string;
        onreset: () => void;
    };

    let { kind, activeFilters = 0, newFlagHref, onreset }: Props = $props();
</script>

<div class="flex flex-col items-center gap-4 p-12 text-center">
    {#if kind === "empty"}
        <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
            [ no flags yet ]
        </p>
        <p class="max-w-sm text-sm text-foreground-softer">
            Create your first flag to start routing evaluations.
        </p>
        <Button href={newFlagHref} class="mt-2">
            <Plus class="h-3.5 w-3.5" /> new flag
        </Button>
    {:else}
        <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
            [ no matches ]
        </p>
        <p class="max-w-sm text-sm text-foreground-softer">
            No flags match the current filters.
        </p>
        {#if activeFilters > 0}
            <Button variant="ghost" size="sm" onclick={onreset}
                >clear filters</Button
            >
        {/if}
    {/if}
</div>
