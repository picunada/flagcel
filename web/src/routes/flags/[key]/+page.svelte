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
	import EvalPlayground from '$lib/components/eval-playground.svelte';
	import { formatFlagValue } from '$lib/values';
	import { fly, slide } from 'svelte/transition';
	import { Trash2, Plus, Pencil, ArrowUp, ArrowDown, FlaskConical, X, ChevronDown } from 'lucide-svelte';
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
	let drawerOpen = $state(false);
	let mobileOpen = $state(false);
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
				<div class="flex items-center gap-2">
					<Button
						size="sm"
						variant="ghost"
						class="hidden lg:inline-flex"
						aria-expanded={drawerOpen}
						onclick={() => (drawerOpen = !drawerOpen)}
					>
						<FlaskConical class="h-3 w-3" /> test
					</Button>
					{#if !creating}
						<Button size="sm" onclick={startCreate}>
							<Plus class="h-3 w-3" /> add rule
						</Button>
					{/if}
				</div>
			</div>

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

		<!-- mobile: collapsed playground below the rules -->
		<div class="lg:hidden">
			<Card class="motion-panel overflow-hidden">
				<button
					type="button"
					onclick={() => (mobileOpen = !mobileOpen)}
					aria-expanded={mobileOpen}
					class="flex w-full items-center justify-between gap-3 p-5 text-left transition-colors hover:bg-[rgba(255,255,255,0.02)]"
				>
					<span class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
						[ evaluation playground ]
					</span>
					<ChevronDown
						class="h-4 w-4 text-muted-foreground transition-transform duration-200 {mobileOpen
							? 'rotate-180'
							: ''}"
					/>
				</button>
				{#if mobileOpen}
					<div transition:slide={{ duration: 200 }} class="border-t border-border/60 p-5">
						<EvalPlayground
							inputId="playground-context-mobile"
							bind:contextJson={playgroundContext}
							result={playgroundResult}
							error={playgroundError}
							running={playgroundRunning}
							onevaluate={evaluatePlayground}
							onreset={resetPlayground}
							oninput={() => (playgroundDirty = true)}
						/>
					</div>
				{/if}
			</Card>
		</div>
	</section>
</div>

<!-- desktop: sticky side drawer -->
{#if drawerOpen}
	<aside
		transition:fly={{ x: 24, duration: 200 }}
		class="glass-panel fixed inset-y-0 right-0 z-40 hidden w-[26rem] max-w-[calc(100vw-1.5rem)] flex-col border-l border-[rgba(255,255,255,0.12)] shadow-[0_0_60px_rgba(0,0,0,0.45)] lg:flex"
		aria-label="evaluation playground"
	>
		<div class="flex items-center justify-between gap-3 border-b border-border/60 px-5 pb-4 pt-28">
			<p class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
				[ evaluation playground ]
			</p>
			<button
				type="button"
				aria-label="close playground"
				onclick={() => (drawerOpen = false)}
				class="motion-press rounded-sm p-1 text-muted-foreground transition-colors hover:bg-[rgba(255,255,255,0.04)] hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
			>
				<X class="h-4 w-4" />
			</button>
		</div>
		<div class="flex-1 overflow-y-auto p-5">
			<EvalPlayground
				inputId="playground-context-desktop"
				bind:contextJson={playgroundContext}
				result={playgroundResult}
				error={playgroundError}
				running={playgroundRunning}
				onevaluate={evaluatePlayground}
				onreset={resetPlayground}
				oninput={() => (playgroundDirty = true)}
			/>
		</div>
	</aside>
{/if}

<svelte:window
	onkeydown={(e) => {
		if (e.key === 'Escape' && drawerOpen) drawerOpen = false;
	}}
/>

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
