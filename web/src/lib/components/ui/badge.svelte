<script lang="ts" module>
	import { tv, type VariantProps } from 'tailwind-variants';

	export const badgeVariants = tv({
		base: 'inline-flex items-center gap-1.5 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground',
		variants: {
			variant: {
				default: '',
				success: 'text-success',
				muted: 'text-muted-foreground',
				destructive: 'text-destructive'
			}
		},
		defaultVariants: { variant: 'default' }
	});

	export type BadgeVariant = VariantProps<typeof badgeVariants>['variant'];
</script>

<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	let {
		variant = 'default',
		dot = false,
		class: className,
		children
	}: { variant?: BadgeVariant; dot?: boolean; class?: string; children?: Snippet } = $props();

	const dotColor = $derived(
		variant === 'success'
			? 'bg-success shadow-[0_0_8px_var(--color-success)]'
			: variant === 'destructive'
				? 'bg-destructive'
				: 'bg-muted-foreground'
	);
</script>

<span class={cn(badgeVariants({ variant }), className)}>
	{#if dot}
		<span class={cn('inline-block h-1.5 w-1.5 rounded-full', dotColor)}></span>
	{/if}
	{@render children?.()}
</span>
