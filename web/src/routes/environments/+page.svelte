<script lang="ts">
    import { invalidateAll } from "$app/navigation";
    import {
        Check,
        Circle,
        LockKeyhole,
        Pencil,
        Plus,
        Trash2,
        X,
    } from "lucide-svelte";
    import { api, APIError, type Environment } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import DestructiveDialog from "$lib/components/ui/destructive-dialog.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import { cn } from "$lib/utils";
    import type { PageProps } from "./$types";

    const DEFAULT_KEY = "production";
    const keyPattern = /^[a-z0-9][a-z0-9-]*$/;

    let { data }: PageProps = $props();

    const environments = $derived(
        [...data.environments].sort((a, b) => {
            if (a.key === DEFAULT_KEY) return 1;
            if (b.key === DEFAULT_KEY) return -1;
            return a.key.localeCompare(b.key);
        }),
    );
    const metrics = $derived(
        new Map(data.environmentMetrics.map((metric) => [metric.environment_id, metric])),
    );

    let createOpen = $state(false);
    let createKey = $state("");
    let createName = $state("");
    let createDescription = $state("");
    let creating = $state(false);
    let createError = $state<string | null>(null);

    let editingID = $state<string | null>(null);
    let editKey = $state("");
    let editName = $state("");
    let editDescription = $state("");
    let saving = $state(false);
    let editError = $state<string | null>(null);

    let deleteOpen = $state(false);
    let deleteTarget = $state<Environment | null>(null);
    let deleting = $state(false);
    let deleteError = $state<string | null>(null);

    const normalizedCreateKey = $derived(createKey.trim().toLowerCase());
    const createKeyTaken = $derived(
        environments.some((environment) => environment.key === normalizedCreateKey),
    );
    const createKeyValid = $derived(
        normalizedCreateKey.length > 0 && keyPattern.test(normalizedCreateKey) && !createKeyTaken,
    );

    function isDefault(environment: Environment) {
        return environment.key === DEFAULT_KEY;
    }

    function startEdit(environment: Environment) {
        editingID = environment.id;
        editKey = environment.key;
        editName = environment.name;
        editDescription = environment.description ?? "";
        editError = null;
        createOpen = false;
    }

    function cancelEdit() {
        editingID = null;
        editError = null;
    }

    async function saveEdit(environment: Environment) {
        const key = isDefault(environment) ? environment.key : editKey.trim().toLowerCase();
        if (!keyPattern.test(key) || saving) return;
        saving = true;
        editError = null;
        try {
            await api.updateEnvironment(environment.id, {
                key,
                name: editName.trim() || undefined,
                description: editDescription.trim() || undefined,
            });
            editingID = null;
            await invalidateAll();
        } catch (error) {
            editError = error instanceof APIError ? error.message : "Failed to save environment";
        } finally {
            saving = false;
        }
    }

    function openCreate() {
        editingID = null;
        createOpen = true;
        createError = null;
    }

    function closeCreate() {
        createOpen = false;
        createKey = "";
        createName = "";
        createDescription = "";
        createError = null;
    }

    async function createEnvironment() {
        if (!createKeyValid || creating) return;
        creating = true;
        createError = null;
        try {
            await api.createEnvironment({
                key: normalizedCreateKey,
                name: createName.trim() || undefined,
                description: createDescription.trim() || undefined,
            });
            closeCreate();
            await invalidateAll();
        } catch (error) {
            createError = error instanceof APIError ? error.message : "Failed to create environment";
        } finally {
            creating = false;
        }
    }

    function requestDelete(environment: Environment) {
        deleteTarget = environment;
        deleteError = null;
        deleteOpen = true;
    }

    async function confirmDelete() {
        if (!deleteTarget) return;
        deleting = true;
        deleteError = null;
        try {
            await api.deleteEnvironment(deleteTarget.id);
            deleteOpen = false;
            deleteTarget = null;
            await invalidateAll();
        } catch (error) {
            deleteError = error instanceof APIError ? error.message : "Failed to delete environment";
        } finally {
            deleting = false;
        }
    }

    function formatUpdated(value: string) {
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return "unknown";
        return new Intl.DateTimeFormat(undefined, {
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "2-digit",
        }).format(date);
    }

    function indicatorClass(environment: Environment, index: number) {
        if (isDefault(environment)) return "text-valid";
        return index % 2 === 0 ? "text-cel-warning" : "text-app-accent-muted";
    }
</script>

