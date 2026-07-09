<script lang="ts">
	import { cn, type WithElementRef } from '$lib/utils';
	import type { HTMLLiAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLLiAttributes> = $props();
</script>

<li
	bind:this={ref}
	data-slot="breadcrumb-separator"
	role="presentation"
	aria-hidden="true"
	class={cn('text-muted-foreground/50 select-none', className)}
	{...restProps}
>
	{#if children}
		{@render children?.()}
	{:else}
		<span>/</span>
	{/if}
</li>
