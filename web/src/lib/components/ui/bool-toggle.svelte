<script lang="ts">
    import SegmentedControl from "$lib/components/ui/segmented-control.svelte";

    const booleanOptions = [
        { value: false, label: "false" },
        { value: true, label: "true" },
    ] as const;

    let {
        value = $bindable(false),
        onchange,
        disabled = false,
        class: className,
    }: {
        value?: boolean;
        onchange?: (value: boolean) => void;
        disabled?: boolean;
        class?: string;
    } = $props();

    function set(next: boolean) {
        if (disabled || next === value) return;
        value = next;
        onchange?.(next);
    }
</script>

<SegmentedControl
    label="Boolean value"
    options={booleanOptions}
    {value}
    onchange={set}
    {disabled}
    class={className}
/>
