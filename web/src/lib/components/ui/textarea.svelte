<script lang="ts">
    import type { HTMLTextareaAttributes } from "svelte/elements";
    import { cn } from "$lib/utils";

    type Props = {
        class?: string;
        value?: string;
    } & Omit<HTMLTextareaAttributes, "class">;

    let { class: className, value = $bindable(""), ...rest }: Props = $props();
</script>

<textarea
    bind:value
    class={cn(
        "ios-corners-sm flex w-full border border-input bg-transparent px-3 py-2 text-sm transition-all duration-200 ease-out",
        "placeholder:text-muted-foreground/60 focus-visible:border-border-hover focus-visible:bg-surface-subtle focus-visible:outline-none",
        "disabled:cursor-not-allowed disabled:opacity-50",
        className,
    )}
    {...rest}
></textarea>
