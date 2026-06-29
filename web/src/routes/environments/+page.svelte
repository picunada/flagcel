<script lang="ts">
    import { Check, Layers3, Pencil, Plus, Trash2, X } from "lucide-svelte";
    import { api, APIError, type Environment } from "$lib/api";
    import { invalidateAll } from "$app/navigation";
    import Badge from "$lib/components/ui/badge.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import DestructiveDialog from "$lib/components/ui/destructive-dialog.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import SectionHeader from "$lib/components/ui/section-header.svelte";
    import type { PageProps } from "./$types";

    const DEFAULT_KEY = "production";

    let { data }: PageProps = $props();
    const environments = $derived<Environment[]>(data.environments);

    // create
    let key = $state("");
    let name = $state("");
    let description = $state("");
    let creating = $state(false);
    let createError = $state<string | null>(null);

    // edit
    let editingId = $state<string | null>(null);
    let editKey = $state("");
    let editName = $state("");
    let editDescription = $state("");
    let saving = $state(false);
    let editError = $state<string | null>(null);

    // delete
    let deleteOpen = $state(false);
    let deleteTarget = $state<Environment | null>(null);
    let deleting = $state(false);
    let deleteError = $state<string | null>(null);

    const isDefault = (env: Environment) => env.key === DEFAULT_KEY;

    async function createEnvironment() {
        if (!key.trim()) return;
        creating = true;
        createError = null;
        try {
            await api.createEnvironment({
                key: key.trim(),
                name: name.trim() || undefined,
                description: description.trim() || undefined,
            });
            key = "";
            name = "";
            description = "";
            await invalidateAll();
        } catch (e) {
            createError = e instanceof APIError ? e.message : "Failed to create environment";
        } finally {
            creating = false;
        }
    }

    function startEdit(env: Environment) {
        editingId = env.id;
        editKey = env.key;
        editName = env.name;
        editDescription = env.description ?? "";
        editError = null;
    }

    function cancelEdit() {
        editingId = null;
        editError = null;
    }

    async function saveEdit(env: Environment) {
        saving = true;
        editError = null;
        try {
            await api.updateEnvironment(env.id, {
                key: isDefault(env) ? env.key : editKey.trim(),
                name: editName.trim() || undefined,
                description: editDescription.trim() || undefined,
            });
            editingId = null;
            await invalidateAll();
        } catch (e) {
            editError = e instanceof APIError ? e.message : "Failed to save environment";
        } finally {
            saving = false;
        }
    }

    function requestDelete(env: Environment) {
        deleteTarget = env;
        deleteError = null;
        deleteOpen = true;
    }

    async function confirmDelete() {
        const env = deleteTarget;
        if (!env) return;
        deleting = true;
        deleteError = null;
        try {
            await api.deleteEnvironment(env.id);
            deleteOpen = false;
            deleteTarget = null;
            await invalidateAll();
        } catch (e) {
            deleteError = e instanceof APIError ? e.message : "Failed to delete environment";
        } finally {
            deleting = false;
        }
    }

    function formatDate(value?: string) {
        if (!value) return "never";
        return new Date(value).toLocaleString();
    }
</script>

