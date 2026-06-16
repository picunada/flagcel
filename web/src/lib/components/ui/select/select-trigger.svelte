<script lang="ts">
	import { Select as SelectPrimitive } from 'bits-ui';
	import { ChevronDown } from 'lucide-svelte';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: SelectPrimitive.TriggerProps & { children?: Snippet } = $props();
</script>

<SelectPrimitive.Trigger
	bind:ref
	data-slot="select-trigger"
	class={cn(
		'group inline-flex h-9 w-full items-center justify-between gap-2 rounded-sm border border-input bg-transparent px-2.5 text-left text-xs text-foreground outline-none transition-colors hover:border-[rgb(var(--app-accent-rgb)/0.48)] hover:bg-[rgb(var(--app-accent-rgb)/0.06)] focus-visible:border-[rgb(var(--app-accent-rgb)/0.64)] focus-visible:ring-1 focus-visible:ring-[rgb(var(--app-accent-rgb)/0.42)] disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	{...restProps}
>
	{@render children?.()}
	<ChevronDown
		class="h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-180"
	/>
</SelectPrimitive.Trigger>
