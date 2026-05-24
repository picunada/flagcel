<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, APIError } from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';
	import ContextPicker from '$lib/components/context-picker.svelte';

	let key = $state('');
	let enabled = $state(true);
	let defaultValue = $state(false);
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
</script>

<div class="space-y-10">
	<a
		href="/"
		class="inline-flex items-center gap-1.5 font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground"
	>
		← all flags
	</a>

	<header class="space-y-3">
		<p class="font-mono text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground">
			[ new flag ]
		</p>
		<h1 class="font-mono text-3xl font-normal tracking-tight sm:text-4xl">
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
					class="font-mono text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
				>
					key
				</label>
				<Input
					id="key"
					bind:value={key}
					placeholder="new-checkout"
					required
					pattern="[a-z0-9][a-z0-9-_]*"
				/>
				<p class="font-mono text-[0.65rem] text-muted-foreground">
					lowercase alphanumerics, dashes, underscores
				</p>
			</div>

			<div class="space-y-2">
				<label
					for="context"
					class="font-mono text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
				>
					context · optional
				</label>
				<ContextPicker
					value={contextId}
					onchange={(v) => (contextId = v)}
				/>
			</div>

			<div class="space-y-5 border-t border-border/60 pt-6">
				<div class="flex items-center justify-between gap-4">
					<div class="space-y-1">
						<p class="font-mono text-sm">enabled</p>
						<p class="text-xs text-muted-foreground">
							when off, the default value is returned for every request
						</p>
					</div>
					<BoolToggle bind:value={enabled} />
				</div>

				<div class="flex items-center justify-between gap-4">
					<div class="space-y-1">
						<p class="font-mono text-sm">default value</p>
						<p class="text-xs text-muted-foreground">
							returned when no rule matches
						</p>
					</div>
					<BoolToggle bind:value={defaultValue} />
				</div>
			</div>

			{#if error}
				<p class="font-mono text-xs text-destructive">{error}</p>
			{/if}

			<div class="flex justify-end gap-2 border-t border-border/60 pt-6">
				<Button variant="ghost" href="/">cancel</Button>
				<Button variant="solid" type="submit" disabled={submitting || !key}>
					{submitting ? 'creating…' : 'create flag'}
				</Button>
			</div>
		</form>
	</Card>
</div>
