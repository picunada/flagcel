<script lang="ts">
	import { untrack } from 'svelte';
	import Button from '$lib/components/ui/button.svelte';
	import RuleControlsRow from '$lib/components/rule-controls-row.svelte';
	import RuleExpressionInput from '$lib/components/rule-expression-input.svelte';
	import ValueEditor from '$lib/components/value-editor.svelte';
	import { cn } from '$lib/utils';
	import type { ContextSchema, CreateRuleRequest, FlagValue, Rule, ValueType } from '$lib/api';
	import { defaultValueForType } from '$lib/values';

	type Props = {
		rule?: Rule;
		context?: ContextSchema | null;
		valueType?: ValueType;
		submitting?: boolean;
		error?: string | null;
		onsave: (data: CreateRuleRequest) => void | Promise<void>;
		oncancel?: () => void;
		submitLabel?: string;
		class?: string;
	};

	let {
		rule,
		context = null,
		valueType = 'boolean',
		submitting = false,
		error = null,
		onsave,
		oncancel,
		submitLabel = 'save rule',
		class: className
	}: Props = $props();

	let expression = $state(untrack(() => rule?.expression ?? ''));
	let percentage = $state(untrack(() => String(rule?.rollout.percentage ?? 100)));
	let bucketBy = $state(untrack(() => rule?.rollout.bucket_by ?? ''));
	let value = $state<FlagValue>(untrack(() => rule?.value ?? defaultValueForType(valueType, true)));
	let valueValid = $state(true);

	const isJsonValue = $derived(valueType === 'json');
	const parsedPercentage = $derived(Number(percentage));
	const canSubmit = $derived(
		expression.trim().length > 0 &&
			Number.isFinite(parsedPercentage) &&
			parsedPercentage >= 0 &&
			parsedPercentage <= 100 &&
			valueValid &&
			!submitting
	);

	async function save() {
		if (!canSubmit) return;
		await onsave({
			expression: expression.trim(),
			rollout: {
				percentage: Math.round(parsedPercentage),
				...(bucketBy.trim() ? { bucket_by: bucketBy.trim() } : {})
			},
			value
		});
	}

	function submit(e: SubmitEvent) {
		e.preventDefault();
		void save();
	}
</script>

<form onsubmit={submit} class={cn('space-y-4', className)}>
	<div class="flex items-center justify-between gap-3">
		<p class="font-mono text-[0.65rem] uppercase tracking-[0.16em] text-muted-foreground">
			{rule ? '[ edit rule ]' : '[ new rule ]'} · cel
			{#if context}
				<span class="text-muted-foreground/70"> · {context.name}</span>
			{/if}
		</p>
		<p class="hidden font-mono text-[0.6rem] uppercase tracking-[0.14em] text-muted-foreground/70 sm:block">
			⌘↵ save · esc cancel
		</p>
	</div>

	{#if isJsonValue}
		<RuleExpressionInput
			value={expression}
			fields={context?.fields ?? []}
			contextName={context?.name ?? null}
			{submitting}
			oninput={(next) => (expression = next)}
			onsave={save}
			oncancel={oncancel}
		>
			{#snippet aside()}
				<div class="flex min-h-0 min-w-0 flex-col gap-2">
					<div class="h-4 shrink-0 text-[0.65rem] uppercase leading-4 tracking-[0.14em] text-muted-foreground">
						value
					</div>
					<div class="min-h-0 flex-1">
						<ValueEditor
							id="rule-value"
							type="json"
							{value}
							fill
							disabled={submitting}
							onchange={(next) => (value = next)}
							onvalid={(valid) => (valueValid = valid)}
						/>
					</div>
				</div>
			{/snippet}
		</RuleExpressionInput>
	{:else}
		<RuleExpressionInput
			value={expression}
			fields={context?.fields ?? []}
			contextName={context?.name ?? null}
			{submitting}
			oninput={(next) => (expression = next)}
			onsave={save}
			oncancel={oncancel}
		/>
	{/if}

	<RuleControlsRow
		{valueType}
		{value}
		{percentage}
		{bucketBy}
		fields={context?.fields ?? []}
		{submitting}
		hideValue={isJsonValue}
		onvalue={(next) => (value = next)}
		onvalid={(valid) => (valueValid = valid)}
		onpercentage={(next) => (percentage = next)}
		onbucket={(next) => (bucketBy = next)}
	/>

	{#if error}
		<p class="whitespace-pre-line text-xs text-destructive">{error}</p>
	{/if}

	<div class="flex justify-end gap-2">
		{#if oncancel}
			<Button variant="ghost" size="sm" type="button" onclick={oncancel} disabled={submitting}>
				cancel
			</Button>
		{/if}
		<Button variant="solid" size="sm" type="submit" disabled={!canSubmit}>
			{submitting ? 'saving...' : submitLabel}
		</Button>
	</div>
</form>
