<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		api,
		APIError,
		type Flag,
		type CreateRuleRequest,
		type ContextSchema
	} from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';
	import SectionHeader from '$lib/components/ui/section-header.svelte';
	import RuleEditor from '$lib/components/rule-editor.svelte';
	import ContextPicker from '$lib/components/context-picker.svelte';
	import { Trash2, Plus, Pencil, ArrowUp, ArrowDown } from 'lucide-svelte';
	import type { PageProps } from './$types';

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

	$effect(() => {
		flag = data.flag;
		context = data.context;
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

	async function patch(
		updates: Partial<Pick<Flag, 'enabled' | 'default_value' | 'context_id'>>
	) {
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
				enabled: flag.enabled,
				default_value: flag.default_value,
				context_id: flag.context_id ?? null,
				rules: flag.rules.map((r) => ({ expression: r.expression, rollout: r.rollout }))
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
		if (!confirm(`Delete flag "${flag.key}"? This removes all of its rules.`)) return;
		try {
			await api.deleteFlag(flag.key);
			await goto('/');
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to delete flag';
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
			createError = e instanceof APIError ? e.message : 'Failed to create rule';
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
			editError = e instanceof APIError ? e.message : 'Failed to update rule';
		} finally {
			ruleSubmitting = false;
		}
	}

	async function deleteRule(id: string) {
		const rule = flag.rules.find((r) => r.id === id);
		if (!rule) return;
		if (!confirm(`Delete this rule?\n\n${rule.expression}`)) return;
		const prev = flag.rules;
		pendingRuleId = id;
		flag = { ...flag, rules: prev.filter((r) => r.id !== id) };
		try {
			await api.deleteRule(flag.key, id);
		} catch (e) {
			flag = { ...flag, rules: prev };
			error = e instanceof APIError ? e.message : 'Failed to delete rule';
		} finally {
			pendingRuleId = null;
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
</script>

<div class="space-y-10">
	<a
		href="/"
		class="inline-flex items-center gap-1.5 font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground"
	>
		← all flags
	</a>

	<header class="flex flex-wrap items-start justify-between gap-4">

			<div class="space-y-3">
				<p
					class="font-mono text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground"
				>
					[ flag ]
				</p>
				<h1 class="font-mono text-3xl font-normal tracking-tight sm:text-4xl">
					{flag.key}
				</h1>
			</div>
			<Button variant="destructive" onclick={remove}>
				<Trash2 class="h-3.5 w-3.5" /> delete
			</Button>
		</header>

		<Card class="motion-panel divide-y divide-border/60">
			<div class="flex items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="font-mono text-sm">enabled</p>
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
			<div class="flex items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="font-mono text-sm">default value</p>
					<p class="text-xs text-muted-foreground">returned when no rule matches</p>
				</div>
				<BoolToggle
					value={flag.default_value}
					disabled={saving}
					onchange={(v) => patch({ default_value: v })}
				/>
			</div>
			<div class="flex flex-wrap items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="font-mono text-sm">context</p>
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
			<p class="font-mono text-xs text-destructive">{error}</p>
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

			{#if flag.rules.length === 0 && !creating}
				<Card class="motion-panel p-8 text-center">
					<p
						class="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground"
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
										class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
									>
										[ editing rule #{String(i + 1).padStart(2, '0')} ]
									</p>
									<RuleEditor
										{rule}
										{context}
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
											class="font-mono text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground"
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
											class="flex flex-wrap items-center gap-4 font-mono text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
										>
											<span>
												rollout
												<span class="text-foreground">{rule.rollout.percentage}%</span>
											</span>
											{#if rule.rollout.bucket_by}
												<span>
													bucket by
													<span class="text-foreground">{rule.rollout.bucket_by}</span>
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
											onclick={() => deleteRule(rule.id)}
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
						class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
					>
						[ new rule ]
					</p>
					<RuleEditor
						{context}
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