<section class="motion-page space-y-7">
    <header>
        <p class="font-mono text-[0.65rem] uppercase tracking-[0.2em] text-muted-foreground">
            [ environments · {environments.length} ]
        </p>
        <div class="mt-2 flex flex-wrap items-baseline gap-x-4 gap-y-2">
            <h1 class="text-3xl font-semibold tracking-tight">Environments</h1>
            <p class="font-mono text-[0.65rem] uppercase tracking-[0.15em] text-muted-foreground">
                workspaces · isolated flag sets
            </p>
        </div>
    </header>

    <Card class="motion-panel overflow-hidden border-border-strong bg-card">
        <div
            class="hidden min-h-10 grid-cols-[1.25rem_12rem_minmax(10rem,1fr)_4rem_5.5rem_8rem_9rem] items-center gap-3 border-b border-border-divider bg-surface-faint px-4 font-mono text-[0.58rem] uppercase tracking-[0.16em] text-muted-foreground lg:grid"
        >
            <span></span>
            <span>key</span>
            <span>name · description</span>
            <span class="text-right">flags</span>
            <span class="text-right">overrides</span>
            <span>updated</span>
            <span></span>
        </div>

        <div class="motion-list">
            {#each environments as environment, index (environment.id)}
                {@const metric = metrics.get(environment.id)}
                {#if editingID === environment.id}
                    <form
                        class="space-y-3 border-b border-border-divider bg-surface-subtle p-4 shadow-[inset_0_0_0_1px_var(--color-border-strong)]"
                        onsubmit={(event) => {
                            event.preventDefault();
                            saveEdit(environment);
                        }}
                    >
                        <p class="font-mono text-[0.62rem] uppercase tracking-[0.17em] text-muted-foreground">
                            [ edit · {environment.key} ]
                        </p>
                        <div class="grid gap-2 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)]">
                            <Input
                                bind:value={editKey}
                                aria-label="Environment key"
                                placeholder="key"
                                class="font-mono"
                                disabled={isDefault(environment)}
                            />
                            <Input bind:value={editName} aria-label="Environment name" placeholder="name (optional)" />
                        </div>
                        <Input
                            bind:value={editDescription}
                            aria-label="Environment description"
                            placeholder="description (optional)"
                        />
                        {#if isDefault(environment)}
                            <p class="font-mono text-[0.6rem] uppercase tracking-[0.1em] text-muted-foreground">
                                the default environment key is locked
                            </p>
                        {:else if editKey.trim() && !keyPattern.test(editKey.trim().toLowerCase())}
                            <p class="text-xs text-destructive">
                                Start with a letter or number, then use lowercase letters, numbers, and hyphens.
                            </p>
                        {/if}
                        {#if editError}<p class="text-sm text-destructive" aria-live="polite">{editError}</p>{/if}
                        <div class="flex gap-2">
                            <Button
                                type="submit"
                                size="sm"
                                variant="solid"
                                disabled={saving || !keyPattern.test(editKey.trim().toLowerCase())}
                            >
                                <Check class="h-3.5 w-3.5" /> {saving ? "saving…" : "save"}
                            </Button>
                            <Button type="button" size="sm" variant="ghost" onclick={cancelEdit} disabled={saving}>
                                <X class="h-3.5 w-3.5" /> cancel
                            </Button>
                        </div>
                    </form>
                {:else}
                    <div
                        class="grid gap-3 border-b border-border-divider px-4 py-4 transition-colors last:border-0 hover:bg-surface-hover lg:grid-cols-[1.25rem_12rem_minmax(10rem,1fr)_4rem_5.5rem_8rem_9rem] lg:items-center lg:py-3"
                    >
                        <Circle
                            class={cn("hidden h-2.5 w-2.5 fill-current lg:block", indicatorClass(environment, index))}
                            aria-hidden="true"
                        />
                        <div class="flex min-w-0 flex-nowrap items-center gap-2">
                            <Circle
                                class={cn("h-2.5 w-2.5 shrink-0 fill-current lg:hidden", indicatorClass(environment, index))}
                                aria-hidden="true"
                            />
                            <span class="truncate font-mono text-sm font-medium">{environment.key}</span>
                            {#if isDefault(environment)}
                                <span class="ios-corners-xs inline-flex shrink-0 items-center gap-1 whitespace-nowrap border border-border-control px-2 py-0.5 font-mono text-[0.55rem] uppercase tracking-[0.1em] text-muted-foreground">
                                    <LockKeyhole class="h-2.5 w-2.5" /> default · locked
                                </span>
                            {/if}
                        </div>
                        <div class="min-w-0">
                            <p class="truncate text-sm text-foreground-soft">{environment.name || "-"}</p>
                            <p class="mt-0.5 truncate text-xs text-muted-foreground">{environment.description || "-"}</p>
                        </div>
                        <div class="ios-corners-sm grid grid-cols-3 gap-3 bg-surface-faint p-3 lg:contents lg:bg-transparent lg:p-0">
                            <div class="lg:text-right">
                                <p class="font-mono text-[0.55rem] uppercase tracking-[0.12em] text-muted-foreground lg:hidden">flags</p>
                                <a
                                    href="/?environment={encodeURIComponent(environment.id)}"
                                    class="mt-1 block font-mono text-xs text-foreground transition-colors hover:text-app-accent lg:mt-0"
                                >
                                    {metric?.flag_count ?? 0}
                                </a>
                            </div>
                            <div class="lg:text-right">
                                <p class="font-mono text-[0.55rem] uppercase tracking-[0.12em] text-muted-foreground lg:hidden">overrides</p>
                                <p class={cn("mt-1 font-mono text-xs lg:mt-0", metric?.override_count ? "text-cel-warning" : "text-muted-foreground")}>
                                    {metric?.override_count ? metric.override_count : "-"}
                                </p>
                            </div>
                            <div>
                                <p class="font-mono text-[0.55rem] uppercase tracking-[0.12em] text-muted-foreground lg:hidden">updated</p>
                                <p class="mt-1 whitespace-nowrap font-mono text-[0.62rem] text-muted-foreground lg:mt-0">
                                    {formatUpdated(environment.updated_at)}
                                </p>
                            </div>
                        </div>
                        <div class="flex justify-end gap-2">
                            <Button size="sm" variant="ghost" onclick={() => startEdit(environment)}>
                                <Pencil class="h-3.5 w-3.5" /> edit
                            </Button>
                            {#if !isDefault(environment)}
                                <Button
                                    size="icon"
                                    variant="destructive"
                                    class="h-7 w-7"
                                    aria-label="Delete {environment.key} environment"
                                    onclick={() => requestDelete(environment)}
                                >
                                    <Trash2 class="h-3.5 w-3.5" />
                                </Button>
                            {/if}
                        </div>
                    </div>
                {/if}
            {/each}
        </div>

        {#if createOpen}
            <form
                class="space-y-3 bg-surface-subtle p-4 shadow-[inset_0_0_0_1px_var(--color-border-strong)]"
                onsubmit={(event) => {
                    event.preventDefault();
                    createEnvironment();
                }}
            >
                <div class="flex flex-wrap items-center justify-between gap-2">
                    <p class="font-mono text-[0.62rem] uppercase tracking-[0.17em] text-muted-foreground">
                        [ new environment ]
                    </p>
                    <p class="font-mono text-[0.55rem] uppercase tracking-[0.1em] text-muted-foreground">
                        key · lowercase letters, numbers, and hyphens
                    </p>
                </div>
                <div class="grid gap-2 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)]">
                    <Input
                        bind:value={createKey}
                        aria-label="New environment key"
                        placeholder="key · e.g. staging"
                        class={cn("font-mono", createKey.trim() && !createKeyValid && "border-destructive")}
                    />
                    <Input bind:value={createName} aria-label="New environment name" placeholder="name (optional)" />
                </div>
                <Input
                    bind:value={createDescription}
                    aria-label="New environment description"
                    placeholder="description (optional)"
                />
                {#if createKey.trim() && !keyPattern.test(normalizedCreateKey)}
                    <p class="text-xs text-destructive">
                        Start with a letter or number, then use lowercase letters, numbers, and hyphens.
                    </p>
                {:else if createKeyTaken}
                    <p class="text-xs text-destructive">That environment key already exists.</p>
                {/if}
                {#if createError}<p class="text-sm text-destructive" aria-live="polite">{createError}</p>{/if}
                <div class="flex gap-2">
                    <Button type="submit" size="sm" variant="solid" disabled={!createKeyValid || creating}>
                        <Plus class="h-3.5 w-3.5" /> {creating ? "creating…" : "create"}
                    </Button>
                    <Button type="button" size="sm" variant="ghost" onclick={closeCreate} disabled={creating}>
                        <X class="h-3.5 w-3.5" /> cancel
                    </Button>
                </div>
            </form>
        {:else}
            <button
                type="button"
                class="flex w-full cursor-pointer items-center gap-3 border-t border-border-divider px-4 py-4 text-left font-mono text-[0.62rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:bg-surface-hover hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-ring"
                onclick={openCreate}
            >
                <span class="h-2.5 w-2.5 rounded-full border border-dashed border-border-strong" aria-hidden="true"></span>
                <Plus class="h-3.5 w-3.5" /> new environment
            </button>
        {/if}
    </Card>

    <p class="font-mono text-[0.58rem] uppercase tracking-[0.12em] text-muted-foreground">
        overrides · flags whose targeting differs from the default environment
    </p>
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
