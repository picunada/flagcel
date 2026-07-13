<script lang="ts">
    import { usageRangeOptions, type UsageRangeHours } from "$lib/usage-analytics";
    import SegmentedControl from "$lib/components/ui/segmented-control.svelte";

    type Props = {
        value: UsageRangeHours;
        onchange: (value: UsageRangeHours) => void;
    };

    let { value, onchange }: Props = $props();
</script>

<SegmentedControl
    label="Usage range"
    options={usageRangeOptions}
    {value}
    {onchange}
/>
