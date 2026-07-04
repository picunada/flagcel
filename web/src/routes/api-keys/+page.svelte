<script lang="ts">
    import { api, APIError, type APIKey, type CreateAPIKeyResponse, type Environment } from "$lib/api";
    import { invalidateAll } from "$app/navigation";
    import APIKeyCreateCard from "$lib/components/api-keys/api-key-create-card.svelte";
    import APIKeyList from "$lib/components/api-keys/api-key-list.svelte";
    import DestructiveDialog from "$lib/components/ui/destructive-dialog.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";
    import type { PageProps } from "./$types";

    let { data }: PageProps = $props();
    const selectedEnvironment = $derived<Environment>(data.selectedEnvironment);
    const keys = $derived<APIKey[]>(
        data.keys.filter((key) => key.environment_id === selectedEnvironment.id),
    );

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
            created = await api.createAPIKey(name, selectedEnvironment.id);
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
    <PageHeader eyebrow="eval access" title="API keys" />

    <APIKeyCreateCard
        bind:name
        {creating}
        {error}
        {created}
        {copied}
        oncreate={createKey}
        oncopy={copyToken}
    />

    <APIKeyList
        {keys}
        {selectedEnvironment}
        {formatDate}
        onrequestRevoke={requestRevokeKey}
    />
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
