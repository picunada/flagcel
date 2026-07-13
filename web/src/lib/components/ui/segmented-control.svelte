<script lang="ts" generics="T extends string | number | boolean">
    import { cn } from "$lib/utils";

    type Option = {
        value: T;
        label: string;
        ariaLabel?: string;
        title?: string;
        icon?: any;
        iconOnly?: boolean;
    };

    type Props = {
        label: string;
        options: readonly Option[];
        value: T;
        onchange: (value: T) => void;
        disabled?: boolean;
        class?: string;
        buttonClass?: string;
    };

    let {
        label,
        options,
        value,
        onchange,
        disabled = false,
        class: className,
        buttonClass,
    }: Props = $props();
</script>

<div
    role="group"
    aria-label={label}
    aria-disabled={disabled || undefined}
    class={cn(
        "ios-corners-sm inline-flex h-9 shrink-0 border border-border p-0.5",
        disabled && "opacity-50",
        className,
    )}
>
    {#each options as option (option.value)}
        {@const Icon = option.icon}
        <button
            type="button"
            title={option.title}
            aria-label={option.ariaLabel ?? (option.iconOnly ? option.label : undefined)}
            aria-pressed={value === option.value}
            {disabled}
            class={cn(
                "ios-corners-xs inline-flex h-full cursor-pointer items-center justify-center gap-1.5 px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed",
                option.iconOnly && "w-8 px-0",
                value === option.value
                    ? "bg-surface-selected text-foreground"
                    : "text-muted-foreground hover:text-foreground",
                buttonClass,
            )}
            onclick={() => onchange(option.value)}
        >
            {#if Icon}
                <Icon class="h-3.5 w-3.5" />
            {/if}
            {#if !option.iconOnly}{option.label}{/if}
        </button>
    {/each}
</div>
