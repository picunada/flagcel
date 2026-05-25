<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		value = $bindable(false),
		onchange,
		disabled = false,
		class: className
	}: {
		value?: boolean;
		onchange?: (v: boolean) => void;
		disabled?: boolean;
		class?: string;
	} = $props();

	function set(v: boolean) {
		if (disabled || v === value) return;
		value = v;
		onchange?.(v);
	}
</script>

<div
	class={cn(
		'inline-flex items-center gap-2 font-mono text-sm uppercase tracking-[0.12em]',
		disabled && 'pointer-events-none opacity-50',
		className
	)}
>
	<button
		type="button"
		onclick={() => set(false)}
		class="motion-press transition-all duration-200 ease-out {!value
			? 'scale-105 text-foreground'
			: 'text-muted-foreground hover:text-foreground/80'}"
	>
		false
	</button>
	<span class="text-muted-foreground/50">/</span>
	<button
		type="button"
		onclick={() => set(true)}
		class="motion-press transition-all duration-200 ease-out {value
			? 'scale-105 text-success [text-shadow:0_0_12px_color-mix(in_oklab,var(--color-success)_60%,transparent)]'
			: 'text-muted-foreground hover:text-foreground/80'}"
	>
		true
	</button>
</div>
