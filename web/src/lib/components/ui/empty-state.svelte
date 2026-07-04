<script lang="ts">
    import type { Snippet } from "svelte";
    import Card from "$lib/components/ui/card.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import { cn } from "$lib/utils";

    type Props = {
        icon?: any;
        title: string;
        description?: string;
        actionHref?: string;
        actionLabel?: string;
        class?: string;
        children?: Snippet;
    };

    let {
        icon: Icon,
        title,
        description,
        actionHref,
        actionLabel,
        class: className,
        children,
    }: Props = $props();
</script>

<Card
    class={cn(
        "motion-panel flex flex-col items-center gap-4 p-10 text-center",
        className,
    )}
>
    {#if Icon}
        <Icon class="h-5 w-5 text-muted-foreground" />
    {/if}
    <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
        {title}
    </p>
    {#if description}
        <p class="max-w-sm text-sm text-foreground-softer">
            {description}
        </p>
    {/if}
    {@render children?.()}
    {#if actionHref && actionLabel}
        <Button href={actionHref} class="mt-2">{actionLabel}</Button>
    {/if}
</Card>
