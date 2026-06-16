<script lang="ts">
	import { Select as SelectPrimitive } from 'bits-ui';
	import { Check } from 'lucide-svelte';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		ref = $bindable(null),
		class: className,
		value,
		label,
		children: childrenProp,
		...restProps
	}: SelectPrimitive.ItemProps & { children?: Snippet } = $props();
</script>

<SelectPrimitive.Item
	bind:ref
	{value}
	{label}
	data-slot="select-item"
	class={cn(
		'flex min-h-8 w-full cursor-pointer select-none items-center justify-between gap-2 rounded-[2px] px-2 py-1.5 text-left text-xs text-muted-foreground outline-none transition-colors data-[highlighted]:bg-[rgb(var(--app-accent-rgb)/0.12)] data-[highlighted]:text-foreground data-[selected]:bg-[rgb(var(--app-accent-rgb)/0.18)] data-[selected]:text-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
		className
	)}
	{...restProps}
>
	{#snippet children({ selected })}
		<span class="min-w-0">
			{#if childrenProp}
				{@render childrenProp()}
			{:else}
				<span class="block truncate">{label}</span>
			{/if}
		</span>
		{#if selected}
			<Check class="h-3.5 w-3.5 shrink-0 text-app-accent" />
		{/if}
	{/snippet}
</SelectPrimitive.Item>
