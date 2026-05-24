<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, APIError, type CreateContextRequest } from '$lib/api';
	import Card from '$lib/components/ui/card.svelte';
	import ContextEditor from '$lib/components/context-editor.svelte';

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
	<a
		href="/contexts"
		class="inline-flex items-center gap-1.5 font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground"
	>
		← all contexts
	</a>

	<header class="space-y-3">
		<p class="font-mono text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground">
			[ new context ]
		</p>
		<h1 class="font-mono text-3xl font-normal tracking-tight sm:text-4xl">
			Define a context
		</h1>
		<p class="max-w-lg text-sm text-[rgba(255,255,255,0.7)]">
			Each field is a dotted path your CEL rules can reference.
		</p>
	</header>

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
