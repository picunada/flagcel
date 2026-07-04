<script lang="ts">
    import { api, APIError, type Environment } from "$lib/api";
    import { invalidateAll } from "$app/navigation";
    import EnvironmentCreateCard from "$lib/components/environments/environment-create-card.svelte";
    import EnvironmentList from "$lib/components/environments/environment-list.svelte";
    import DestructiveDialog from "$lib/components/ui/destructive-dialog.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";
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
    <PageHeader
        eyebrow="workspaces · isolated flag sets"
        title="Environments"
        description="Each environment holds its own flags, rules, and API keys. Switch the active environment from the selector in the header."
    />

    <EnvironmentCreateCard
        bind:environmentKey={key}
        bind:name
        bind:description
        {creating}
        error={createError}
        oncreate={createEnvironment}
    />

    <EnvironmentList
        {environments}
        {editingId}
        bind:editKey
        bind:editName
        bind:editDescription
        {saving}
        {editError}
        {isDefault}
        {formatDate}
        onstartEdit={startEdit}
        oncancelEdit={cancelEdit}
        onsaveEdit={saveEdit}
        onrequestDelete={requestDelete}
    />
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
