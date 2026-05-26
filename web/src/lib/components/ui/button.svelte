<script lang="ts" module>
	import { tv, type VariantProps } from 'tailwind-variants';

	export const buttonVariants = tv({
		base: 'motion-press inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-sm text-xs uppercase tracking-[0.1em] transition-all duration-200 ease-out focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-40',
		variants: {
			variant: {
				default:
					'border border-[rgba(255,255,255,0.16)] bg-transparent text-foreground hover:border-[rgba(255,255,255,0.36)] hover:bg-[rgba(255,255,255,0.04)]',
				solid:
					'bg-foreground text-primary-foreground hover:bg-foreground/90',
				ghost:
					'text-muted-foreground hover:text-foreground hover:bg-[rgba(255,255,255,0.04)]',
				destructive:
					'border border-[rgba(255,107,107,0.3)] bg-transparent text-destructive hover:border-destructive hover:bg-[rgba(255,107,107,0.06)]'
			},
			size: {
				default: 'h-9 px-4',
				sm: 'h-7 px-3 text-[0.65rem]',
				lg: 'h-11 px-6',
				icon: 'h-9 w-9'
			}
		},
		defaultVariants: { variant: 'default', size: 'default' }
	});

	export type ButtonVariant = VariantProps<typeof buttonVariants>['variant'];
	export type ButtonSize = VariantProps<typeof buttonVariants>['size'];
</script>

<script lang="ts">
	import type { HTMLButtonAttributes, HTMLAnchorAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Props = {
		variant?: ButtonVariant;
		size?: ButtonSize;
		class?: string;
		href?: string;
		children?: Snippet;
	} & Omit<HTMLButtonAttributes, 'class'> &
		Omit<HTMLAnchorAttributes, 'class'>;

	let {
		variant = 'default',
		size = 'default',
		class: className,
		href,
		children,
		...rest
	}: Props = $props();
</script>

{#if href}
	<a {href} class={cn(buttonVariants({ variant, size }), className)} {...rest}>
		{@render children?.()}
	</a>
{:else}
	<button class={cn(buttonVariants({ variant, size }), className)} {...rest}>
		{@render children?.()}
	</button>
{/if}
