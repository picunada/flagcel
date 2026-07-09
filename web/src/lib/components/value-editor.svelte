<script lang="ts">
	import { untrack } from 'svelte';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import Textarea from '$lib/components/ui/textarea.svelte';
	import type { FlagValue, ValueType } from '$lib/api';
	import { cn } from '$lib/utils';
	import { formatFlagValue } from '$lib/values';

	type Props = {
		type: ValueType;
		value: FlagValue;
		disabled?: boolean;
		id?: string;
		align?: 'start' | 'end';
		compact?: boolean;
		fill?: boolean;
		class?: string;
		onchange: (value: FlagValue) => void;
		onvalid?: (valid: boolean) => void;
	};

	let {
		type,
		value,
		disabled = false,
		id = 'value',
		align = 'start',
		compact = false,
		fill = false,
		class: className,
		onchange,
		onvalid
	}: Props = $props();

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
	<div class={cn('min-w-0 space-y-1.5', fill && 'flex h-full flex-col')}>
		<Textarea
			{id}
			bind:value={text}
			oninput={handleTextInput}
			rows={compact ? 3 : 5}
			spellcheck="false"
			{disabled}
			class={cn(
				compact
					? 'h-16 min-h-16 w-56 resize-none font-mono text-xs leading-5'
					: fill
						? 'min-h-16 w-full flex-1 resize-y font-mono text-xs leading-5'
						: 'min-h-28 resize-y font-mono leading-5',
				className
			)}
		></Textarea>
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
			class={compact
				? 'h-8 w-36 font-mono text-xs'
				: type === 'string'
					? 'font-mono'
					: ''}
		/>
		{#if error}
			<p class="text-[0.65rem] text-destructive">{error}</p>
		{/if}
	</div>
{/if}
