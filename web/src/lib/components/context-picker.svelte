<script lang="ts">
	import { onMount } from 'svelte';
	import { api, APIError, type ContextSchema } from '$lib/api';
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

	function handleChange(e: Event) {
		const v = (e.target as HTMLSelectElement).value;
		onchange?.(v === '' ? null : v);
	}
</script>

<div class={cn('space-y-1', className)}>
	<select
		{disabled}
		value={value ?? ''}
		onchange={handleChange}
		class="h-9 w-full rounded-sm border border-input bg-transparent px-3 font-mono text-sm transition-colors focus-visible:outline-none focus-visible:border-[rgba(255,255,255,0.36)] disabled:cursor-not-allowed disabled:opacity-50"
	>
		<option value="">— no context —</option>
		{#each contexts as ctx (ctx.id)}
			<option value={ctx.id}>{ctx.name}</option>
		{/each}
	</select>
	{#if loading}
		<p class="font-mono text-[0.65rem] text-muted-foreground">loading contexts…</p>
	{:else if error}
		<p class="font-mono text-[0.65rem] text-destructive">{error}</p>
	{:else if contexts.length === 0}
		<p class="font-mono text-[0.65rem] text-muted-foreground">
			no contexts defined ·
			<a href="/contexts/new" class="underline">create one</a>
		</p>
	{/if}
</div>
