<script lang="ts">
    import type { EvalTrace } from "$lib/api";
    import EvalPlayground from "$lib/components/eval-playground.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import { ChevronDown, X } from "lucide-svelte";
    import { fly, slide } from "svelte/transition";

    type Props = {
        drawerOpen: boolean;
        mobileOpen: boolean;
        playgroundContext: string;
        playgroundResult: EvalTrace | null;
        playgroundError: string | null;
        playgroundRunning: boolean;
        evaluatePlayground: () => void | Promise<void>;
        resetPlayground: () => void;
        markDirty: () => void;
    };

    let {
        drawerOpen = $bindable(),
        mobileOpen = $bindable(),
        playgroundContext = $bindable(),
        playgroundResult,
        playgroundError,
        playgroundRunning,
        evaluatePlayground,
        resetPlayground,
        markDirty,
    }: Props = $props();
</script>

<div class="lg:hidden">
    <Card class="motion-panel overflow-hidden">
        <button
            type="button"
            onclick={() => (mobileOpen = !mobileOpen)}
            aria-expanded={mobileOpen}
            class="flex w-full items-center justify-between gap-3 p-5 text-left transition-colors hover:bg-surface-faint"
        >
            <span
                class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
            >
                [ evaluation playground ]
            </span>
            <ChevronDown
                class="h-4 w-4 text-muted-foreground transition-transform duration-200 {mobileOpen
                    ? 'rotate-180'
                    : ''}"
            />
        </button>
        {#if mobileOpen}
            <div
                transition:slide={{ duration: 200 }}
                class="border-t border-border/60 p-5"
            >
                <EvalPlayground
                    inputId="playground-context-mobile"
                    bind:contextJson={playgroundContext}
                    result={playgroundResult}
                    error={playgroundError}
                    running={playgroundRunning}
                    onevaluate={evaluatePlayground}
                    onreset={resetPlayground}
                    oninput={markDirty}
                />
            </div>
        {/if}
    </Card>
</div>

{#if drawerOpen}
    <aside
        transition:fly={{ x: 24, duration: 200 }}
        class="glass-panel fixed inset-y-0 right-0 z-40 hidden w-[26rem] max-w-[calc(100vw-1.5rem)] flex-col border-l border-border shadow-drawer lg:flex"
        aria-label="evaluation playground"
    >
        <div
            class="flex items-center justify-between gap-3 border-b border-border/60 px-5 pb-4 pt-28"
        >
            <p
                class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
            >
                [ evaluation playground ]
            </p>
            <button
                type="button"
                aria-label="close playground"
                onclick={() => (drawerOpen = false)}
                class="motion-press rounded-sm p-1 text-muted-foreground transition-colors hover:bg-surface-hover hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            >
                <X class="h-4 w-4" />
            </button>
        </div>
        <div class="flex-1 overflow-y-auto p-5">
            <EvalPlayground
                inputId="playground-context-desktop"
                bind:contextJson={playgroundContext}
                result={playgroundResult}
                error={playgroundError}
                running={playgroundRunning}
                onevaluate={evaluatePlayground}
                onreset={resetPlayground}
                oninput={markDirty}
            />
        </div>
    </aside>
{/if}

<svelte:window
    onkeydown={(e) => {
        if (e.key === "Escape" && drawerOpen) drawerOpen = false;
    }}
/>
