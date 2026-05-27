<script lang="ts">
	import { untrack } from 'svelte';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import type { FlagValue, ValueType } from '$lib/api';
	import { formatFlagValue } from '$lib/values';

	type Props = {
		type: ValueType;
		value: FlagValue;
		disabled?: boolean;
		id?: string;
		align?: 'start' | 'end';
		onchange: (value: FlagValue) => void;
		onvalid?: (valid: boolean) => void;
	};

	let { type, value, disabled = false, id = 'value', align = 'start', onchange, onvalid }: Props = $props();

	let text = $state(untrack(() => textForValue(type, value)));
	let error = $state<string | null>(null);
	let lastSync = $state('');

	$effect(() => {
		const next = syncKey(type, value);
		if (next !== lastSync) {
			text = textForValue(type, value);
			lastSync = next;
			validate(false);
		}
	});

	function textForValue(valueType: ValueType, v: FlagValue): string {
		if (valueType === 'json') return JSON.stringify(v ?? null, null, 2);
		return formatFlagValue(v);
	}

	function syncKey(valueType: ValueType, v: FlagValue): string {
		return `${valueType}:${JSON.stringify(v)}`;
	}

	function setBoolean(next: boolean) {
		error = null;
		onvalid?.(true);
		onchange(next);
	}

	function handleTextInput() {
		validate(true);
	}

	function validate(emit: boolean) {
		if (type === 'string') {
			error = null;
			onvalid?.(true);
			if (emit) onchange(text);
			return;
		}
		if (type === 'number') {
			const next = Number(text);
			if (text.trim() === '' || !Number.isFinite(next)) {
				error = 'Enter a valid number.';
				onvalid?.(false);
				return;
			}
			error = null;
			onvalid?.(true);
			if (emit) onchange(next);
			return;
		}
		if (type === 'json') {
			try {
				const parsed = JSON.parse(text);
				error = null;
				onvalid?.(true);
				if (emit) onchange(parsed);
			} catch (e) {
				error = e instanceof Error ? e.message : 'Invalid JSON.';
				onvalid?.(false);
			}
		}
	}
</script>

{#if type === 'boolean'}
	<div class={align === 'end' ? 'flex justify-end' : ''}>
		<BoolToggle value={Boolean(value)} {disabled} onchange={setBoolean} />
	</div>
{:else if type === 'json'}
	<div class="min-w-0 space-y-1.5">
		<textarea
			{id}
			bind:value={text}
			oninput={handleTextInput}
			rows="5"
			spellcheck="false"
			{disabled}
			class="min-h-28 w-full resize-y rounded-sm border border-input bg-transparent px-3 py-2 font-mono text-sm leading-5 transition-colors placeholder:text-muted-foreground/60 focus-visible:border-[rgba(255,255,255,0.36)] focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
		></textarea>
		{#if error}
			<p class="text-[0.65rem] text-destructive">{error}</p>
		{/if}
	</div>
{:else}
	<div class="min-w-0 space-y-1.5">
		<Input
			{id}
			type={type === 'number' ? 'number' : 'text'}
			step={type === 'number' ? 'any' : undefined}
			bind:value={text}
			oninput={handleTextInput}
			{disabled}
			class={type === 'string' ? 'font-mono' : ''}
		/>
		{#if error}
			<p class="text-[0.65rem] text-destructive">{error}</p>
		{/if}
	</div>
{/if}
