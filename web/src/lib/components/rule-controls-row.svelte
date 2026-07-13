<script lang="ts">
	import { tick } from 'svelte';
	import type { ContextField, FlagValue, ValueType } from '$lib/api';
	import ValueEditor from '$lib/components/value-editor.svelte';
	import { cn } from '$lib/utils';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';

	type Props = {
		valueType: ValueType;
		value: FlagValue;
		percentage: string;
		bucketBy: string;
		fields?: ContextField[];
		disabled?: boolean;
		submitting?: boolean;
		hideValue?: boolean;
		onvalue: (value: FlagValue) => void;
		onvalid: (valid: boolean) => void;
		onpercentage: (value: string) => void;
		onbucket: (value: string) => void;
	};

	let {
		valueType,
		value,
		percentage,
		bucketBy,
		fields = [],
		disabled = false,
		submitting = false,
		hideValue = false,
		onvalue,
		onvalid,
		onpercentage,
		onbucket
	}: Props = $props();

	let bucketInput: HTMLInputElement | null = $state(null);
	let bucketAcOpen = $state(false);
	let bucketAcIndex = $state(0);

	const bucketSuggestions = $derived.by(() => {
		const token = bucketBy.trim().toLowerCase();
		if (!token || !fields.length) return [];
		return fields
			.filter((field) => field.path.toLowerCase().startsWith(token) && field.path !== bucketBy.trim())
			.slice(0, 8);
	});

	$effect(() => {
		if (bucketSuggestions.length === 0) {
			bucketAcOpen = false;
			bucketAcIndex = 0;
		} else if (bucketAcIndex >= bucketSuggestions.length) {
			bucketAcIndex = 0;
		}
	});

	function setBoolean(next: boolean) {
		onvalid(true);
		onvalue(next);
	}

	function handleBucketInput(e: Event) {
		onbucket((e.currentTarget as HTMLInputElement).value);
		if (bucketSuggestions.length > 0) bucketAcOpen = true;
	}

	async function insertBucket(field: ContextField) {
		onbucket(field.path);
		bucketAcOpen = false;
		await tick();
		bucketInput?.focus();
		bucketInput?.setSelectionRange(field.path.length, field.path.length);
	}

	function handleBucketKeydown(e: KeyboardEvent) {
		if (bucketAcOpen && bucketSuggestions.length > 0) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				bucketAcIndex = (bucketAcIndex + 1) % bucketSuggestions.length;
				return;
			}
			if (e.key === 'ArrowUp') {
				e.preventDefault();
				bucketAcIndex = (bucketAcIndex - 1 + bucketSuggestions.length) % bucketSuggestions.length;
				return;
			}
			if (e.key === 'Enter' || e.key === 'Tab') {
				e.preventDefault();
				void insertBucket(bucketSuggestions[bucketAcIndex]);
				return;
			}
			if (e.key === 'Escape') {
				e.preventDefault();
				bucketAcOpen = false;
			}
		}
	}
</script>

<div
	class="ios-corners-md flex flex-wrap items-start gap-x-5 gap-y-3 border border-border-control bg-surface-faint p-3"
>
	{#if !hideValue}
		<div class="min-w-0 space-y-2">
			<div class="h-4 text-[0.65rem] uppercase leading-4 tracking-[0.14em] text-muted-foreground">
				value
			</div>
			{#if valueType === 'boolean'}
				<BoolToggle
					value={Boolean(value)}
					disabled={disabled || submitting}
					onchange={setBoolean}
					class="h-8"
				/>
			{:else}
				<ValueEditor
					id="rule-value"
					type={valueType}
					{value}
					compact
					disabled={disabled || submitting}
					onchange={onvalue}
					onvalid={onvalid}
				/>
			{/if}
		</div>
	{/if}

	<div class="w-[11rem] shrink-0 space-y-2">
		<div
			class="flex h-4 items-center gap-2 text-[0.65rem] uppercase leading-4 tracking-[0.14em] text-muted-foreground"
		>
			<span>rollout</span>
			<span class="font-mono text-foreground">{percentage || 0}%</span>
		</div>
		<input
			type="range"
			min="0"
			max="100"
			step="1"
			value={percentage}
			disabled={disabled || submitting}
			oninput={(e) => onpercentage((e.currentTarget as HTMLInputElement).value)}
			class="rule-rollout-slider h-8 w-full cursor-pointer disabled:cursor-not-allowed disabled:opacity-50"
			aria-label="rule rollout percentage"
		/>
	</div>

	<div class="w-[10.5rem] shrink-0 space-y-2">
		<label
			for="bucket-by"
			class="block h-4 text-[0.65rem] uppercase leading-4 tracking-[0.14em] text-muted-foreground"
		>
			bucket by · optional
		</label>
		<div class="relative">
			<input
				id="bucket-by"
				bind:this={bucketInput}
				value={bucketBy}
				oninput={handleBucketInput}
				onfocus={() => {
					if (bucketSuggestions.length > 0) bucketAcOpen = true;
				}}
				onkeydown={handleBucketKeydown}
				onblur={() => setTimeout(() => (bucketAcOpen = false), 120)}
				placeholder="user.id"
				autocomplete="off"
				autocapitalize="off"
				autocorrect="off"
				spellcheck="false"
				data-1p-ignore
				data-lpignore="true"
				data-form-type="other"
				disabled={disabled || submitting}
				class="ios-corners-sm flex h-8 w-full border border-border-control bg-background px-2.5 py-1 font-mono text-xs transition-all duration-200 ease-out placeholder:text-muted-foreground/60 focus-visible:border-border-hover focus-visible:bg-surface-subtle focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
			/>
			{#if bucketAcOpen && bucketSuggestions.length > 0}
				<div
					role="listbox"
					class="glass-panel motion-pop absolute left-0 right-0 top-full z-20 mt-1 max-h-64 overflow-y-auto rounded-sm py-1 shadow-lg"
				>
					{#each bucketSuggestions as field, i (field.path)}
						<button
							type="button"
							role="option"
							aria-selected={i === bucketAcIndex}
							tabindex="-1"
							onmousedown={(e) => {
								e.preventDefault();
								void insertBucket(field);
							}}
							onmouseenter={() => (bucketAcIndex = i)}
							class={cn(
								'flex w-full cursor-pointer items-center justify-between gap-3 px-3 py-1.5 text-left font-mono text-xs transition-colors',
								i === bucketAcIndex
									? 'bg-surface-active text-foreground'
									: 'text-muted-foreground hover:text-foreground'
							)}
						>
							<span class="truncate">{field.path}</span>
							<span class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground/70">
								{field.type}
							</span>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.rule-rollout-slider {
		appearance: none;
		background: transparent;
	}

	.rule-rollout-slider::-webkit-slider-runnable-track {
		height: 2px;
		border-radius: 999px;
		background: color-mix(in oklab, var(--color-foreground) 18%, transparent);
	}

	.rule-rollout-slider::-webkit-slider-thumb {
		appearance: none;
		width: 14px;
		height: 14px;
		margin-top: -6px;
		border-radius: 999px;
		background: var(--color-foreground);
		border: 1px solid color-mix(in oklab, var(--color-foreground) 70%, transparent);
	}

	.rule-rollout-slider::-moz-range-track {
		height: 2px;
		border-radius: 999px;
		background: color-mix(in oklab, var(--color-foreground) 18%, transparent);
	}

	.rule-rollout-slider::-moz-range-thumb {
		width: 14px;
		height: 14px;
		border-radius: 999px;
		background: var(--color-foreground);
		border: 1px solid color-mix(in oklab, var(--color-foreground) 70%, transparent);
	}
</style>
