<script lang="ts">
	import { onMount } from 'svelte';
	import { api, APIError, type ContextSchema } from '$lib/api';
	import ThemedSelect from '$lib/components/ui/themed-select.svelte';
	import { cn } from '$lib/utils';

	type Props = {
		value: string | null | undefined;
		onchange?: (id: string | null) => void;
		disabled?: boolean;
		class?: string;
	};

	let { value, onchange, disabled = false, class: className }: Props = $props();

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
</script>

<div class={cn('space-y-1', className)}>
	<ThemedSelect
		{disabled}
		value={value ?? ''}
		{options}
		onchange={(v) => onchange?.(v === '' ? null : v)}
		buttonClass="text-sm px-3"
	/>
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
</div>
