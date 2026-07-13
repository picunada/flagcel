<script lang="ts">
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
    import PageHeader from "$lib/components/ui/page-header.svelte";
    import SegmentedControl from "$lib/components/ui/segmented-control.svelte";
    import { cn } from "$lib/utils";
    import type { PageProps } from "./$types";

    type StatusFilter = "active" | "revoked" | "all";

    let { data }: PageProps = $props();

    const selectedEnvironment = $derived(data.selectedEnvironment);
    const keys = $derived<APIKey[]>(
        data.keys.filter((key) => key.environment_id === selectedEnvironment.id),
    );
    const activeCount = $derived(keys.filter((key) => !key.revoked_at).length);
    const revokedCount = $derived(keys.filter((key) => Boolean(key.revoked_at)).length);
    const statusFilterOptions = $derived([
        { value: "active" as const, label: `active · ${activeCount}` },
        { value: "revoked" as const, label: `revoked · ${revokedCount}` },
        { value: "all" as const, label: `all · ${keys.length}` },
    ]);
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
    let creating = $state(false);
    let createError = $state<string | null>(null);
    let created = $state<CreateAPIKeyResponse | null>(null);
    let revealCopied = $state(false);
    let revokeOpen = $state(false);
    let revokeTarget = $state<APIKey | null>(null);
    let revoking = $state(false);
    let revokeError = $state<string | null>(null);

    function openCreate() {
        createOpen = true;
        createError = null;
    }

    function closeCreate() {
        createOpen = false;
        createName = "";
        createError = null;
    }

    async function createKey() {
        if (!createName.trim() || creating) return;
        creating = true;
        createError = null;
        created = null;
        try {
            created = await api.createAPIKey(createName.trim(), selectedEnvironment.id);
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

</script>

{#snippet headerActions()}
    <SegmentedControl
        label="API key status filter"
        options={statusFilterOptions}
        value={filter}
        onchange={(value) => (filter = value)}
    />
{/snippet}

<section class="motion-page space-y-6">
    <PageHeader
        eyebrow={`[ api keys · ${selectedEnvironment.key} · ${keys.length} ]`}
        actions={headerActions}
    />

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
            class="hidden min-h-10 grid-cols-[1.25rem_minmax(12rem,1fr)_9rem_8rem_7rem] items-center gap-3 border-b border-border-divider bg-surface-faint px-4 font-mono text-[0.58rem] uppercase tracking-[0.16em] text-muted-foreground xl:grid"
        >
            <span></span>
            <span>name</span>
            <span>last used</span>
            <span>created</span>
            <span></span>
        </div>

        <div class="motion-list">
            {#each filteredKeys as key (key.id)}
                {@const revoked = Boolean(key.revoked_at)}
                <div
                    class={cn(
                        "grid gap-3 border-b border-border-divider px-4 py-4 transition-colors last:border-0 hover:bg-surface-hover xl:grid-cols-[1.25rem_minmax(12rem,1fr)_9rem_8rem_7rem] xl:items-center xl:py-3",
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
                    <div class="ios-corners-xs inline-flex items-center gap-2 border border-border-strong bg-surface-active px-3 py-2 text-xs text-foreground">
                        <span class="font-mono text-[0.58rem] uppercase tracking-[0.12em] text-muted-foreground">env</span>
                        <span class="font-mono">{selectedEnvironment.key}</span>
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
