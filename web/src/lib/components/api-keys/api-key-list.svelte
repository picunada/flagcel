<script lang="ts">
    import { KeyRound, Trash2 } from "lucide-svelte";
    import type { APIKey, Environment } from "$lib/api";
    import Badge from "$lib/components/ui/badge.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import EmptyState from "$lib/components/ui/empty-state.svelte";
    import SectionHeader from "$lib/components/ui/section-header.svelte";

    let {
        keys,
        selectedEnvironment,
        formatDate,
        onrequestRevoke,
    }: {
        keys: APIKey[];
        selectedEnvironment: Environment;
        formatDate: (value?: string) => string;
        onrequestRevoke: (key: APIKey) => void;
    } = $props();
</script>

<div class="space-y-3">
    <SectionHeader>keys · {selectedEnvironment.key} · {keys.length}</SectionHeader>
    {#if keys.length === 0}
        <EmptyState
            icon={KeyRound}
            title="[ no keys in {selectedEnvironment.key} yet ]"
        />
    {:else}
        <div class="motion-list space-y-3">
            {#each keys as key (key.id)}
                <Card
                    class="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between"
                >
                    <div class="min-w-0 space-y-2">
                        <div class="flex items-center gap-2">
                            <p class="truncate text-base">{key.name}</p>
                            {#if key.revoked_at}
                                <Badge variant="muted">revoked</Badge>
                            {:else}
                                <Badge variant="success" dot>active</Badge>
                            {/if}
                        </div>
                        <p class="text-xs text-muted-foreground">
                            <span class="font-mono">{key.prefix}</span> · created {formatDate(key.created_at)} · last used
                            {formatDate(key.last_used_at)}
                        </p>
                    </div>
                    {#if !key.revoked_at}
                        <Button
                            variant="destructive"
                            size="sm"
                            onclick={() => onrequestRevoke(key)}
                        >
                            <Trash2 class="h-3.5 w-3.5" /> revoke
                        </Button>
                    {/if}
                </Card>
            {/each}
        </div>
    {/if}
</div>
