<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		class: className,
		hoverable = false,
		children
	}: { class?: string; hoverable?: boolean; children?: Snippet } = $props();
</script>

<div
	class={cn(
		'glass-panel rounded-sm transition-all duration-200',
		hoverable &&
			'hover:border-[rgba(255,255,255,0.36)] hover:-translate-y-0.5 hover:bg-[rgba(255,255,255,0.04)]',
		className
	)}
>
	{@render children?.()}
</div>
