<script lang="ts">
	import type { HTMLOlAttributes } from 'svelte/elements';
	import { cn, type WithElementRef } from '$lib/utils';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLOlAttributes> = $props();
</script>

<ol
	bind:this={ref}
	data-slot="breadcrumb-list"
	class={cn(
		'flex flex-wrap items-center gap-1.5 break-words font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</ol>
