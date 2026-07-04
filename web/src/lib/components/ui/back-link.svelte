<script lang="ts">
    import { cn } from "$lib/utils";

    let {
        href,
        label,
        class: className,
    }: { href: string; label: string; class?: string } = $props();
</script>

<a
    {href}
    class={cn(
        "inline-flex items-center gap-1.5 text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground",
        className,
    )}
>
    ← {label}
</a>
