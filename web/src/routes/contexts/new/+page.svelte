<script lang="ts">
    import { tick } from "svelte";
    import { goto } from "$app/navigation";
    import { AlertTriangle, Check, FileJson, Plus, X } from "lucide-svelte";
    import { api, APIError } from "$lib/api";
    import { formatJSON } from "$lib/code-highlighting";
    import { inferPayload } from "$lib/context-schema";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import CodeTextarea from "$lib/components/ui/code-textarea.svelte";
    import FieldLabel from "$lib/components/ui/field-label.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";
    import { cn } from "$lib/utils";

    let name = $state("");
    let description = $state("");
    let payloadText = $state("");
    let submitting = $state(false);
    let error = $state<string | null>(null);

    const inference = $derived(inferPayload(payloadText));
    const canCreate = $derived(name.trim().length > 0 && !inference.error && !submitting);

    async function create() {
        if (!canCreate) return;
        submitting = true;
        error = null;
        try {
            const context = await api.createContext({
                name: name.trim(),
                description: description.trim(),
                fields: inference.fields,
            });
            await goto(`/contexts/${encodeURIComponent(context.id)}`);
        } catch (cause) {
            error = cause instanceof APIError ? cause.message : "Failed to create context";
        } finally {
            submitting = false;
        }
    }

    function formatPayload() {
        payloadText = formatJSON(payloadText);
    }

    async function formatPastedPayload() {
        await tick();
        formatPayload();
    }
</script>

<section class="motion-page min-h-full p-5 sm:p-7 lg:p-8">
    <PageHeader
        eyebrow="[ new context ]"
        title="Define a context"
        description="Paste a representative payload to infer its field schema, or create an empty context and add fields manually."
    />

    <div class="mt-7 grid items-start gap-4 [grid-template-columns:repeat(auto-fit,minmax(min(100%,24rem),1fr))]">
        <div class="space-y-4">
            <Card class="border-border-strong bg-card p-5">
                <div class="grid gap-4 sm:grid-cols-2">
                    <label class="block space-y-2">
                        <FieldLabel>name</FieldLabel>
                        <Input bind:value={name} required placeholder="web-user" spellcheck={false} />
                    </label>
                    <label class="block space-y-2">
                        <FieldLabel>description · optional</FieldLabel>
                        <Input bind:value={description} placeholder="Attributes sent by web clients" />
                    </label>
                </div>
            </Card>

            <Card class="overflow-hidden border-border-strong bg-card">
                <div class="flex min-h-11 flex-wrap items-center justify-between gap-2 border-b border-border-divider px-4 py-2">
                    <span class="font-mono text-[0.62rem] uppercase tracking-[0.16em] text-muted-foreground">
                        paste payload JSON
                    </span>
                    <span
                        class={cn(
                            "font-mono text-[0.6rem] uppercase tracking-[0.1em]",
                            inference.error
                                ? "text-destructive"
                                : inference.fields.length
                                  ? "text-valid"
                                  : "text-muted-foreground",
                        )}
                        aria-live="polite"
                    >
                        {inference.error
                            ? "invalid JSON"
                            : inference.fields.length
                              ? `${inference.fields.length} fields inferred`
                              : "optional"}
                    </span>
                </div>
                <CodeTextarea
                    bind:value={payloadText}
                    syntax="json"
                    spellcheck={false}
                    aria-label="Payload JSON for schema inference"
                    class="min-h-72 bg-surface-code"
                    placeholder={'{\n  "user": { "id": "usr_991", "plan": "pro", "beta": true },\n  "device": { "os": "ios", "version": 17 }\n}'}
                    onblur={formatPayload}
                    onpaste={formatPastedPayload}
                />
                {#if inference.error || inference.warnings.length}
                    <div class="space-y-2 border-t border-border-divider px-4 py-3" aria-live="polite">
                        {#if inference.error}
                            <p class="flex gap-2 text-xs text-destructive">
                                <X class="mt-0.5 h-3.5 w-3.5 shrink-0" /> {inference.error}
                            </p>
                        {/if}
                        {#each inference.warnings as warning (warning)}
                            <p class="flex gap-2 text-xs text-cel-warning">
                                <AlertTriangle class="mt-0.5 h-3.5 w-3.5 shrink-0" /> {warning}
                            </p>
                        {/each}
                    </div>
                {/if}
            </Card>
        </div>

        <Card class="overflow-hidden border-border-strong bg-card">
            <div class="border-b border-border-divider px-4 py-3 font-mono text-[0.62rem] uppercase tracking-[0.16em] text-muted-foreground">
                inferred fields
            </div>
            <div class="min-h-72">
                {#if inference.fields.length === 0}
                    <div class="flex gap-3 px-5 py-7 text-sm leading-6 text-muted-foreground">
                        <FileJson class="mt-1 h-4 w-4 shrink-0" />
                        <p>
                            Fields appear here as you paste. Nested objects become dotted paths such as
                            <span class="font-mono text-foreground-softer">user.plan</span>.
                        </p>
                    </div>
                {:else}
                    {#each inference.fields as field (field.path)}
                        <div class="motion-panel flex items-center gap-4 border-b border-border-divider px-4 py-3 last:border-0">
                            <Check class="h-3.5 w-3.5 shrink-0 text-valid" />
                            <span class="min-w-0 flex-1 truncate font-mono text-sm">{field.path}</span>
                            <span class="rounded-sm border border-border-control bg-surface-faint px-2 py-0.5 font-mono text-[0.62rem] uppercase tracking-[0.1em] text-app-accent-muted">
                                {field.type}
                            </span>
                        </div>
                    {/each}
                {/if}
            </div>
            <div class="border-t border-border-divider p-4">
                {#if error}<p class="mb-3 text-sm text-destructive" aria-live="polite">{error}</p>{/if}
                <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                    <Button variant="ghost" disabled={submitting} onclick={() => goto("/contexts")}>cancel</Button>
                    <Button variant="solid" disabled={!canCreate} onclick={create}>
                        <Plus class="h-3.5 w-3.5" /> {submitting ? "creating…" : "create context"}
                    </Button>
                </div>
                <p class="mt-3 text-right font-mono text-[0.6rem] text-muted-foreground">
                    {payloadText.trim() ? "schema will use the inferred fields above" : "creates an empty schema"}
                </p>
            </div>
        </Card>
    </div>
</section>