<section class="space-y-10">
    <header class="space-y-3">
        <p class="text-xs uppercase tracking-[0.18em] text-muted-foreground">
            workspaces · isolated flag sets
        </p>
        <h1 class="text-3xl font-normal leading-tight sm:text-4xl">
            Environments
        </h1>
        <p class="max-w-xl text-sm text-foreground-soft sm:text-base">
            Each environment holds its own flags, rules, and API keys. Switch the active
            environment from the selector in the header.
        </p>
    </header>

    <Card class="motion-panel space-y-4 p-5">
        <SectionHeader>new environment</SectionHeader>
        <div class="grid gap-2 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)_auto]">
            <Input
                bind:value={key}
                placeholder="key · e.g. staging"
                class="font-mono"
                onkeydown={(e) => e.key === "Enter" && createEnvironment()}
            />
            <Input
                bind:value={name}
                placeholder="name (optional)"
                onkeydown={(e) => e.key === "Enter" && createEnvironment()}
            />
            <Button onclick={createEnvironment} disabled={creating || !key.trim()}>
                <Plus class="h-3.5 w-3.5" /> create
            </Button>
        </div>
        <Input bind:value={description} placeholder="description (optional)" />
        <p class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
            key · lowercase letters, numbers, and hyphens
        </p>
        {#if createError}
            <p class="text-sm text-destructive">{createError}</p>
        {/if}
    </Card>

    <div class="space-y-3">
        <SectionHeader>environments · {environments.length}</SectionHeader>
        {#if environments.length === 0}
            <Card class="motion-panel p-10 text-center">
                <Layers3 class="mx-auto h-5 w-5 text-muted-foreground" />
                <p class="mt-4 text-xs uppercase tracking-[0.14em] text-muted-foreground">
                    [ no environments yet ]
                </p>
            </Card>
        {:else}
            <div class="motion-list space-y-3">
                {#each environments as env (env.id)}
                    <Card class="p-5">
                        {#if editingId === env.id}
                            <div class="space-y-3">
                                <div class="grid gap-2 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)]">
                                    <Input
                                        bind:value={editKey}
                                        placeholder="key"
                                        class="font-mono"
                                        disabled={isDefault(env)}
                                    />
                                    <Input bind:value={editName} placeholder="name" />
                                </div>
                                <Input bind:value={editDescription} placeholder="description (optional)" />
                                {#if isDefault(env)}
                                    <p class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
                                        the default environment key cannot be changed
                                    </p>
                                {/if}
                                {#if editError}
                                    <p class="text-sm text-destructive">{editError}</p>
                                {/if}
                                <div class="flex items-center gap-2">
                                    <Button size="sm" onclick={() => saveEdit(env)} disabled={saving}>
                                        <Check class="h-3.5 w-3.5" /> save
                                    </Button>
                                    <Button size="sm" variant="ghost" onclick={cancelEdit} disabled={saving}>
                                        <X class="h-3.5 w-3.5" /> cancel
                                    </Button>
                                </div>
                            </div>
                        {:else}
                            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                                <div class="min-w-0 space-y-2">
                                    <div class="flex items-center gap-2">
                                        <p class="truncate font-mono text-base">{env.key}</p>
                                        {#if isDefault(env)}
                                            <Badge variant="muted">default · locked</Badge>
                                        {/if}
                                    </div>
                                    <p class="text-sm text-foreground-soft">{env.name}</p>
                                    {#if env.description}
                                        <p class="text-sm text-muted-foreground">{env.description}</p>
                                    {/if}
                                    <p class="text-xs text-muted-foreground">
                                        created {formatDate(env.created_at)} · updated {formatDate(env.updated_at)}
                                    </p>
                                </div>
                                <div class="flex shrink-0 items-center gap-2">
                                    <Button size="sm" variant="ghost" onclick={() => startEdit(env)}>
                                        <Pencil class="h-3.5 w-3.5" /> edit
                                    </Button>
                                    {#if !isDefault(env)}
                                        <Button size="sm" variant="destructive" onclick={() => requestDelete(env)}>
                                            <Trash2 class="h-3.5 w-3.5" /> delete
                                        </Button>
                                    {/if}
                                </div>
                            </div>
                        {/if}
                    </Card>
                {/each}
            </div>
        {/if}
    </div>
</section>

<DestructiveDialog
    bind:open={deleteOpen}
    title="Delete environment"
    description="This permanently removes the environment. It must have no flags, rules, or API keys."
    details={deleteTarget ? `${deleteTarget.key}\n${deleteTarget.name}` : null}
    confirmationValue={deleteTarget?.key ?? null}
    actionLabel="delete environment"
    submitting={deleting}
    error={deleteError}
    onconfirm={confirmDelete}
    oncancel={() => {
        deleteTarget = null;
        deleteError = null;
    }}
/>
