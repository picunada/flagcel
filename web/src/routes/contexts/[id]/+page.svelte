<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, APIError, type ContextSchema, type CreateContextRequest } from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import DestructiveDialog from '$lib/components/ui/destructive-dialog.svelte';
	import ContextEditor from '$lib/components/context-editor.svelte';
	import { Trash2 } from 'lucide-svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let schema: ContextSchema = $state(untrack(() => data.schema));
	let saveError = $state<string | null>(null);
	let submitting = $state(false);
	let deleteOpen = $state(false);
	let deleting = $state(false);
	let deleteError = $state<string | null>(null);

	$effect(() => {
		schema = data.schema;
	});

	async function save(form: CreateContextRequest) {
		submitting = true;
		saveError = null;
		try {
			schema = await api.updateContext(schema.id, form);
		} catch (e) {
			saveError = e instanceof APIError ? e.message : 'Failed to save context';
		} finally {
			submitting = false;
		}
	}

	async function remove() {
		deleting = true;
		deleteError = null;
		try {
			await api.deleteContext(schema.id);
			await goto('/contexts');
		} catch (e) {
			deleteError = e instanceof APIError ? e.message : 'Failed to delete context';
		} finally {
			deleting = false;
		}
	}
</script>

<div class="space-y-10">
	<a
		href="/contexts"
		class="inline-flex items-center gap-1.5 text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground"
	>
		← all contexts
	</a>

	<header class="flex flex-wrap items-start justify-between gap-4">
		<div class="space-y-3">
			<p
				class="text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground"
			>
				[ context ]
			</p>
			<h1 class="text-3xl font-normal tracking-tight sm:text-4xl">
				{schema.name}
			</h1>
		</div>
		<Button
			variant="destructive"
			onclick={() => {
				deleteError = null;
				deleteOpen = true;
			}}
		>
			<Trash2 class="h-3.5 w-3.5" /> delete
		</Button>
	</header>

	<Card class="motion-panel p-8">
		<ContextEditor
			{schema}
			{submitting}
			error={saveError}
			submitLabel="save changes"
			onsave={save}
		/>
	</Card>
</div>

<DestructiveDialog
	bind:open={deleteOpen}
	title="Delete context"
	description="Flags referencing this context will be unlinked."
	confirmationValue={schema.name}
	actionLabel="delete context"
	submitting={deleting}
	error={deleteError}
	onconfirm={remove}
/>
