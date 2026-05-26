<script lang="ts">
    import { Check, Copy, KeyRound, Plus, Trash2 } from "lucide-svelte";
    import { api, APIError, type APIKey, type CreateAPIKeyResponse } from "$lib/api";
    import { invalidateAll } from "$app/navigation";
    import Badge from "$lib/components/ui/badge.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import DestructiveDialog from "$lib/components/ui/destructive-dialog.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import SectionHeader from "$lib/components/ui/section-header.svelte";
    import type { PageProps } from "./$types";

    let { data }: PageProps = $props();
    const keys = $derived<APIKey[]>(data.keys);

    let name = $state("");
    let creating = $state(false);
    let error = $state<string | null>(null);
    let created = $state<CreateAPIKeyResponse | null>(null);
    let copied = $state(false);
    let revokeOpen = $state(false);
    let revokeTarget = $state<APIKey | null>(null);
    let revoking = $state(false);
    let revokeError = $state<string | null>(null);

    async function createKey() {
        if (!name.trim()) return;
        creating = true;
        error = null;
        created = null;
        try {
            created = await api.createAPIKey(name);
            name = "";
            await invalidateAll();
        } catch (e) {
            error = e instanceof APIError ? e.message : "Failed to create API key";
        } finally {
            creating = false;
        }
    }

    function requestRevokeKey(key: APIKey) {
        revokeTarget = key;
        revokeError = null;
        revokeOpen = true;
    }

    async function revokeKey() {
        const key = revokeTarget;
        if (!key) return;
        revoking = true;
        revokeError = null;
        try {
            await api.revokeAPIKey(key.id);
            revokeOpen = false;
            revokeTarget = null;
            await invalidateAll();
        } catch (e) {
            revokeError = e instanceof APIError ? e.message : "Failed to revoke API key";
        } finally {
            revoking = false;
        }
    }

    async function copyToken() {
        if (!created?.token) return;
        await navigator.clipboard.writeText(created.token);
        copied = true;
        setTimeout(() => (copied = false), 1200);
    }

    function formatDate(value?: string) {
        if (!value) return "never";
        return new Date(value).toLocaleString();
    }
</script>

<section class="space-y-10">
    <header class="space-y-3">
        <p class="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">
            eval access
        </p>
        <h1 class="font-mono text-3xl font-normal leading-tight sm:text-4xl">
            API keys
        </h1>
    </header>

    <Card class="motion-panel space-y-4 p-5">
        <SectionHeader>new key</SectionHeader>
        <div class="flex flex-col gap-2 sm:flex-row">
            <Input bind:value={name} placeholder="key name" onkeydown={(e) => e.key === "Enter" && createKey()} />
            <Button onclick={createKey} disabled={creating || !name.trim()}>
                <Plus class="h-3.5 w-3.5" /> create
            </Button>
        </div>
        {#if created}
            <div
                class="space-y-3 rounded-sm border border-[rgba(255,255,255,0.14)] bg-[rgba(255,255,255,0.035)] p-4"
            >
                <div class="flex items-center justify-between gap-3">
                    <p class="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
                        copy this token now
                    </p>
                    <Button size="sm" variant="ghost" onclick={copyToken}>
                        {#if copied}
                            <Check class="h-3.5 w-3.5" /> copied
                        {:else}
                            <Copy class="h-3.5 w-3.5" /> copy
                        {/if}
                    </Button>
                </div>
                <code
                    class="block overflow-x-auto whitespace-nowrap rounded-sm bg-black/30 p-3 font-mono text-xs text-[rgba(255,255,255,0.86)]"
                >
                    {created.token}
                </code>
            </div>
        {/if}
        {#if error}
            <p class="text-sm text-destructive">{error}</p>
        {/if}
    </Card>

    <div class="space-y-3">
        <SectionHeader>keys · {keys.length}</SectionHeader>
        {#if keys.length === 0}
            <Card class="motion-panel p-10 text-center">
                <KeyRound class="mx-auto h-5 w-5 text-muted-foreground" />
                <p class="mt-4 font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
                    [ no keys yet ]
                </p>
            </Card>
        {:else}
            <div class="motion-list space-y-3">
                {#each keys as key (key.id)}
                    <Card class="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
                        <div class="min-w-0 space-y-2">
                            <div class="flex items-center gap-2">
                                <p class="truncate font-mono text-base">{key.name}</p>
                                {#if key.revoked_at}
                                    <Badge variant="muted">revoked</Badge>
                                {:else}
                                    <Badge variant="success" dot>active</Badge>
                                {/if}
                            </div>
                            <p class="font-mono text-xs text-muted-foreground">
                                {key.prefix} · created {formatDate(key.created_at)} · last used
                                {formatDate(key.last_used_at)}
                            </p>
                        </div>
                        {#if !key.revoked_at}
                            <Button variant="destructive" size="sm" onclick={() => requestRevokeKey(key)}>
                                <Trash2 class="h-3.5 w-3.5" /> revoke
                            </Button>
                        {/if}
                    </Card>
                {/each}
            </div>
        {/if}
    </div>
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
