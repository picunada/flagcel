<script lang="ts">
	import { onMount } from 'svelte';
	import { api, APIError, type ContextSchema } from '$lib/api';
	import ThemedSelect from '$lib/components/ui/themed-select.svelte';
	import { cn } from '$lib/utils';

	type Props = {
		value: string | null | undefined;
		onchange?: (id: string | null) => void;
		disabled?: boolean;
		compact?: boolean;
		class?: string;
		buttonClass?: string;
	};

	let {
		value,
		onchange,
		disabled = false,
		compact = false,
		class: className,
		buttonClass
	}: Props = $props();

	let contexts = $state<ContextSchema[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			contexts = await api.listContexts();
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to load contexts';
		} finally {
			loading = false;
		}
	});

	const options = $derived([
		{ value: '', label: 'no context' },
		...contexts.map((ctx) => ({ value: ctx.id, label: ctx.name }))
	]);

	const statusMessage = $derived(
		loading
			? 'loading contexts…'
			: error
				? error
				: contexts.length === 0
					? 'no contexts defined'
					: null
	);
</script>

<div class={cn(compact ? 'min-w-0' : 'space-y-1', className)}>
	<ThemedSelect
		{disabled}
		value={value ?? ''}
		{options}
		onchange={(v) => onchange?.(v === '' ? null : v)}
		buttonClass={cn(compact ? 'text-xs' : 'text-sm px-3', buttonClass)}
		class={compact ? 'min-w-0' : undefined}
	/>
	{#if !compact}
		{#if loading}
			<p class="text-[0.65rem] text-muted-foreground">loading contexts…</p>
		{:else if error}
			<p class="text-[0.65rem] text-destructive">{error}</p>
		{:else if contexts.length === 0}
			<p class="text-[0.65rem] text-muted-foreground">
				no contexts defined ·
				<a href="/contexts/new" class="underline">create one</a>
			</p>
		{/if}
	{:else if statusMessage}
		<span class="sr-only">{statusMessage}</span>
	{/if}
</div>
