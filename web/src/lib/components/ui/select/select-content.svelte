<script lang="ts">
    import { Select as SelectPrimitive } from "bits-ui";
    import type { Snippet } from "svelte";
    import { cn } from "$lib/utils";

    let {
        ref = $bindable(null),
        class: className,
        sideOffset = 4,
        align = "start",
        children,
        ...restProps
    }: SelectPrimitive.ContentProps & { children?: Snippet } = $props();
</script>

<SelectPrimitive.Portal>
    <SelectPrimitive.Content
        bind:ref
        {sideOffset}
        {align}
        data-slot="select-content"
        class={cn(
            "ios-corners-md motion-pop z-50 max-h-72 min-w-[var(--bits-select-anchor-width)] overflow-hidden  bg-popover-solid shadow-popover backdrop-blur-xl",
            className,
        )}
        {...restProps}
    >
        <SelectPrimitive.Viewport class="max-h-72 overflow-y-auto p-1">
            {@render children?.()}
        </SelectPrimitive.Viewport>
    </SelectPrimitive.Content>
</SelectPrimitive.Portal>
