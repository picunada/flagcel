<script lang="ts">
	import Button from '$lib/components/ui/button.svelte';
	import FieldLabel from '$lib/components/ui/field-label.svelte';
	import Textarea from '$lib/components/ui/textarea.svelte';
	import { Play, RotateCcw } from 'lucide-svelte';
	import { formatFlagValue } from '$lib/values';
	import type { EvalTrace } from '$lib/api';

	type Props = {
		contextJson: string;
		result: EvalTrace | null;
		error: string | null;
		running: boolean;
		inputId?: string;
		onevaluate: () => void;
		onreset: () => void;
		oninput: () => void;
	};

	let {
		contextJson = $bindable(''),
		result,
		error,
		running,
		inputId = 'playground-context',
		onevaluate,
		onreset,
		oninput
	}: Props = $props();

	function reasonLabel(reason: string): string {
		switch (reason) {
			case 'matched_rule':
				return 'matched rule';
			case 'default_no_match':
				return 'default';
			case 'cel_error':
				return 'cel error';
			default:
				return reason.replaceAll('_', ' ');
		}
	}

	function hasRuleErrors(trace: EvalTrace): boolean {
		return trace.rule_results.some((ruleResult) => Boolean(ruleResult.error));
	}
</script>

<div class="space-y-4">
	<div class="space-y-2">
		<div class="flex items-center justify-between gap-2">
			<FieldLabel for={inputId} class="font-mono">context json</FieldLabel>
			<div class="flex items-center gap-2">
				<Button size="sm" variant="ghost" type="button" onclick={onreset} disabled={running}>
					<RotateCcw class="h-3 w-3" /> reset
				</Button>
				<Button
					size="sm"
					variant="solid"
					type="button"
					onclick={onevaluate}
					disabled={running || contextJson.trim().length === 0}
				>
					<Play class="h-3 w-3" /> {running ? 'evaluating...' : 'evaluate'}
				</Button>
			</div>
		</div>
		<Textarea
			id={inputId}
			bind:value={contextJson}
			{oninput}
			rows={10}
			spellcheck="false"
			class="min-h-[15rem] resize-y bg-surface-faint font-mono leading-6 text-foreground"
		></Textarea>
	</div>

	<div class="space-y-3">
		<FieldLabel class="font-mono">result</FieldLabel>

		{#if error}
			<div class="rounded-sm border border-destructive/30 bg-destructive-surface p-3">
				<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-destructive">
					error
				</p>
				<p class="mt-2 break-words font-mono text-xs text-destructive">{error}</p>
			</div>
		{:else if result}
			<div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
				<div class="min-w-0 rounded-sm border border-border/70 bg-surface-faint p-3">
					<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
						value
					</p>
					<div class="mt-2 min-w-0 space-y-2">
						<p
							class="break-all font-mono text-xs leading-5 text-foreground"
							class:text-destructive={Boolean(result.error) || hasRuleErrors(result)}
						>
							{formatFlagValue(result.value)}
						</p>
						<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
							{result.value_type}
						</p>
					</div>
				</div>
				<div class="min-w-0 rounded-sm border border-border/70 bg-surface-faint p-3">
					<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
						path
					</p>
					<p class="mt-2 break-words font-mono text-xs text-foreground">
						{reasonLabel(result.reason)}
					</p>
				</div>
			</div>

			{#if result.error}
				<div class="rounded-sm border border-destructive/30 bg-destructive-surface p-3">
					<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-destructive">
						cel error
					</p>
					<p class="mt-2 break-words font-mono text-xs text-destructive">
						{result.error}
					</p>
				</div>
			{/if}

			<div
				class={result.matched_rule
					? 'rounded-sm border border-valid-border bg-valid-surface p-3'
					: 'rounded-sm border border-border/70 bg-surface-faint p-3'}
			>
				<p
					class={result.matched_rule
						? 'font-mono text-[0.65rem] uppercase tracking-[0.12em] text-valid-muted'
						: 'font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground'}
				>
					matched rule
				</p>
				{#if result.matched_rule}
					<p class="mt-2 font-mono text-xs text-foreground">
						#{String(result.matched_rule.index + 1).padStart(2, '0')}
					</p>
					<p class="mt-2 break-all font-mono text-xs text-muted-foreground">
						value <span class="text-foreground">{formatFlagValue(result.matched_rule.value)}</span>
					</p>
					<pre
						class="mt-2 max-h-32 overflow-auto whitespace-pre-wrap break-words border-l-2 border-valid-border pl-3 font-mono text-xs text-muted-foreground">{result.matched_rule.expression}</pre>
				{:else}
					<p class="mt-2 font-mono text-xs text-muted-foreground">none</p>
				{/if}
			</div>

			{#if result.bucket}
				<div class="rounded-sm border border-border/70 bg-surface-faint p-3">
					<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
						bucket
					</p>
					<div class="mt-2 grid gap-2 font-mono text-xs sm:grid-cols-2">
						<p class="min-w-0 break-all text-muted-foreground">
							by <span class="text-foreground">{result.bucket.bucket_by}</span>
						</p>
						<p class="min-w-0 break-all text-muted-foreground">
							value
							<span class="text-foreground">
								{result.bucket.missing ? 'missing' : result.bucket.bucket_value}
							</span>
						</p>
						<p class="text-muted-foreground">
							bucket
							<span class="text-foreground">
								{result.bucket.bucket_number ?? 'n/a'}
							</span>
						</p>
						<p class="text-muted-foreground">
							rollout
							<span class="text-foreground">{result.bucket.percentage}%</span>
						</p>
					</div>
				</div>
			{/if}

			{#if result.rule_results.length > 0}
				<div class="rounded-sm border border-border/70 bg-surface-faint p-3">
					<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
						rule trace
					</p>
					<div class="mt-2 space-y-1.5">
						{#each result.rule_results as ruleResult (ruleResult.id || ruleResult.index)}
							<div
								class={ruleResult.matched && !ruleResult.error
									? 'rounded-sm border border-valid-border bg-valid-surface px-2.5 py-2'
									: 'border-t border-border/60 pt-2 first:border-t-0 first:pt-0'}
							>
								<div class="flex items-center justify-between gap-3 font-mono text-xs">
									<span class="text-muted-foreground">
										#{String(ruleResult.index + 1).padStart(2, '0')}
									</span>
									<span
										class={ruleResult.error
											? 'text-destructive'
											: ruleResult.matched
												? 'text-valid'
												: 'text-muted-foreground'}
									>
										{ruleResult.error ? 'error' : ruleResult.matched ? 'match' : 'no match'}
									</span>
								</div>
								{#if ruleResult.error}
									<p class="mt-2 break-words font-mono text-xs text-destructive">
										{ruleResult.error}
									</p>
								{/if}
								<pre
									class="mt-2 max-h-24 overflow-auto whitespace-pre-wrap break-words border-l-2 pl-3 font-mono text-xs text-muted-foreground {ruleResult.matched &&
									!ruleResult.error
										? 'border-valid-border'
										: 'border-border'}">{ruleResult.expression}</pre>
								<p class="mt-2 break-all font-mono text-xs text-muted-foreground">
									value <span class="text-foreground">{formatFlagValue(ruleResult.value)}</span>
								</p>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		{:else}
			<div class="rounded-sm border border-border/70 bg-surface-faint p-6 text-center">
				<p class="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
					[ not evaluated ]
				</p>
			</div>
		{/if}
	</div>
</div>
