<script lang="ts">
    import { tick, untrack } from "svelte";
    import { beforeNavigate, goto, invalidateAll } from "$app/navigation";
    import {
        AlertTriangle,
        Check,
        FileJson,
        LockKeyhole,
        Pencil,
        Plus,
        RotateCcw,
        Save,
        Trash2,
        X,
        Zap,
    } from "lucide-svelte";
    import { api, APIError, type ContextField, type ContextSchema } from "$lib/api";
    import {
        parseSchema,
        samplePayload,
        serializeSchema,
        validatePayload,
    } from "$lib/context-schema";
    import { formatJSON } from "$lib/code-highlighting";
    import ContextDetailsDialog from "$lib/components/contexts/context-details-dialog.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import CodeTextarea from "$lib/components/ui/code-textarea.svelte";
    import DestructiveDialog from "$lib/components/ui/destructive-dialog.svelte";
    import { cn } from "$lib/utils";
    import type { PageProps } from "./$types";

    let { data }: PageProps = $props();

    let schema: ContextSchema = $state(untrack(() => data.schema));
    let schemaText = $state(untrack(() => serializeSchema(data.schema.fields)));
    let payloadText = $state("");
    let saveError = $state<string | null>(null);
    let saveDetails: string[] = $state([]);
    let submitting = $state(false);
    let deleteOpen = $state(false);
    let deleting = $state(false);
    let deleteError = $state<string | null>(null);
    let editOpen = $state(false);
    let editing = $state(false);
    let editError = $state<string | null>(null);

    $effect(() => {
        schema = data.schema;
        schemaText = serializeSchema(data.schema.fields);
        saveError = null;
        saveDetails = [];
    });

    const references = $derived(
        data.contextReferences.filter((reference) => reference.context_id === schema.id),
    );
    const parsed = $derived(parseSchema(schemaText));
    const persistedText = $derived(serializeSchema(schema.fields));
    const dirty = $derived(schemaText !== persistedText);
    const payloadResults = $derived(validatePayload(payloadText, parsed.fields));
    const extraResults = $derived(payloadResults.filter((result) => result.status === "extra"));
    const referencedFieldCounts = $derived.by(() => {
        const counts = new Map<string, number>();
        for (const reference of references) {
            for (const field of reference.referenced_fields) {
                counts.set(field.path, (counts.get(field.path) ?? 0) + field.rule_count);
            }
        }
        return counts;
    });
    const missingReferencedFields = $derived(
        [...referencedFieldCounts.entries()].filter(
            ([path]) => !parsed.fields.some((field) => field.path === path),
        ),
    );
    const lineNumbers = $derived(Array.from({ length: Math.max(1, parsed.lineCount) }, (_, i) => i + 1));
    const canSave = $derived(
        dirty && !submitting && parsed.errors.length === 0 && missingReferencedFields.length === 0,
    );

    beforeNavigate((navigation) => {
        if (dirty && !window.confirm("Discard unsaved context schema changes?")) {
            navigation.cancel();
        }
    });

    async function saveFields() {
        if (!canSave) return;
        submitting = true;
        saveError = null;
        saveDetails = [];
        try {
            schema = await api.updateContext(schema.id, {
                name: schema.name,
                description: schema.description ?? "",
                fields: parsed.fields,
            });
            schemaText = serializeSchema(schema.fields);
            await invalidateAll();
        } catch (error) {
            saveError = error instanceof APIError ? error.message : "Failed to save context";
            saveDetails =
                error instanceof APIError
                    ? (error.details ?? []).map((detail) => detail.message)
                    : [];
        } finally {
            submitting = false;
        }
    }

    function revertFields() {
        schemaText = serializeSchema(schema.fields);
        saveError = null;
        saveDetails = [];
    }

    function addFields(fields: ContextField[]) {
        if (parsed.errors.length) return;
        const existing = new Set(parsed.fields.map((field) => field.path));
        schemaText = serializeSchema([
            ...parsed.fields,
            ...fields.filter((field) => !existing.has(field.path)),
        ]);
    }

    function formatPayload() {
        payloadText = formatJSON(payloadText);
    }

    async function formatPastedPayload() {
        await tick();
        formatPayload();
    }

    async function editDetails(name: string, description: string) {
        editing = true;
        editError = null;
        try {
            schema = await api.updateContext(schema.id, {
                name,
                description,
                fields: schema.fields,
            });
            editOpen = false;
            await invalidateAll();
        } catch (error) {
            editError = error instanceof APIError ? error.message : "Failed to update context";
        } finally {
            editing = false;
        }
    }

    async function remove() {
        if (references.length) return;
        deleting = true;
        deleteError = null;
        try {
            await api.deleteContext(schema.id);
            await goto("/contexts");
        } catch (error) {
            deleteError = error instanceof APIError ? error.message : "Failed to delete context";
        } finally {
            deleting = false;
        }
    }

    function flagHref(environmentID: string, flagKey: string) {
        return `/flags/${encodeURIComponent(flagKey)}?environment=${encodeURIComponent(environmentID)}`;
    }
