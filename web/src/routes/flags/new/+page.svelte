<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, APIError, type FlagValue, type ValueType } from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';
	import ContextPicker from '$lib/components/context-picker.svelte';
	import ValueEditor from '$lib/components/value-editor.svelte';
	import { defaultValueForType } from '$lib/values';

	let key = $state('');
	let type = $state<ValueType>('boolean');
	let enabled = $state(true);
	let defaultValue = $state<FlagValue>(false);
	let defaultValid = $state(true);
	let contextId = $state<string | null>(null);
	let submitting = $state(false);
	let error = $state<string | null>(null);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		submitting = true;
		error = null;
		try {
			const flag = await api.createFlag({
				key,
				type,
				enabled,
				default_value: defaultValue,
				rules: [],
				context_id: contextId
			});
			await goto(`/flags/${encodeURIComponent(flag.key)}`);
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to create flag';
		} finally {
			submitting = false;
		}
	}

	function setType(next: ValueType) {
		type = next;
		defaultValue = defaultValueForType(next);
		defaultValid = true;
	}
</script>

<div class="space-y-10">
	<a
		href="/"
		class="inline-flex items-center gap-1.5 text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground"
	>
		← all flags
	</a>

	<header class="space-y-3">
		<p class="text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground">
			[ new flag ]
		</p>
		<h1 class="text-3xl font-normal tracking-tight sm:text-4xl">
			Define a flag
		</h1>
		<p class="max-w-lg text-sm text-[rgba(255,255,255,0.7)]">
			Pick a unique key. Add targeting rules after creation.
		</p>
	</header>

	<Card class="p-8">
		<form onsubmit={submit} class="space-y-6">
			<div class="space-y-2">
				<label
					for="key"
					class="text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
				>
					key
				</label>
				<Input
					id="key"
					bind:value={key}
					placeholder="new-checkout"
					required
					pattern="[a-z0-9][a-z0-9-_]*"
					class="font-mono"
				/>
				<p class="text-[0.65rem] text-muted-foreground">
					lowercase alphanumerics, dashes, underscores
				</p>
			</div>

			<div class="space-y-2">
				<label
					for="context"
					class="text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
				>
					context · optional
				</label>
				<ContextPicker
					value={contextId}
					onchange={(v) => (contextId = v)}
				/>
			</div>

			<div class="space-y-2">
				<label
					for="type"
					class="text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
				>
					type
				</label>
				<select
					id="type"
					value={type}
					onchange={(e) => setType(e.currentTarget.value as ValueType)}
					class="h-9 w-full rounded-sm border border-input bg-transparent px-2.5 text-sm text-foreground outline-none transition-colors focus-visible:ring-1 focus-visible:ring-ring [&>option]:bg-background"
				>
					<option value="boolean">boolean</option>
					<option value="string">string</option>
					<option value="number">number</option>
					<option value="json">json</option>
				</select>
			</div>

			<div class="space-y-5 border-t border-border/60 pt-6">
				<div class="flex items-center justify-between gap-4">
					<div class="space-y-1">
						<p class="text-sm">enabled</p>
						<p class="text-xs text-muted-foreground">
							when off, the default value is returned for every request
						</p>
					</div>
					<BoolToggle bind:value={enabled} />
				</div>

				<div class="flex flex-wrap items-start justify-between gap-4">
					<div class="space-y-1">
						<p class="text-sm">default value</p>
						<p class="text-xs text-muted-foreground">
							returned when no rule matches
						</p>
					</div>
					<div class="min-w-48 max-w-full flex-1 sm:flex-none sm:basis-72">
						<ValueEditor
							type={type}
							value={defaultValue}
							id="default-value"
							align="end"
							onchange={(v) => (defaultValue = v)}
							onvalid={(v) => (defaultValid = v)}
						/>
					</div>
				</div>
			</div>

			{#if error}
				<p class="text-xs text-destructive">{error}</p>
			{/if}

			<div class="flex justify-end gap-2 border-t border-border/60 pt-6">
				<Button variant="ghost" href="/">cancel</Button>
				<Button variant="solid" type="submit" disabled={submitting || !key || !defaultValid}>
					{submitting ? 'creating…' : 'create flag'}
				</Button>
			</div>
		</form>
	</Card>
</div>
