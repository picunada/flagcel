<script lang="ts">
    import { Tooltip as TooltipPrimitive } from "bits-ui";
    import { cn } from "$lib/utils.js";
    import TooltipPortal from "./tooltip-portal.svelte";
    import type { ComponentProps } from "svelte";

    let {
        ref = $bindable(null),
        class: className,
        sideOffset = 0,
        side = "top",
        children,
        arrowClasses,
        portalProps,
        ...restProps
    }: TooltipPrimitive.ContentProps & {
        arrowClasses?: string;
        portalProps?: Omit<
            ComponentProps<typeof TooltipPortal>,
            "children" | "child"
        >;
    } = $props();
</script>

<TooltipPortal {...portalProps}>
    <TooltipPrimitive.Content
        bind:ref
        data-slot="tooltip-content"
        {sideOffset}
        {side}
        class={cn(
            "motion-tooltip inline-flex items-center gap-1.5 rounded-md border border-border bg-surface-subtle px-3 py-1.5 text-xs text-foreground-soft backdrop-blur-xl has-data-[slot=kbd]:pr-1.5 **:data-[slot=kbd]:relative **:data-[slot=kbd]:isolate **:data-[slot=kbd]:z-50 **:data-[slot=kbd]:rounded-sm z-50 w-fit max-w-xs",
            className,
        )}
        {...restProps}
    >
        {@render children?.()}
    </TooltipPrimitive.Content>
</TooltipPortal>
