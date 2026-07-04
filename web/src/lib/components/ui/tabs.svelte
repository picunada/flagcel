<script lang="ts" generics="T extends string">
    import { cn } from "$lib/utils";

    type TabItem<T extends string = string> = {
        value: T;
        label: string;
        count?: number;
        icon?: any;
    };

    let {
        items,
        value = $bindable<T>(),
        class: className,
    }: { items: TabItem<T>[]; value: T; class?: string } = $props();
</script>

<div class={cn("flex items-center gap-5", className)}>
    {#each items as item (item.value)}
        {@const Icon = item.icon}
        <button
            type="button"
            onclick={() => (value = item.value)}
            class={cn(
                "-mb-px inline-flex items-center gap-1.5 border-b-2 pb-2 text-[0.7rem] uppercase tracking-[0.18em] transition-colors",
                value === item.value
                    ? "border-success text-foreground"
                    : "border-transparent text-muted-foreground hover:text-foreground",
            )}
        >
            {#if Icon}
                <Icon class="h-3 w-3" />
            {/if}
            {item.label}
            {#if item.count !== undefined && item.count > 0}
                <span class="text-muted-foreground">({item.count})</span>
            {/if}
        </button>
    {/each}
</div>
