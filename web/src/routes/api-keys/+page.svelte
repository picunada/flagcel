<script lang="ts">
    import { untrack } from "svelte";
    import { invalidateAll } from "$app/navigation";
    import {
        Check,
        Circle,
        Copy,
        KeyRound,
        Plus,
        ShieldAlert,
        X,
    } from "lucide-svelte";
    import { api, APIError, type APIKey, type CreateAPIKeyResponse } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import CodeBlock from "$lib/components/ui/code-block.svelte";
    import DestructiveDialog from "$lib/components/ui/destructive-dialog.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import { cn } from "$lib/utils";
    import type { PageProps } from "./$types";

    type StatusFilter = "active" | "revoked" | "all";

    let { data }: PageProps = $props();

    const environments = $derived(data.environments);
    const environmentByID = $derived(
        new Map(environments.map((environment) => [environment.id, environment])),
    );
    const keys = $derived<APIKey[]>(data.keys);
    const activeCount = $derived(keys.filter((key) => !key.revoked_at).length);
    const revokedCount = $derived(keys.filter((key) => Boolean(key.revoked_at)).length);
    const filteredKeys = $derived(
        keys.filter((key) => {
            if (filter === "active") return !key.revoked_at;
            if (filter === "revoked") return Boolean(key.revoked_at);
            return true;
        }),
    );

    let filter = $state<StatusFilter>("active");
    let createOpen = $state(false);
    let createName = $state("");
    let createEnvironmentID = $state(untrack(() => data.selectedEnvironment.id));
    let creating = $state(false);
    let createError = $state<string | null>(null);
    let created = $state<CreateAPIKeyResponse | null>(null);
    let revealCopied = $state(false);
    let copiedPrefixID = $state<string | null>(null);
    let revokeOpen = $state(false);
    let revokeTarget = $state<APIKey | null>(null);
    let revoking = $state(false);
    let revokeError = $state<string | null>(null);

    function openCreate() {
        createOpen = true;
        createError = null;
        createEnvironmentID = data.selectedEnvironment.id;
    }

    function closeCreate() {
        createOpen = false;
        createName = "";
        createError = null;
    }

    async function createKey() {
        if (!createName.trim() || !createEnvironmentID || creating) return;
        creating = true;
        createError = null;
        created = null;
        try {
            created = await api.createAPIKey(createName.trim(), createEnvironmentID);
            revealCopied = false;
            closeCreate();
            if (filter === "revoked") filter = "active";
            await invalidateAll();
        } catch (error) {
            createError = error instanceof APIError ? error.message : "Failed to create API key";
        } finally {
            creating = false;
        }
    }

    function requestRevoke(key: APIKey) {
        revokeTarget = key;
        revokeError = null;
        revokeOpen = true;
    }

    async function revokeKey() {
        if (!revokeTarget) return;
        revoking = true;
        revokeError = null;
        try {
            await api.revokeAPIKey(revokeTarget.id);
            revokeOpen = false;
            revokeTarget = null;
            await invalidateAll();
        } catch (error) {
            revokeError = error instanceof APIError ? error.message : "Failed to revoke API key";
        } finally {
            revoking = false;
        }
    }

    async function copyText(value: string) {
        try {
            await navigator.clipboard.writeText(value);
            return true;
        } catch {
            return false;
        }
    }

    async function copyReveal() {
        if (created?.token && (await copyText(created.token))) {
            revealCopied = true;
        }
    }

    async function copyPrefix(key: APIKey) {
        if (await copyText(key.prefix)) {
            copiedPrefixID = key.id;
            setTimeout(() => {
                if (copiedPrefixID === key.id) copiedPrefixID = null;
            }, 1200);
        }
    }

    function formatDate(value?: string) {
        if (!value) return "never";
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return "unknown";
        return new Intl.DateTimeFormat(undefined, {
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "2-digit",
        }).format(date);
    }

    function environmentIndicator(key: string) {
        if (key === "production") return "text-valid";
        if (key === "staging") return "text-app-accent-muted";
        return "text-cel-warning";
    }

    function filterCount(status: StatusFilter) {
        if (status === "active") return activeCount;
        if (status === "revoked") return revokedCount;
        return keys.length;
    }
</script>

