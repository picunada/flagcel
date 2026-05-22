<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type Props = { class?: string; value?: string } & Omit<HTMLInputAttributes, 'class'>;
	let { class: className, value = $bindable(''), ...rest }: Props = $props();
</script>

<input
	bind:value
	class={cn(
		'flex h-9 w-full rounded-sm border border-input bg-transparent px-3 py-1 font-mono text-sm transition-colors',
		'placeholder:text-muted-foreground placeholder:lowercase',
		'focus-visible:outline-none focus-visible:border-[rgba(255,255,255,0.36)]',
		'disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	{...rest}
/>
