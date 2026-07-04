<script lang="ts">
    import { Check, Layers3, Pencil, Trash2, X } from "lucide-svelte";
    import type { Environment } from "$lib/api";
    import Badge from "$lib/components/ui/badge.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import EmptyState from "$lib/components/ui/empty-state.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import SectionHeader from "$lib/components/ui/section-header.svelte";

    let {
        environments,
        editingId,
        editKey = $bindable(""),
        editName = $bindable(""),
        editDescription = $bindable(""),
        saving = false,
        editError = null,
        isDefault,
        formatDate,
        onstartEdit,
        oncancelEdit,
        onsaveEdit,
        onrequestDelete,
    }: {
        environments: Environment[];
        editingId: string | null;
        editKey: string;
        editName: string;
        editDescription: string;
        saving?: boolean;
        editError?: string | null;
        isDefault: (env: Environment) => boolean;
        formatDate: (value?: string) => string;
        onstartEdit: (env: Environment) => void;
        oncancelEdit: () => void;
        onsaveEdit: (env: Environment) => void | Promise<void>;
        onrequestDelete: (env: Environment) => void;
    } = $props();
</script>

<div class="space-y-3">
    <SectionHeader>environments · {environments.length}</SectionHeader>
    {#if environments.length === 0}
        <EmptyState icon={Layers3} title="[ no environments yet ]" />
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
                            <Input
                                bind:value={editDescription}
                                placeholder="description (optional)"
                            />
                            {#if isDefault(env)}
                                <p
                                    class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
                                >
                                    the default environment key cannot be changed
                                </p>
                            {/if}
                            {#if editError}
                                <p class="text-sm text-destructive">{editError}</p>
                            {/if}
                            <div class="flex items-center gap-2">
                                <Button
                                    size="sm"
                                    onclick={() => onsaveEdit(env)}
                                    disabled={saving}
                                >
                                    <Check class="h-3.5 w-3.5" /> save
                                </Button>
                                <Button
                                    size="sm"
                                    variant="ghost"
                                    onclick={oncancelEdit}
                                    disabled={saving}
                                >
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
                                    <p class="text-sm text-muted-foreground">
                                        {env.description}
                                    </p>
                                {/if}
                                <p class="text-xs text-muted-foreground">
                                    created {formatDate(env.created_at)} · updated {formatDate(env.updated_at)}
                                </p>
                            </div>
                            <div class="flex shrink-0 items-center gap-2">
                                <Button
                                    size="sm"
                                    variant="ghost"
                                    onclick={() => onstartEdit(env)}
                                >
                                    <Pencil class="h-3.5 w-3.5" /> edit
                                </Button>
                                {#if !isDefault(env)}
                                    <Button
                                        size="sm"
                                        variant="destructive"
                                        onclick={() => onrequestDelete(env)}
                                    >
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
