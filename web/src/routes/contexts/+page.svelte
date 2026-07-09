<script lang="ts">
	import { APIError, type ContextSchema } from '$lib/api';
	import { invalidateAll } from '$app/navigation';
	import ContextCardGrid from '$lib/components/contexts/context-card-grid.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import EmptyState from '$lib/components/ui/empty-state.svelte';
	import AppBreadcrumbs from '$lib/components/ui/app-breadcrumbs.svelte';
	import PageHeader from '$lib/components/ui/page-header.svelte';
	import SectionHeader from '$lib/components/ui/section-header.svelte';
	import { Plus } from 'lucide-svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const contexts = $derived<ContextSchema[]>(data.contexts);
	let error = $state<string | null>(null);

	async function refresh() {
		error = null;
		try {
			await invalidateAll();
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to load contexts';
		}
	}
</script>

<section class="space-y-12">
	<div class="space-y-3">
		<AppBreadcrumbs items={[{ label: 'contexts' }]} />
		<PageHeader eyebrow="evaluation contexts · field shapes" />
	</div>

	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
		<SectionHeader>contexts · {contexts.length}</SectionHeader>
		<Button href="/contexts/new">
			<Plus class="h-3.5 w-3.5" /> new context
		</Button>
	</div>

	{#if error}
		<Card class="motion-panel p-8 text-center">
			<p class="text-sm text-destructive">{error}</p>
			<Button class="mt-4" onclick={refresh}>retry</Button>
		</Card>
	{:else if contexts.length === 0}
		<EmptyState
			title="[ no contexts yet ]"
			description="Define one to enable autocomplete on rules."
			class="p-12"
		>
			<Button href="/contexts/new">
				<Plus class="h-3.5 w-3.5" /> new context
			</Button>
		</EmptyState>
	{:else}
		<ContextCardGrid {contexts} />
	{/if}
</section>
