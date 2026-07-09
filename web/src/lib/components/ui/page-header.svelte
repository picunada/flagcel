<script lang="ts">
    import type { Snippet } from "svelte";
    import { cn } from "$lib/utils";

    type Props = {
        eyebrow?: string;
        title?: string;
        description?: string;
        class?: string;
        titleClass?: string;
        actions?: Snippet;
    };

    let {
        eyebrow,
        title,
        description,
        class: className,
        titleClass,
        actions,
    }: Props = $props();
</script>

<header
    class={cn(
        "flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between",
        className,
    )}
>
    <div class="min-w-0 space-y-3">
        {#if eyebrow}
            <p class="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                {eyebrow}
            </p>
        {/if}
        {#if title}
            <h1
                class={cn(
                    "text-balance text-3xl font-normal leading-tight sm:text-4xl",
                    titleClass,
                )}
            >
                {title}
            </h1>
        {/if}
        {#if description}
            <p class="max-w-xl text-sm text-foreground-soft sm:text-base">
                {description}
            </p>
        {/if}
    </div>
    {#if actions}
        <div class="flex shrink-0 items-center gap-2">
            {@render actions()}
        </div>
    {/if}
</header>
