<script lang="ts">
    import { usageRangeOptions, type UsageRangeHours } from "$lib/usage-analytics";
    import { cn } from "$lib/utils";

    type Props = {
        value: UsageRangeHours;
        onchange: (value: UsageRangeHours) => void;
    };

    let { value, onchange }: Props = $props();
</script>

<div
    class="ios-corners-sm inline-flex h-9 shrink-0 border border-border p-0.5"
    aria-label="Usage range"
>
    {#each usageRangeOptions as option (option.value)}
        <button
            type="button"
            aria-pressed={value === option.value}
            class={cn(
                "ios-corners-xs h-full px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors",
                value === option.value
                    ? "bg-surface-selected text-foreground"
                    : "text-muted-foreground hover:text-foreground",
            )}
            onclick={() => onchange(option.value)}
        >
            {option.label}
        </button>
    {/each}
</div>
