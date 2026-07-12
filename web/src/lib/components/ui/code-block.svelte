<script lang="ts">
    import { cn } from "$lib/utils";

    let {
        class: className,
        value,
    }: { class?: string; value: string } = $props();
</script>

<code
    class={cn(
        "ios-corners-sm block overflow-x-auto whitespace-nowrap bg-surface-code p-3 font-mono text-xs text-foreground-strong",
        className,
    )}
>
    {value}
</code>
