<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type Props = {
		class?: string;
		value?: string;
		ref?: HTMLInputElement | null;
	} & Omit<HTMLInputAttributes, 'class'>;

	let {
		class: className,
		value = $bindable(''),
		ref = $bindable(null),
		...rest
	}: Props = $props();
</script>

<input
	bind:this={ref}
	bind:value
	class={cn(
		'ios-corners-sm flex h-9 w-full border border-input bg-transparent px-3 py-1 text-sm transition-all duration-200 ease-out',
		'placeholder:text-muted-foreground placeholder:lowercase',
		'focus-visible:outline-none focus-visible:border-border-hover focus-visible:bg-surface-subtle',
		'disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	{...rest}
/>
