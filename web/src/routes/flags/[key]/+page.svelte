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
		type EvalTrace,
		type AuditEntry
	} from '$lib/api';
	import { describeChanges } from '$lib/history';
	import BackLink from '$lib/components/ui/back-link.svelte';
	import FlagHistoryList from '$lib/components/flags/flag-history-list.svelte';
	import FlagPlaygroundPanels from '$lib/components/flags/flag-playground-panels.svelte';
	import FlagRulesPanel from '$lib/components/flags/flag-rules-panel.svelte';
	import FlagSettingsCard from '$lib/components/flags/flag-settings-card.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import DestructiveDialog from '$lib/components/ui/destructive-dialog.svelte';
	import PageHeader from '$lib/components/ui/page-header.svelte';
	import Tabs from '$lib/components/ui/tabs.svelte';
	import { Trash2, Plus, FlaskConical, History } from 'lucide-svelte';
	import type { PageProps } from './$types';

	type Rule = Flag['rules'][number];

	let { data }: PageProps = $props();

	let flag = $state<Flag>(untrack(() => data.flag));
	let selectedEnvironment = $state(untrack(() => data.selectedEnvironment));
	let context = $state<ContextSchema | null>(untrack(() => data.context));
	let history = $state<AuditEntry[]>(untrack(() => data.history ?? []));
	let selectedTab = $state<'rules' | 'history'>('rules');
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
		selectedEnvironment = data.selectedEnvironment;
		context = data.context;
		history = data.history ?? [];
		if (!playgroundDirty) {
			playgroundContext = sampleContext(data.context);
		}
	});

	const timeFormatter = new Intl.DateTimeFormat(undefined, {
		month: 'short',
		day: 'numeric',
		year: 'numeric',
		hour: 'numeric',
		minute: '2-digit'
	});

	function formatTimestamp(iso: string): string {
		const date = new Date(iso);
		if (Number.isNaN(date.getTime())) return iso;
		return timeFormatter.format(date);
	}

	// History is ordered newest-first; diff each version against the next-older one.
	const historyView = $derived(
		history.map((entry, i) => ({
			entry,
			changes: describeChanges(history[i + 1]?.snapshot ?? null, entry.snapshot ?? null)
		}))
	);

	// Prefer the human-readable actor labels captured in the audit log over the raw
	// user IDs stored on the flag itself.
	const createdBy = $derived(history.find((e) => e.action === 'created')?.actor_label ?? null);
	const updatedBy = $derived(history[0]?.actor_label ?? null);

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

	// Re-fetch the change history after a successful mutation so new versions
	// appear without a full page reload. Failures keep the existing timeline.
	async function refreshHistory() {
		try {
			history = await api.getFlagAudit(selectedEnvironment.id, flag.key);
		} catch {
			// ignore; the timeline will catch up on the next change or reload
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
			await api.createFlag(selectedEnvironment.id, {
				key: flag.key,
				type: flag.type,
				enabled: flag.enabled,
				default_value: flag.default_value,
				context_id: flag.context_id ?? null,
				rules: flag.rules.map((r) => ({
					id: r.id,
					description: r.description,
					expression: r.expression,
					rollout: r.rollout,
					value: r.value
				}))
			});
			if ('context_id' in updates) {
				await loadContext(flag.context_id ?? null);
			}
			void refreshHistory();
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
			await api.deleteFlag(selectedEnvironment.id, flag.key);
			await goto(`/?environment=${encodeURIComponent(selectedEnvironment.id)}`);
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
			const rule = await api.createRule(selectedEnvironment.id, flag.key, form);
			flag = { ...flag, rules: [...flag.rules, rule] };
			creating = false;
			void refreshHistory();
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
			const updated = await api.updateRule(selectedEnvironment.id, flag.key, id, form);
			flag = {
				...flag,
				rules: flag.rules.map((r) => (r.id === id ? updated : r))
			};
			editingRuleId = null;
			void refreshHistory();
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
			await api.deleteRule(selectedEnvironment.id, flag.key, rule.id);
			deleteRuleOpen = false;
			deleteRuleTarget = null;
			void refreshHistory();
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
				selectedEnvironment.id,
				flag.key,
				next.map((r) => r.id)
			);
			void refreshHistory();
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
			playgroundResult = await api.evaluateFlag(selectedEnvironment.id, flag.key, parsed as Record<string, unknown>);
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

	function cancelCreate() {
		creating = false;
		createError = null;
	}

	function cancelEdit() {
		editingRuleId = null;
		editError = null;
	}

	function markPlaygroundDirty() {
		playgroundDirty = true;
	}

</script>

<div class="space-y-10">
	<BackLink
		href={`/?environment=${encodeURIComponent(selectedEnvironment.id)}`}
		label="all flags"
	/>

	<PageHeader eyebrow="[ flag ]" title={flag.key} titleClass="font-mono tracking-tight">
		{#snippet actions()}
		<Button
			variant="destructive"
			onclick={() => {
				deleteFlagError = null;
				deleteFlagOpen = true;
			}}
		>
			<Trash2 class="h-3.5 w-3.5" /> delete
		</Button>
		{/snippet}
	</PageHeader>

		<FlagSettingsCard
			{flag}
			{saving}
			{createdBy}
			{updatedBy}
			{formatTimestamp}
			{patch}
			{updateDefaultValue}
		/>

		{#if error}
			<p class="text-xs text-destructive">{error}</p>
		{/if}

		<section class="space-y-4">
			<div class="flex flex-wrap items-center justify-between gap-3 border-b border-border/60">
				<Tabs
					bind:value={selectedTab}
					items={[
						{ value: 'rules', label: 'rules' },
						{
							value: 'history',
							label: 'history',
							icon: History,
							count: history.length
						}
					]}
				/>
				{#if selectedTab === 'rules'}
					<div class="flex items-center gap-2 pb-2">
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
				{/if}
			</div>

		{#if selectedTab === 'rules'}
			<FlagRulesPanel
				{flag}
				{context}
				{creating}
				{editingRuleId}
				{ruleSubmitting}
				{createError}
				{editError}
				{pendingRuleId}
				{createRule}
				{updateRule}
				{move}
				{startEdit}
				{requestDeleteRule}
				{cancelCreate}
				{cancelEdit}
			/>

			<FlagPlaygroundPanels
				bind:drawerOpen
				bind:mobileOpen
				bind:playgroundContext
				{playgroundResult}
				{playgroundError}
				{playgroundRunning}
				{evaluatePlayground}
				{resetPlayground}
				markDirty={markPlaygroundDirty}
			/>
		{:else}
			<FlagHistoryList {historyView} {formatTimestamp} />
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
