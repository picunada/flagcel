<script lang="ts">
    import type {
        ContextSchema,
        CreateRuleRequest,
        Flag,
    } from "$lib/api";
    import RuleEditor from "$lib/components/rule-editor.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import { Progress } from "$lib/components/ui/progress";
    import { formatFlagValue } from "$lib/values";
    import { ArrowDown, ArrowUp, Pencil, Trash2 } from "lucide-svelte";

    type Rule = Flag["rules"][number];

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
        updateRule: (
            id: string,
            form: CreateRuleRequest,
        ) => void | Promise<void>;
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
        cancelEdit,
    }: Props = $props();
</script>

<p class="text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground">
    evaluated top-to-bottom
</p>

{#if flag.rules.length === 0 && !creating}
    <Card class="motion-panel p-8 text-center">
        <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
            [ no rules ]
        </p>
        <p class="mt-3 text-sm text-foreground-softer">
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
                            [ editing rule #{String(i + 1).padStart(2, "0")} ]
                        </p>
                        <RuleEditor
                            {rule}
                            {context}
                            valueType={flag.type}
                            submitting={ruleSubmitting}
                            error={editError}
                            submitLabel="save changes"
                            onsave={(data) => updateRule(rule.id, data)}
                            oncancel={cancelEdit}
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
                                #{String(i + 1).padStart(2, "0")}
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
                                class="overflow-x-auto border-l-2 border-success/40 bg-surface-faint py-2 pl-3 font-mono text-sm text-foreground">{rule.expression}</pre>
                            <div
                                class="flex flex-wrap items-center gap-x-4 gap-y-2 pl-3 text-[0.7rem] uppercase leading-none tracking-[0.12em] text-muted-foreground"
                            >
                                <span class="inline-flex h-4 items-center gap-1">
                                    value
                                    <span class="font-mono text-foreground"
                                        >{formatFlagValue(rule.value)}</span
                                    >
                                </span>
                                <span class="inline-flex h-4 items-center gap-2">
                                    <span>rollout</span>
                                    <Progress
                                        value={rule.rollout.percentage}
                                        max={100}
                                        aria-label={`rule ${i + 1} rollout ${rule.rollout.percentage}%`}
                                        class="h-1.5 w-24 shrink-0 bg-rollout-stack-track-front"
                                    />
                                    <span class="font-mono text-foreground"
                                        >{rule.rollout.percentage}%</span
                                    >
                                </span>
                                {#if rule.rollout.bucket_by}
                                    <span
                                        class="inline-flex h-4 items-center gap-1"
                                    >
                                        bucket by
                                        <span class="font-mono text-foreground"
                                            >{rule.rollout.bucket_by}</span
                                        >
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
        <p class="text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
            [ new rule ]
        </p>
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
