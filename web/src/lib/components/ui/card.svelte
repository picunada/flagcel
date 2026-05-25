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
                "hover:border-[rgba(255,255,255,0.36)] hover:bg-[rgba(255,255,255,0.04)] hover:shadow-[0_12px_40px_rgba(0,0,0,0.22)]",
            className,
        ),
    );
</script>

{#if hoverable}
    <BorderGlow
        class={cardClass}
        borderRadius={4}
        backgroundColor="rgba(15, 15, 15, 0.9)"
        edgeSensitivity={26}
        glowColor="var(--app-accent-hsl)"
        glowRadius={22}
        glowIntensity={0.45}
        coneSpread={5}
        colors={[
            "rgb(var(--app-accent-rgb) / 0.75)",
            "rgba(255,255,255,0.72)",
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
