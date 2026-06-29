<script lang="ts">
    import type { Snippet } from "svelte";
    import { cn } from "$lib/utils";
    import BorderGlow from "$lib/components/svelte-bits/border-glow.svelte";

    let {
        class: className,
        hoverable = false,
        children,
    }: { class?: string; hoverable?: boolean; children?: Snippet } = $props();

    const cardClass = $derived(
        cn(
            "glass-panel rounded-sm transition-all duration-200 ease-out",
            hoverable &&
                "hover:border-border-hover hover:bg-surface-hover hover:shadow-card-hover",
            className,
        ),
    );
</script>

{#if hoverable}
    <BorderGlow
        class={cardClass}
        borderRadius={4}
        backgroundColor="var(--color-card)"
        edgeSensitivity={26}
        glowColor="var(--app-accent-hsl)"
        glowRadius={22}
        glowIntensity={0.45}
        coneSpread={5}
        colors={[
            "rgb(var(--app-accent-rgb) / 0.75)",
            "var(--color-foreground-soft)",
            "rgb(var(--app-accent-rgb) / 0.45)",
        ]}
        fillOpacity={0}
        edgeFillOpacity={0.2}
        insetGlow={false}
    >
        {@render children?.()}
    </BorderGlow>
{:else}
    <div class={cardClass}>
        {@render children?.()}
    </div>
{/if}
