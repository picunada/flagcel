<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		class: className,
		children
	}: { class?: string; children?: Snippet } = $props();
</script>

<h2
	class={cn(
		'font-mono text-xs uppercase tracking-[0.12em] text-muted-foreground',
		className
	)}
>
	[ {@render children?.()} ]
</h2>
