<script lang="ts">
	import { Popover as PopoverPrimitive } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		ref = $bindable(null),
		class: className,
		sideOffset = 6,
		align = 'start',
		children,
		...restProps
	}: PopoverPrimitive.ContentProps & { children?: Snippet } = $props();
</script>

<PopoverPrimitive.Portal>
	<PopoverPrimitive.Content
		bind:ref
		{sideOffset}
		{align}
		data-slot="popover-content"
		class={cn(
			'ios-corners-md motion-pop z-50 w-80 border border-border bg-popover-solid p-3 shadow-popover backdrop-blur-xl outline-none',
			className
		)}
		{...restProps}
	>
		{@render children?.()}
	</PopoverPrimitive.Content>
</PopoverPrimitive.Portal>
