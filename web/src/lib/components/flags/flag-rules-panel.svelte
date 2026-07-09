<script lang="ts">
	import type { ContextSchema, CreateRuleRequest, Flag } from '$lib/api';
	import RuleEditor from '$lib/components/rule-editor.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import { Progress } from '$lib/components/ui/progress';
	import { cn } from '$lib/utils';
	import { formatFlagValue } from '$lib/values';
	import { ArrowDown, ArrowUp, Pencil, Trash2 } from 'lucide-svelte';

	type Rule = Flag['rules'][number];

	type Props = {
		flag: Flag;
		context: ContextSchema | null;
		creating: boolean;
		editingRuleId: string | null;
		ruleSubmitting: boolean;
		createError: string | null;
		editError: string | null;
		pendingRuleId: string | null;
		createRule: (form: CreateRuleRequest) => void | Promise<void>;
		updateRule: (id: string, form: CreateRuleRequest) => void | Promise<void>;
		move: (index: number, direction: -1 | 1) => void | Promise<void>;
		startEdit: (id: string) => void;
		requestDeleteRule: (rule: Rule) => void;
		cancelCreate: () => void;
		cancelEdit: () => void;
	};

	let {
		flag,
		context,
		creating,
		editingRuleId,
		ruleSubmitting,
		createError,
		editError,
		pendingRuleId,
		createRule,
		updateRule,
		move,
		startEdit,
		requestDeleteRule,
		cancelCreate,
		cancelEdit
	}: Props = $props();
</script>

<div class="space-y-3">
	<div class="flex items-center justify-between gap-3">
		<p class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
			evaluated top-to-bottom
		</p>
		{#if flag.rules.length > 0}
			<p class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
				{flag.rules.length} {flag.rules.length === 1 ? 'rule' : 'rules'}
			</p>
		{/if}
	</div>

	{#if flag.rules.length === 0 && !creating}
		<Card class="motion-panel border-dashed p-8 text-center">
			<p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">[ no rules ]</p>
			<p class="mt-3 text-sm text-foreground-softer">
				Requests fall through to the default value.
			</p>
		</Card>
	{:else}
		<div class="motion-list space-y-2">
			{#each flag.rules as rule, i (rule.id)}
				<Card
					class={cn(
						'p-0',
						editingRuleId === rule.id && 'border-border-strong bg-surface-subtle'
					)}
				>
					<div class="flex gap-3 p-4 sm:p-5">
						<div class="flex w-8 shrink-0 flex-col items-center gap-1">
							<button
								type="button"
								aria-label="move rule up"
								disabled={i === 0 || pendingRuleId !== null || editingRuleId === rule.id}
								onclick={() => move(i, -1)}
								class="cursor-pointer text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
							>
								<ArrowUp class="h-3.5 w-3.5" />
							</button>
							<div class="font-mono text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground">
								#{String(i + 1).padStart(2, '0')}
							</div>
							<button
								type="button"
								aria-label="move rule down"
								disabled={i === flag.rules.length - 1 ||
									pendingRuleId !== null ||
									editingRuleId === rule.id}
								onclick={() => move(i, 1)}
								class="cursor-pointer text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
							>
								<ArrowDown class="h-3.5 w-3.5" />
							</button>
						</div>

						<div class="min-w-0 flex-1">
							{#if editingRuleId === rule.id}
								<RuleEditor
									{rule}
									{context}
									valueType={flag.type}
									submitting={ruleSubmitting}
									error={editError}
									submitLabel="save"
									onsave={(data) => updateRule(rule.id, data)}
									oncancel={cancelEdit}
								/>
							{:else}
								<div class="flex min-w-0 flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
									<div class="min-w-0 flex-1 space-y-3">
										<pre
											class="overflow-x-auto border-l-2 border-success/40 bg-surface-code py-2 pl-3 font-mono text-sm leading-6 text-foreground"
										>{rule.expression}</pre>
										<div
											class="flex flex-wrap items-center gap-x-4 gap-y-2 pl-3 text-[0.65rem] uppercase leading-none tracking-[0.12em] text-muted-foreground"
										>
											<span class="inline-flex h-4 items-center gap-1">
												value
												<span class="font-mono text-foreground">
													{formatFlagValue(rule.value)}
												</span>
											</span>
											<span class="inline-flex h-4 items-center gap-2">
												<span>rollout</span>
												<Progress
													value={rule.rollout.percentage}
													max={100}
													aria-label={`rule ${i + 1} rollout ${rule.rollout.percentage}%`}
													class="h-1.5 w-24 shrink-0 bg-rollout-stack-track-front"
												/>
												<span class="font-mono text-foreground">
													{rule.rollout.percentage}%
												</span>
											</span>
											<span class="inline-flex h-4 items-center gap-1">
												bucket
												<span class="font-mono text-foreground">
													{rule.rollout.bucket_by || 'none'}
												</span>
											</span>
										</div>
									</div>
									<div class="flex shrink-0 items-center gap-1">
										<Button
											size="sm"
											variant="ghost"
											onclick={() => startEdit(rule.id)}
											disabled={pendingRuleId !== null}
										>
											<Pencil class="h-3 w-3" /> edit
										</Button>
										<Button
											size="sm"
											variant="destructive"
											disabled={pendingRuleId === rule.id || pendingRuleId !== null}
											onclick={() => requestDeleteRule(rule)}
											aria-label="delete rule"
										>
											<Trash2 class="h-3 w-3" />
										</Button>
									</div>
								</div>
							{/if}
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{/if}

	<Card class="border-dashed bg-transparent p-4">
		<div class="flex flex-wrap items-center gap-3">
			<span class="font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
				no rule matches → default
			</span>
			<span class="font-mono text-xs font-semibold text-foreground">
				{formatFlagValue(flag.default_value)}
			</span>
		</div>
	</Card>

	{#if creating}
		<Card class="motion-panel border-border-strong bg-surface-subtle p-4 sm:p-5">
			<RuleEditor
				{context}
				valueType={flag.type}
				submitting={ruleSubmitting}
				error={createError}
				submitLabel="add rule"
				onsave={createRule}
				oncancel={cancelCreate}
			/>
		</Card>
	{/if}
</div>