<section class="motion-page space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
        <div>
            <p class="font-mono text-[0.65rem] uppercase tracking-[0.2em] text-muted-foreground">
                [ api keys · {keys.length} ]
            </p>
            <div class="mt-2 flex flex-wrap items-baseline gap-x-4 gap-y-2">
                <h1 class="text-3xl font-semibold tracking-tight">API keys</h1>
                <p class="font-mono text-[0.65rem] uppercase tracking-[0.15em] text-muted-foreground">
                    eval access · scoped per environment
                </p>
            </div>
        </div>

        <div class="ios-corners-sm flex flex-wrap gap-1 border border-border-control bg-surface-faint p-1" aria-label="API key status filter">
            {#each ["active", "revoked", "all"] as status (status)}
                <button
                    type="button"
                    class={cn(
                        "ios-corners-xs cursor-pointer px-3 py-2 font-mono text-[0.6rem] uppercase tracking-[0.12em] transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
                        filter === status
                            ? "bg-foreground font-semibold text-primary-foreground"
                            : "text-muted-foreground hover:bg-surface-hover hover:text-foreground",
                    )}
                    aria-pressed={filter === status}
                    onclick={() => (filter = status as StatusFilter)}
                >
                    {status} · {filterCount(status as StatusFilter)}
                </button>
            {/each}
        </div>
    </header>

    {#if created}
        <Card class="motion-panel border-valid-border bg-valid-surface p-4">
            <div class="flex items-start justify-between gap-4">
                <div class="min-w-0">
                    <p class="flex items-center gap-2 font-mono text-[0.62rem] uppercase tracking-[0.16em] text-valid">
                        <Check class="h-3.5 w-3.5" /> key created · {created.name}
                    </p>
                    <p class="mt-2 text-xs text-muted-foreground">
                        This is the only time the full key is shown. Store it securely.
                    </p>
                </div>
                <Button
                    size="icon"
                    variant="ghost"
                    class="h-7 w-7"
                    aria-label="Dismiss created key"
                    onclick={() => (created = null)}
                >
                    <X class="h-3.5 w-3.5" />
                </Button>
            </div>
            <div class="mt-3 grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
                <CodeBlock value={created.token} class="border border-border-control" />
                <Button variant="solid" onclick={copyReveal}>
                    {#if revealCopied}<Check class="h-3.5 w-3.5" /> copied{:else}<Copy class="h-3.5 w-3.5" /> copy{/if}
                </Button>
            </div>
        </Card>
    {/if}

    <Card class="motion-panel overflow-hidden border-border-strong bg-card">
        <div
            class="hidden min-h-10 grid-cols-[1.25rem_minmax(12rem,1fr)_10rem_7rem_9rem_8rem_7rem] items-center gap-3 border-b border-border-divider bg-surface-faint px-4 font-mono text-[0.58rem] uppercase tracking-[0.16em] text-muted-foreground xl:grid"
        >
            <span></span>
            <span>name</span>
            <span>key</span>
            <span>env</span>
            <span>last used</span>
            <span>created</span>
            <span></span>
        </div>

        <div class="motion-list">
            {#each filteredKeys as key (key.id)}
                {@const environment = environmentByID.get(key.environment_id)}
                {@const revoked = Boolean(key.revoked_at)}
                <div
                    class={cn(
                        "grid gap-3 border-b border-border-divider px-4 py-4 transition-colors last:border-0 hover:bg-surface-hover xl:grid-cols-[1.25rem_minmax(12rem,1fr)_10rem_7rem_9rem_8rem_7rem] xl:items-center xl:py-3",
                        revoked && "opacity-55",
                    )}
                >
                    <Circle
                        class={cn(
                            "hidden h-2.5 w-2.5 fill-current xl:block",
                            revoked ? "text-muted-foreground" : "text-valid",
                        )}
                        aria-label={revoked ? "Revoked" : "Active"}
                    />
                    <div class="flex min-w-0 flex-wrap items-center gap-2">
                        <Circle
                            class={cn(
                                "h-2.5 w-2.5 shrink-0 fill-current xl:hidden",
                                revoked ? "text-muted-foreground" : "text-valid",
                            )}
                            aria-hidden="true"
                        />
                        <span class="truncate font-mono text-sm font-medium">{key.name}</span>
                        {#if revoked}
                            <span class="ios-corners-xs shrink-0 whitespace-nowrap border border-destructive-border px-2 py-0.5 font-mono text-[0.55rem] uppercase tracking-[0.1em] text-destructive">
                                revoked
                            </span>
                        {:else if !key.last_used_at}
                            <span class="ios-corners-xs shrink-0 whitespace-nowrap border border-cel-warning/30 px-2 py-0.5 font-mono text-[0.55rem] uppercase tracking-[0.1em] text-cel-warning">
                                never used
                            </span>
                        {/if}
                    </div>

                    <div class="flex items-center gap-2">
                        <code class="min-w-0 truncate font-mono text-xs text-muted-foreground">{key.prefix}</code>
                        <Button
                            size="icon"
                            variant="ghost"
                            class="h-7 w-7 shrink-0"
                            aria-label="Copy prefix for {key.name}"
                            onclick={() => copyPrefix(key)}
                        >
                            {#if copiedPrefixID === key.id}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                        </Button>
                    </div>

                    <div class="flex items-center gap-2 text-xs text-foreground-soft">
                        <Circle
                            class={cn("h-2 w-2 fill-current", environmentIndicator(environment?.key ?? ""))}
                            aria-hidden="true"
                        />
                        <span class="truncate">{environment?.key ?? "missing"}</span>
                    </div>

                    <div class="ios-corners-sm grid grid-cols-2 gap-3 bg-surface-faint p-3 xl:contents xl:bg-transparent xl:p-0">
                        <div>
                            <p class="font-mono text-[0.55rem] uppercase tracking-[0.12em] text-muted-foreground xl:hidden">last used</p>
                            <p class="mt-1 whitespace-nowrap font-mono text-[0.62rem] text-muted-foreground xl:mt-0">
                                {formatDate(key.last_used_at)}
                            </p>
                        </div>
                        <div>
                            <p class="font-mono text-[0.55rem] uppercase tracking-[0.12em] text-muted-foreground xl:hidden">created</p>
                            <p class="mt-1 whitespace-nowrap font-mono text-[0.62rem] text-muted-foreground xl:mt-0">
                                {formatDate(key.created_at)}
                            </p>
                        </div>
                    </div>

                    <div class="flex justify-end">
                        {#if revoked}
                            <span class="font-mono text-[0.58rem] uppercase tracking-[0.1em] text-muted-foreground">
                                {formatDate(key.revoked_at)}
                            </span>
                        {:else}
                            <Button size="sm" variant="destructive" onclick={() => requestRevoke(key)}>
                                revoke
                            </Button>
                        {/if}
                    </div>
                </div>
            {:else}
                <div class="flex items-center gap-3 px-5 py-8 text-sm text-muted-foreground">
                    <KeyRound class="h-4 w-4" /> No {filter === "all" ? "" : `${filter} `}API keys.
                </div>
            {/each}
        </div>

        {#if createOpen}
            <form
                class="space-y-3 bg-surface-subtle p-4 shadow-[inset_0_0_0_1px_var(--color-border-strong)]"
                onsubmit={(event) => {
                    event.preventDefault();
                    createKey();
                }}
            >
                <div class="flex flex-wrap items-center justify-between gap-2">
                    <p class="font-mono text-[0.62rem] uppercase tracking-[0.17em] text-muted-foreground">
                        [ new key ]
                    </p>
                    <p class="font-mono text-[0.55rem] uppercase tracking-[0.1em] text-muted-foreground">
                        full key is shown once after create
                    </p>
                </div>
                <div class="grid gap-3 lg:grid-cols-[minmax(12rem,1fr)_auto] lg:items-center">
                    <Input bind:value={createName} aria-label="API key name" placeholder="key name · e.g. ci-runner" />
                    <div class="flex flex-wrap items-center gap-1">
                        <span class="mr-1 font-mono text-[0.58rem] uppercase tracking-[0.12em] text-muted-foreground">env</span>
                        {#each environments as environment (environment.id)}
                            <button
                                type="button"
                                class={cn(
                                    "ios-corners-xs inline-flex cursor-pointer items-center gap-2 border px-3 py-2 text-xs transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
                                    createEnvironmentID === environment.id
                                        ? "border-border-strong bg-surface-active text-foreground"
                                        : "border-border-control text-muted-foreground hover:bg-surface-hover hover:text-foreground",
                                )}
                                aria-pressed={createEnvironmentID === environment.id}
                                onclick={() => (createEnvironmentID = environment.id)}
                            >
                                <Circle class={cn("h-2 w-2 fill-current", environmentIndicator(environment.key))} />
                                {environment.key}
                            </button>
                        {/each}
                    </div>
                </div>
                {#if createError}<p class="text-sm text-destructive" aria-live="polite">{createError}</p>{/if}
                <div class="flex gap-2">
                    <Button type="submit" size="sm" variant="solid" disabled={!createName.trim() || creating}>
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
                <Plus class="h-3.5 w-3.5" /> new key
            </button>
        {/if}
    </Card>

    <p class="flex items-center gap-2 font-mono text-[0.58rem] uppercase tracking-[0.12em] text-muted-foreground">
        <ShieldAlert class="h-3.5 w-3.5" /> revoked keys stop evaluating immediately · key prefixes are safe to share
    </p>
</section>

<DestructiveDialog
    bind:open={revokeOpen}
    title="Revoke API key"
    description="Requests using this key will stop working immediately."
    details={revokeTarget ? `${revokeTarget.name}\n${revokeTarget.prefix}` : null}
    actionLabel="revoke key"
    submitting={revoking}
    error={revokeError}
    onconfirm={revokeKey}
    oncancel={() => {
        revokeTarget = null;
        revokeError = null;
    }}
/>
