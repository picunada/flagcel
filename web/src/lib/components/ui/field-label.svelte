<script lang="ts">
    import type { Snippet } from "svelte";
    import { cn } from "$lib/utils";

    let {
        for: forId,
        class: className,
        children,
    }: { for?: string; class?: string; children?: Snippet } = $props();
</script>

<label
    for={forId}
    class={cn(
        "text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground",
        className,
    )}
>
    {@render children?.()}
</label>
