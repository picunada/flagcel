<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		api,
		APIError,
		type Flag,
		type FlagValue,
		type CreateRuleRequest,
		type ContextSchema,
		type EvalTrace
	} from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import Badge from '$lib/components/ui/badge.svelte';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';
	import DestructiveDialog from '$lib/components/ui/destructive-dialog.svelte';
	import SectionHeader from '$lib/components/ui/section-header.svelte';
	import RuleEditor from '$lib/components/rule-editor.svelte';
	import ContextPicker from '$lib/components/context-picker.svelte';
	import ValueEditor from '$lib/components/value-editor.svelte';
	import { formatFlagValue, valueBadgeVariant } from '$lib/values';
	import { Trash2, Plus, Pencil, ArrowUp, ArrowDown, Play, RotateCcw } from 'lucide-svelte';
	import type { PageProps } from './$types';

	type Rule = Flag['rules'][number];

	let { data }: PageProps = $props();

	let flag = $state<Flag>(untrack(() => data.flag));
	let context = $state<ContextSchema | null>(untrack(() => data.context));
	let error = $state<string | null>(null);
	let saving = $state(false);

	let creating = $state(false);
	let createError = $state<string | null>(null);
	let editingRuleId = $state<string | null>(null);
	let editError = $state<string | null>(null);
	let ruleSubmitting = $state(false);
	let pendingRuleId = $state<string | null>(null);
	let playgroundContext = $state(untrack(() => sampleContext(data.context)));
	let playgroundDirty = $state(false);
	let playgroundResult = $state<EvalTrace | null>(null);
	let playgroundError = $state<string | null>(null);
	let playgroundRunning = $state(false);
	let deleteFlagOpen = $state(false);
	let deleteFlagSubmitting = $state(false);
	let deleteFlagError = $state<string | null>(null);
	let deleteRuleOpen = $state(false);
	let deleteRuleTarget = $state<Rule | null>(null);
	let deleteRuleSubmitting = $state(false);
	let deleteRuleError = $state<string | null>(null);

	$effect(() => {
		flag = data.flag;
		context = data.context;
		if (!playgroundDirty) {
			playgroundContext = sampleContext(data.context);
		}
	});

	async function loadContext(id: string | null) {
		if (!id) {
			context = null;
			return;
		}
		try {
			context = await api.getContext(id);
		} catch {
			context = null;
		}
	}

	async function patch(updates: Partial<Pick<Flag, 'enabled' | 'default_value' | 'context_id'>>) {
		const prev = {
			enabled: flag.enabled,
			default_value: flag.default_value,
			context_id: flag.context_id
		};
		flag = { ...flag, ...updates };
		saving = true;
		error = null;
		try {
			await api.createFlag({
				key: flag.key,
				type: flag.type,
				enabled: flag.enabled,
				default_value: flag.default_value,
				context_id: flag.context_id ?? null,
				rules: flag.rules.map((r) => ({
					expression: r.expression,
					rollout: r.rollout,
					value: r.value
				}))
			});
			if ('context_id' in updates) {
				await loadContext(flag.context_id ?? null);
			}
		} catch (e) {
			flag = { ...flag, ...prev };
			error = e instanceof APIError ? e.message : 'Failed to save flag';
		} finally {
			saving = false;
		}
	}

	async function remove() {
		deleteFlagSubmitting = true;
		deleteFlagError = null;
		try {
			await api.deleteFlag(flag.key);
			await goto('/');
		} catch (e) {
			deleteFlagError = e instanceof APIError ? e.message : 'Failed to delete flag';
		} finally {
			deleteFlagSubmitting = false;
		}
	}

	async function createRule(form: CreateRuleRequest) {
		ruleSubmitting = true;
		createError = null;
		try {
			const rule = await api.createRule(flag.key, form);
			flag = { ...flag, rules: [...flag.rules, rule] };
			creating = false;
		} catch (e) {
			createError = formatRuleError(e, 'Failed to create rule');
		} finally {
			ruleSubmitting = false;
		}
	}

	async function updateRule(id: string, form: CreateRuleRequest) {
		ruleSubmitting = true;
		editError = null;
		try {
			const updated = await api.updateRule(flag.key, id, form);
			flag = {
				...flag,
				rules: flag.rules.map((r) => (r.id === id ? updated : r))
			};
			editingRuleId = null;
		} catch (e) {
			editError = formatRuleError(e, 'Failed to update rule');
		} finally {
			ruleSubmitting = false;
		}
	}

	function formatRuleError(e: unknown, fallback: string) {
		if (!(e instanceof APIError)) return fallback;
		if (e.details?.length) {
			return e.details.map((detail) => detail.message).join('\n');
		}
		return e.message;
	}

	function requestDeleteRule(rule: Rule) {
		deleteRuleTarget = rule;
		deleteRuleError = null;
		deleteRuleOpen = true;
	}

	async function deleteRule() {
		const rule = deleteRuleTarget;
		if (!rule) return;
		const prev = flag.rules;
		pendingRuleId = rule.id;
		deleteRuleSubmitting = true;
		deleteRuleError = null;
		flag = { ...flag, rules: prev.filter((r) => r.id !== rule.id) };
		try {
			await api.deleteRule(flag.key, rule.id);
			deleteRuleOpen = false;
			deleteRuleTarget = null;
		} catch (e) {
			flag = { ...flag, rules: prev };
			deleteRuleError = e instanceof APIError ? e.message : 'Failed to delete rule';
		} finally {
			pendingRuleId = null;
			deleteRuleSubmitting = false;
		}
	}

	async function move(index: number, direction: -1 | 1) {
		const target = index + direction;
		if (target < 0 || target >= flag.rules.length) return;
		const prev = flag.rules;
		const next = [...prev];
		[next[index], next[target]] = [next[target], next[index]];
		flag = { ...flag, rules: next };
		try {
			await api.reorderRules(
				flag.key,
				next.map((r) => r.id)
			);
		} catch (e) {
			flag = { ...flag, rules: prev };
			error = e instanceof APIError ? e.message : 'Failed to reorder rules';
		}
	}

	function startCreate() {
		editingRuleId = null;
		editError = null;
		createError = null;
		creating = true;
	}

	function startEdit(id: string) {
		creating = false;
		createError = null;
		editError = null;
		editingRuleId = id;
	}

	function sampleContext(ctx: ContextSchema | null): string {
		const sample: Record<string, unknown> = {};
		const fields = ctx?.fields ?? [];
		if (fields.length === 0) {
			return JSON.stringify(
				{
					user: {
						id: 'u_123',
						country: 'US'
					}
				},
				null,
				2
			);
		}

		for (const field of fields) {
			setPath(sample, field.path, sampleValue(field.type));
		}
		return JSON.stringify(sample, null, 2);
	}

	function sampleValue(type: ContextSchema['fields'][number]['type']): unknown {
		switch (type) {
			case 'int':
				return 42;
			case 'double':
				return 42.5;
			case 'bool':
				return true;
			case 'timestamp':
				return '2026-01-01T00:00:00Z';
			case 'list':
				return [];
			case 'map':
				return {};
			case 'string':
			default:
				return 'example';
		}
	}

	function setPath(target: Record<string, unknown>, path: string, value: unknown) {
		const parts = path.split('.').filter(Boolean);
		if (parts.length === 0) return;

		let cursor: Record<string, unknown> = target;
		for (const part of parts.slice(0, -1)) {
			const next = cursor[part];
			if (!next || typeof next !== 'object' || Array.isArray(next)) {
				cursor[part] = {};
			}
			cursor = cursor[part] as Record<string, unknown>;
		}
		cursor[parts[parts.length - 1]] = value;
	}

	async function evaluatePlayground() {
		playgroundError = null;
		playgroundResult = null;

		let parsed: unknown;
		try {
			parsed = JSON.parse(playgroundContext);
		} catch (e) {
			playgroundError = e instanceof Error ? e.message : 'Invalid JSON';
			return;
		}
		if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
			playgroundError = 'Context must be a JSON object';
			return;
		}

		playgroundRunning = true;
		try {
			playgroundResult = await api.evaluateFlag(flag.key, parsed as Record<string, unknown>);
		} catch (e) {
			playgroundError = e instanceof APIError ? e.message : 'Failed to evaluate flag';
		} finally {
			playgroundRunning = false;
		}
	}

	function resetPlayground() {
		playgroundContext = sampleContext(context);
		playgroundDirty = false;
		playgroundResult = null;
		playgroundError = null;
	}

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

	async function updateDefaultValue(value: FlagValue) {
		await patch({ default_value: value });
	}
