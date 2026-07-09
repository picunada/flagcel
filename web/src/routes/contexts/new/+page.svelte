<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, APIError, type CreateContextRequest } from '$lib/api';
	import AppBreadcrumbs from '$lib/components/ui/app-breadcrumbs.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import ContextEditor from '$lib/components/context-editor.svelte';
	import PageHeader from '$lib/components/ui/page-header.svelte';

	let submitting = $state(false);
	let error = $state<string | null>(null);

	async function save(data: CreateContextRequest) {
		submitting = true;
		error = null;
		try {
			const ctx = await api.createContext(data);
			await goto(`/contexts/${encodeURIComponent(ctx.id)}`);
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to create context';
		} finally {
			submitting = false;
		}
	}
</script>

<div class="space-y-10">
	<div class="space-y-3">
		<AppBreadcrumbs items={[{ label: 'contexts', href: '/contexts' }, { label: 'new' }]} />
		<PageHeader
			eyebrow="[ new context ]"
			title="Define a context"
			description="Each field is a dotted path your CEL rules can reference."
		/>
	</div>

	<Card class="p-8">
		<ContextEditor
			{submitting}
			{error}
			submitLabel="create context"
			onsave={save}
			oncancel={() => goto('/contexts')}
		/>
	</Card>
</div>
