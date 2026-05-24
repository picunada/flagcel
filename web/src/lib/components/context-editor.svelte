<script lang="ts">
	import { untrack } from 'svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import { cn } from '$lib/utils';
	import type {
		ContextField,
		ContextSchema,
		ContextType,
		CreateContextRequest
	} from '$lib/api';
	import { Plus, Trash2 } from 'lucide-svelte';

	type Props = {
		schema?: ContextSchema;
		submitting?: boolean;
		error?: string | null;
		onsave: (data: CreateContextRequest) => void | Promise<void>;
		oncancel?: () => void;
		submitLabel?: string;
		class?: string;
	};

	let {
		schema,
		submitting = false,
		error = null,
		onsave,
		oncancel,
		submitLabel = 'save',
		class: className
	}: Props = $props();

	const types: ContextType[] = [
		'string',
		'int',
		'double',
		'bool',
		'timestamp',
		'list',
		'map'
	];

	let name = $state(untrack(() => schema?.name ?? ''));
	let description = $state(untrack(() => schema?.description ?? ''));
	let fields = $state<ContextField[]>(
		untrack(() =>
			schema?.fields
				? schema.fields.map((f) => ({ ...f }))
				: [{ path: '', type: 'string' as ContextType }]
		)
	);

	const cleanedFields = $derived(
		fields
			.map((f) => ({ path: f.path.trim(), type: f.type }))
			.filter((f) => f.path.length > 0)
	);

	const canSubmit = $derived(name.trim().length > 0 && !submitting);

	function addField() {
		fields = [...fields, { path: '', type: 'string' }];
	}

	function removeField(i: number) {
		fields = fields.filter((_, idx) => idx !== i);
		if (fields.length === 0) addField();
	}

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		if (!canSubmit) return;
		await onsave({
			name: name.trim(),
			description: description.trim(),
			fields: cleanedFields
		});
	}
</script>

<form onsubmit={submit} class={cn('space-y-6', className)}>
	<div class="grid gap-4 sm:grid-cols-2">
		<div class="space-y-2">
			<label
				for="ctx-name"
				class="font-mono text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
			>
				name
			</label>
			<Input id="ctx-name" bind:value={name} required placeholder="web-user" />
		</div>
		<div class="space-y-2">
			<label
				for="ctx-desc"
				class="font-mono text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
			>
				description · optional
			</label>
			<Input
				id="ctx-desc"
				bind:value={description}
				placeholder="evaluation shape for the web app"
			/>
		</div>
	</div>

	<div class="space-y-3">
		<div class="flex items-center justify-between">
			<p
				class="font-mono text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
			>
				fields · dotted paths
			</p>
			<Button size="sm" variant="ghost" type="button" onclick={addField}>
				<Plus class="h-3 w-3" /> add field
			</Button>
		</div>

		<div class="space-y-2">
			{#each fields as field, i (i)}
				<div class="flex items-center gap-2">
					<Input
						bind:value={fields[i].path}
						placeholder="user.country"
						class="flex-1"
					/>
					<select
						bind:value={fields[i].type}
						class="h-9 rounded-sm border border-input bg-transparent px-2 font-mono text-xs uppercase tracking-[0.1em] transition-colors focus-visible:outline-none focus-visible:border-[rgba(255,255,255,0.36)]"
					>
						{#each types as t (t)}
							<option value={t}>{t}</option>
						{/each}
					</select>
					<Button
						size="icon"
						variant="ghost"
						type="button"
						onclick={() => removeField(i)}
						aria-label="remove field"
					>
						<Trash2 class="h-3.5 w-3.5" />
					</Button>
				</div>
			{/each}
		</div>
		<p class="font-mono text-[0.65rem] text-muted-foreground">
			example: user.id · user.country · request.path · time
		</p>
	</div>

	{#if error}
		<p class="font-mono text-xs text-destructive">{error}</p>
	{/if}

	<div class="flex justify-end gap-2 border-t border-border/60 pt-4">
		{#if oncancel}
			<Button variant="ghost" type="button" onclick={oncancel} disabled={submitting}>
				cancel
			</Button>
		{/if}
		<Button variant="solid" type="submit" disabled={!canSubmit}>
			{submitting ? 'saving…' : submitLabel}
		</Button>
	</div>
</form>