</script>

{#snippet contextUsage()}
    <span
        class={cn(
            "rounded-sm border px-2.5 py-1 font-mono text-[0.6rem] uppercase tracking-[0.12em]",
            references.length
                ? "border-valid-border bg-valid-surface text-valid"
                : "border-border-strong text-muted-foreground",
        )}
    >
        {references.length
            ? `used by ${references.length} flags · ${references.reduce((total, reference) => total + reference.rule_count, 0)} rules`
            : "unused"}
    </span>
{/snippet}

{#snippet contextActions()}
    <Button
        variant="ghost"
        size="sm"
        disabled={dirty}
        title={dirty ? "Save or revert schema changes before editing details" : "Edit context details"}
        onclick={() => {
            editError = null;
            editOpen = true;
        }}
    >
        <Pencil class="h-3.5 w-3.5" /> edit details
    </Button>
    <Button
        variant={references.length ? "default" : "destructive"}
        size="sm"
        disabled={references.length > 0 || dirty}
        title={references.length
            ? "Remove flag references before deleting"
            : dirty
              ? "Save or revert schema changes before deleting"
              : "Delete context"}
        onclick={() => (deleteOpen = true)}
    >
        {#if references.length}<LockKeyhole class="h-3.5 w-3.5" />{:else}<Trash2 class="h-3.5 w-3.5" />{/if}
        delete
    </Button>
{/snippet}

<section class="motion-page min-h-full p-5 sm:p-7 lg:p-8">
    <PageHeader
        eyebrow="[ context ]"
        title={schema.name}
        description={schema.description || "No description"}
        titleClass="break-words font-mono"
        titleAfter={contextUsage}
        actions={contextActions}
    />

        {#if references.length}
            <div class="mt-5 flex flex-wrap items-center gap-2">
                <span class="mr-1 font-mono text-[0.6rem] uppercase tracking-[0.16em] text-muted-foreground">
                    used by
                </span>
                {#each references as reference (`${reference.environment_id}:${reference.flag_key}`)}
                    <a
                        href={flagHref(reference.environment_id, reference.flag_key)}
                        class="cursor-pointer rounded-md border border-border-control bg-surface-faint px-3 py-1.5 font-mono text-xs text-foreground transition-colors hover:border-border-hover hover:bg-surface-hover focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                    >
                        {reference.flag_key}
                        <span class="text-muted-foreground">· {reference.environment_key} · {reference.rule_count} rules</span>
                    </a>
                {/each}
            </div>
        {/if}

    <div class="mt-6 grid items-start gap-4 [grid-template-columns:repeat(auto-fit,minmax(min(100%,24rem),1fr))]">
        <Card class="overflow-hidden border-border-strong bg-card">
            <div class="flex min-h-11 flex-wrap items-center justify-between gap-2 border-b border-border-divider px-4 py-2">
                <span class="font-mono text-[0.62rem] uppercase tracking-[0.16em] text-muted-foreground">
                    schema · one field per line
                </span>
                <span
                    class={cn(
                        "font-mono text-[0.6rem] uppercase tracking-[0.1em]",
                        parsed.errors.length ? "text-destructive" : "text-valid",
                    )}
                    aria-live="polite"
                >
                    {parsed.errors.length
                        ? `${parsed.errors.length} invalid ${parsed.errors.length === 1 ? "line" : "lines"}`
                        : `${parsed.fields.length} fields parsed`}
                </span>
            </div>

            <div class="flex min-h-64 bg-surface-code">
                <div class="w-10 shrink-0 border-r border-border-divider py-3" aria-hidden="true">
                    {#each lineNumbers as number}
                        <span class="block pr-2 text-right font-mono text-xs leading-6 text-muted-foreground/50">{number}</span>
                    {/each}
                </div>
                <CodeTextarea
                    bind:value={schemaText}
                    syntax="schema"
                    aria-label="Context schema"
                    spellcheck={false}
                    class="min-h-64 flex-1 bg-transparent"
                    editorClass="text-[0.8rem]"
                    placeholder="user.id  string&#10;user.plan  string&#10;request.path  string"
                />
            </div>

            {#if parsed.errors.length || missingReferencedFields.length || saveError}
                <div class="space-y-2 border-t border-border-divider px-4 py-3" aria-live="polite">
                    {#each parsed.errors as error (`${error.line}:${error.message}`)}
                        <p class="flex gap-2 text-xs text-destructive">
                            <X class="mt-0.5 h-3.5 w-3.5 shrink-0" /> line {error.line}: {error.message}
                        </p>
                    {/each}
                    {#each missingReferencedFields as [path, count] (path)}
                        <p class="flex gap-2 text-xs text-cel-warning">
                            <AlertTriangle class="mt-0.5 h-3.5 w-3.5 shrink-0" />
                            {path} is referenced by {count} {count === 1 ? "rule" : "rules"}; restore it before saving.
                        </p>
                    {/each}
                    {#if saveError}
                        <p class="text-xs text-destructive">{saveError}</p>
                        {#each saveDetails as detail}<p class="text-xs text-destructive/80">{detail}</p>{/each}
                    {/if}
                </div>
            {/if}

            <div class="flex flex-wrap items-center gap-3 border-t border-border-divider px-4 py-2.5">
                <span class="mr-auto font-mono text-[0.58rem] text-muted-foreground">
                    types: string · int · double · bool · timestamp · list · map
                </span>
                {#if dirty}
                    <Button size="sm" variant="ghost" onclick={revertFields} disabled={submitting}>
                        <RotateCcw class="h-3 w-3" /> revert
                    </Button>
                {/if}
                <Button size="sm" variant="solid" onclick={saveFields} disabled={!canSave}>
                    <Save class="h-3 w-3" /> {submitting ? "saving…" : dirty ? "save changes" : "saved"}
                </Button>
            </div>
        </Card>

        <Card class="overflow-hidden border-border-strong bg-card">
            <div class="flex min-h-11 flex-wrap items-center justify-between gap-2 border-b border-border-divider px-4 py-2">
                <span class="font-mono text-[0.62rem] uppercase tracking-[0.16em] text-muted-foreground">
                    try it · paste a real payload
                </span>
                <Button
                    size="sm"
                    variant="ghost"
                    onclick={() => (payloadText = samplePayload(parsed.fields))}
                    disabled={!parsed.fields.length}
                >
                    <Zap class="h-3 w-3" /> insert sample
                </Button>
            </div>
            <CodeTextarea
                bind:value={payloadText}
                syntax="json"
                aria-label="Payload JSON"
                spellcheck={false}
                class="min-h-52 bg-surface-code"
                placeholder={'{\n  "user": { "id": "usr_4f2a", "plan": "pro" }\n}'}
                onblur={formatPayload}
                onpaste={formatPastedPayload}
            />

            <div class="border-t border-border-divider">
                {#if !payloadText.trim()}
                    <div class="flex items-center gap-3 px-4 py-6 text-xs text-muted-foreground">
                        <FileJson class="h-4 w-4" /> Paste JSON to compare it with this schema.
                    </div>
                {:else}
                    {#each payloadResults as result (`${result.status}:${result.path}`)}
                        <div class="flex items-center gap-3 border-b border-border-divider px-4 py-2.5 last:border-0">
                            {#if result.status === "valid"}
                                <Check class="h-3.5 w-3.5 shrink-0 text-valid" />
                            {:else if result.status === "extra"}
                                <Plus class="h-3.5 w-3.5 shrink-0 text-app-accent-muted" />
                            {:else if result.status === "missing"}
                                <X class="h-3.5 w-3.5 shrink-0 text-destructive" />
                            {:else}
                                <AlertTriangle class="h-3.5 w-3.5 shrink-0 text-cel-warning" />
                            {/if}
                            <span class="min-w-0 flex-1 truncate font-mono text-xs text-foreground">{result.path}</span>
                            <span class="text-right font-mono text-[0.62rem] text-muted-foreground">{result.message}</span>
                            {#if result.field}
                                <Button size="sm" variant="ghost" onclick={() => addFields([result.field!])} disabled={parsed.errors.length > 0}>
                                    add
                                </Button>
                            {/if}
                        </div>
                    {/each}
                    {#if extraResults.length > 1}
                        <div class="flex justify-end border-t border-border-divider px-4 py-2.5">
                            <Button
                                size="sm"
                                variant="ghost"
                                onclick={() => addFields(extraResults.flatMap((result) => (result.field ? [result.field] : [])))}
                                disabled={parsed.errors.length > 0}
                            >
                                <Plus class="h-3 w-3" /> add all {extraResults.length} fields
                            </Button>
                        </div>
                    {/if}
                {/if}
            </div>
        </Card>
    </div>
</section>

<ContextDetailsDialog
    bind:open={editOpen}
    initialName={schema.name}
    initialDescription={schema.description ?? ""}
    submitting={editing}
    error={editError}
    onsave={editDetails}
/>

<DestructiveDialog
    bind:open={deleteOpen}
    title="Delete context"
    description="This permanently removes the context schema."
    confirmationValue={schema.name}
    actionLabel="delete context"
    submitting={deleting}
    error={deleteError}
    onconfirm={remove}
/>