</script>

<div class="space-y-10">
	<a
		href="/"
		class="inline-flex items-center gap-1.5 text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground"
	>
		← all flags
	</a>

	<header class="flex flex-wrap items-start justify-between gap-4">
		<div class="space-y-3">
			<p
				class="text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground"
			>
				[ flag ]
			</p>
			<h1 class="font-mono text-3xl font-normal tracking-tight sm:text-4xl">
				{flag.key}
			</h1>
		</div>
		<Button
			variant="destructive"
			onclick={() => {
				deleteFlagError = null;
				deleteFlagOpen = true;
			}}
		>
			<Trash2 class="h-3.5 w-3.5" /> delete
		</Button>
	</header>

		<Card class="motion-panel divide-y divide-border/60">
			<div class="flex items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="text-sm">type</p>
					<p class="text-xs text-muted-foreground">value shape returned by evaluation</p>
				</div>
				<Badge variant="muted">{flag.type}</Badge>
			</div>
			<div class="flex items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="text-sm">enabled</p>
					<p class="text-xs text-muted-foreground">
						when off, the default value is returned for every request
					</p>
				</div>
				<BoolToggle
					value={flag.enabled}
					disabled={saving}
					onchange={(v) => patch({ enabled: v })}
				/>
			</div>
			<div class="flex flex-wrap items-start justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="text-sm">default value</p>
					<p class="text-xs text-muted-foreground">returned when no rule matches</p>
				</div>
				<div class="min-w-48 max-w-full flex-1 sm:flex-none sm:basis-80">
					<ValueEditor
						type={flag.type}
						value={flag.default_value}
						id="default-value"
						align="end"
						disabled={saving}
						onchange={updateDefaultValue}
					/>
				</div>
			</div>
			<div class="flex flex-wrap items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="text-sm">context</p>
					<p class="text-xs text-muted-foreground">
						selects the evaluation shape used for autocomplete in rules
					</p>
				</div>
				<div class="min-w-48">
					<ContextPicker
						value={flag.context_id ?? null}
						disabled={saving}
						onchange={(v) => patch({ context_id: v })}
					/>
				</div>
			</div>
		</Card>

		{#if error}
			<p class="text-xs text-destructive">{error}</p>
		{/if}

		<section class="space-y-4">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<SectionHeader>rules · evaluated top-to-bottom</SectionHeader>
				{#if !creating}
					<Button size="sm" onclick={startCreate}>
						<Plus class="h-3 w-3" /> add rule
					</Button>
				{/if}
			</div>

			<Card class="motion-panel space-y-4 p-5">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<p
						class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
					>
						[ evaluation playground ]
					</p>
					<div class="flex items-center gap-2">
						<Button
							size="sm"
							variant="ghost"
							type="button"
							onclick={resetPlayground}
							disabled={playgroundRunning}
						>
							<RotateCcw class="h-3 w-3" /> reset
						</Button>
						<Button
							size="sm"
							variant="solid"
							type="button"
							onclick={evaluatePlayground}
							disabled={playgroundRunning || playgroundContext.trim().length === 0}
						>
							<Play class="h-3 w-3" /> {playgroundRunning ? 'evaluating...' : 'evaluate'}
						</Button>
					</div>
				</div>

				<div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,0.85fr)]">
					<div class="space-y-2">
						<label
							for="playground-context"
							class="font-mono text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
						>
							context json
						</label>
						<textarea
							id="playground-context"
							bind:value={playgroundContext}
							oninput={() => (playgroundDirty = true)}
							rows="12"
							spellcheck="false"
							class="min-h-72 w-full resize-y rounded-sm border border-input bg-[rgba(255,255,255,0.02)] px-3 py-2 font-mono text-sm leading-6 text-foreground transition-colors placeholder:text-muted-foreground/60 focus-visible:border-[rgba(255,255,255,0.36)] focus-visible:outline-none"
						></textarea>
					</div>

					<div class="space-y-3">
						<p
							class="font-mono text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
						>
							result
						</p>

						{#if playgroundError}
							<div class="rounded-sm border border-destructive/30 bg-[rgba(255,107,107,0.05)] p-3">
								<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-destructive">
									error
								</p>
								<p class="mt-2 break-words font-mono text-xs text-destructive">{playgroundError}</p>
							</div>
						{:else if playgroundResult}
							<div class="grid grid-cols-2 gap-2">
								<div class="rounded-sm border border-border/70 bg-[rgba(255,255,255,0.02)] p-3">
									<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
										value
									</p>
									<div class="mt-2">
										<Badge
											dot
											variant={playgroundResult.error
												? 'destructive'
												: valueBadgeVariant(playgroundResult.value)}
										>
											{formatFlagValue(playgroundResult.value)}
										</Badge>
										<p class="mt-2 font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
											{playgroundResult.value_type}
										</p>
									</div>
								</div>
								<div class="rounded-sm border border-border/70 bg-[rgba(255,255,255,0.02)] p-3">
									<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
										path
									</p>
									<p class="mt-2 font-mono text-xs text-foreground">
										{reasonLabel(playgroundResult.reason)}
									</p>
								</div>
							</div>

							{#if playgroundResult.error}
								<div class="rounded-sm border border-destructive/30 bg-[rgba(255,107,107,0.05)] p-3">
									<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-destructive">
										cel error
									</p>
									<p class="mt-2 break-words font-mono text-xs text-destructive">
										{playgroundResult.error}
									</p>
								</div>
							{/if}

							<div class="rounded-sm border border-border/70 bg-[rgba(255,255,255,0.02)] p-3">
								<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
									matched rule
								</p>
								{#if playgroundResult.matched_rule}
									<p class="mt-2 font-mono text-xs text-foreground">
										#{String(playgroundResult.matched_rule.index + 1).padStart(2, '0')}
									</p>
									<p class="mt-2 font-mono text-xs text-muted-foreground">
										value <span class="text-foreground">{formatFlagValue(playgroundResult.matched_rule.value)}</span>
									</p>
									<pre
										class="mt-2 max-h-32 overflow-auto whitespace-pre-wrap break-words border-l-2 border-success/40 pl-3 font-mono text-xs text-muted-foreground">{playgroundResult.matched_rule.expression}</pre>
								{:else}
									<p class="mt-2 font-mono text-xs text-muted-foreground">none</p>
								{/if}
							</div>

							{#if playgroundResult.bucket}
								<div class="rounded-sm border border-border/70 bg-[rgba(255,255,255,0.02)] p-3">
									<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
										bucket
									</p>
									<div class="mt-2 grid gap-2 font-mono text-xs sm:grid-cols-2">
										<p class="text-muted-foreground">
											by <span class="text-foreground">{playgroundResult.bucket.bucket_by}</span>
										</p>
										<p class="text-muted-foreground">
											value
											<span class="text-foreground">
												{playgroundResult.bucket.missing
													? 'missing'
													: playgroundResult.bucket.bucket_value}
											</span>
										</p>
										<p class="text-muted-foreground">
											bucket
											<span class="text-foreground">
												{playgroundResult.bucket.bucket_number ?? 'n/a'}
											</span>
										</p>
										<p class="text-muted-foreground">
											rollout
											<span class="text-foreground">{playgroundResult.bucket.percentage}%</span>
										</p>
									</div>
								</div>
							{/if}

							{#if playgroundResult.rule_results.length > 0}
								<div class="rounded-sm border border-border/70 bg-[rgba(255,255,255,0.02)] p-3">
									<p class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
										rule trace
									</p>
									<div class="mt-2 space-y-1.5">
										{#each playgroundResult.rule_results as result (result.id || result.index)}
											<div class="flex items-center justify-between gap-3 font-mono text-xs">
												<span class="text-muted-foreground">
													#{String(result.index + 1).padStart(2, '0')}
												</span>
												<span
													class={result.error
														? 'text-destructive'
														: result.matched
															? 'text-success'
															: 'text-muted-foreground'}
												>
													{result.error ? 'error' : result.matched ? 'match' : 'no match'}
												</span>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						{:else}
							<div class="rounded-sm border border-border/70 bg-[rgba(255,255,255,0.02)] p-6 text-center">
								<p class="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
									[ not evaluated ]
								</p>
							</div>
						{/if}
					</div>
				</div>
			</Card>

			{#if flag.rules.length === 0 && !creating}
				<Card class="motion-panel p-8 text-center">
					<p
						class="text-xs uppercase tracking-[0.14em] text-muted-foreground"
					>
						[ no rules ]
					</p>
					<p class="mt-3 text-sm text-[rgba(255,255,255,0.7)]">
						Requests fall through to the default value.
					</p>
				</Card>
			{:else}
				<div class="motion-list space-y-2">
					{#each flag.rules as rule, i (rule.id)}
						<Card class="p-5">
							{#if editingRuleId === rule.id}
								<div class="motion-panel space-y-4">
									<p
										class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
									>
										[ editing rule #{String(i + 1).padStart(2, '0')} ]
									</p>
									<RuleEditor
										{rule}
										{context}
										valueType={flag.type}
										submitting={ruleSubmitting}
										error={editError}
										submitLabel="save changes"
										onsave={(data) => updateRule(rule.id, data)}
										oncancel={() => {
											editingRuleId = null;
											editError = null;
										}}
									/>
								</div>
							{:else}
								<div class="flex items-start gap-4">
									<div class="flex flex-col items-center gap-1">
										<button
											type="button"
											aria-label="move up"
											disabled={i === 0 || pendingRuleId !== null}
											onclick={() => move(i, -1)}
											class="text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
										>
											<ArrowUp class="h-3.5 w-3.5" />
										</button>
										<div
											class="text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground"
										>
											#{String(i + 1).padStart(2, '0')}
										</div>
										<button
											type="button"
											aria-label="move down"
											disabled={i === flag.rules.length - 1 ||
												pendingRuleId !== null}
											onclick={() => move(i, 1)}
											class="text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
										>
											<ArrowDown class="h-3.5 w-3.5" />
										</button>
									</div>
									<div class="min-w-0 flex-1 space-y-3">
										<pre
											class="overflow-x-auto border-l-2 border-success/40 bg-[rgba(255,255,255,0.02)] py-2 pl-3 font-mono text-sm text-foreground">{rule.expression}</pre>
										<div
											class="flex flex-wrap items-center gap-4 text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
										>
											<span>
												value
												<span class="font-mono text-foreground">{formatFlagValue(rule.value)}</span>
											</span>
											<span>
												rollout
												<span class="text-foreground">{rule.rollout.percentage}%</span>
											</span>
											{#if rule.rollout.bucket_by}
												<span>
													bucket by
													<span class="font-mono text-foreground">{rule.rollout.bucket_by}</span>
												</span>
											{/if}
										</div>
									</div>
									<div class="flex shrink-0 items-center gap-1">
										<Button
											size="sm"
											variant="ghost"
											onclick={() => startEdit(rule.id)}
										>
											<Pencil class="h-3 w-3" /> edit
										</Button>
										<Button
											size="sm"
											variant="destructive"
											disabled={pendingRuleId === rule.id}
											onclick={() => requestDeleteRule(rule)}
										>
											<Trash2 class="h-3 w-3" />
										</Button>
									</div>
								</div>
							{/if}
						</Card>
					{/each}
				</div>
			{/if}

			{#if creating}
				<Card class="motion-panel space-y-4 p-5">
					<p
						class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
					>
						[ new rule ]
					</p>
					<RuleEditor
						{context}
						valueType={flag.type}
						submitting={ruleSubmitting}
						error={createError}
						submitLabel="add rule"
						onsave={createRule}
						oncancel={() => {
							creating = false;
							createError = null;
						}}
					/>
				</Card>
		{/if}
	</section>
</div>

<DestructiveDialog
	bind:open={deleteFlagOpen}
	title="Delete flag"
	description="This permanently deletes the flag and all of its rules."
	confirmationValue={flag.key}
	actionLabel="delete flag"
	submitting={deleteFlagSubmitting}
	error={deleteFlagError}
	onconfirm={remove}
/>

<DestructiveDialog
	bind:open={deleteRuleOpen}
	title="Delete rule"
	description="This removes the rule from the flag evaluation order."
	details={deleteRuleTarget?.expression}
	actionLabel="delete rule"
	submitting={deleteRuleSubmitting}
	error={deleteRuleError}
	onconfirm={deleteRule}
	oncancel={() => {
		deleteRuleTarget = null;
		deleteRuleError = null;
	}}
/>
