<script lang="ts">
    import { Switch as SwitchPrimitive } from "bits-ui";
    import { cn } from "$lib/utils.js";

    let {
        ref = $bindable(null),
        class: className,
        checked = $bindable(false),
        size = "default",
        ...restProps
    }: SwitchPrimitive.RootProps & {
        size?: "sm" | "default";
    } = $props();
</script>

<SwitchPrimitive.Root
    bind:ref
    bind:checked
    data-slot="switch"
    data-size={size}
    class={cn(
        "data-[state=checked]:bg-app-accent data-[state=unchecked]:bg-input focus-visible:border-ring focus-visible:ring-ring/50 shrink-0\
         rounded-full border border-transparent focus-visible:ring-3 data-disabled:cursor-not-allowed data-disabled:opacity-50\
         data-[size=default]:h-[18.4px] data-[size=default]:w-[32px] data-[size=default]:[--switch-thumb-size:16px] data-[size=default]:[--switch-thumb-x:14px] data-[size=sm]:h-[14px] data-[size=sm]:w-[24px] data-[size=sm]:[--switch-thumb-size:12px] data-[size=sm]:[--switch-thumb-x:10px] peer group/switch\
         relative inline-flex items-center transition-all outline-none after:absolute after:-inset-x-3 after:-inset-y-2",
        className,
    )}
    {...restProps}
>
    <SwitchPrimitive.Thumb
        data-slot="switch-thumb"
        class="pointer-events-none block size-[var(--switch-thumb-size)] rounded-full bg-background transition-transform data-[state=checked]:translate-x-[var(--switch-thumb-x)] data-[state=unchecked]:translate-x-0 rtl:data-[state=checked]:-translate-x-[var(--switch-thumb-x)]"
    />
</SwitchPrimitive.Root>
