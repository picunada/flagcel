<script lang="ts">
    import { Check, Copy, Plus } from "lucide-svelte";
    import type { CreateAPIKeyResponse } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import CodeBlock from "$lib/components/ui/code-block.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import SectionHeader from "$lib/components/ui/section-header.svelte";

    let {
        name = $bindable(""),
        creating = false,
        error = null,
        created = null,
        copied = false,
        oncreate,
        oncopy,
    }: {
        name: string;
        creating?: boolean;
        error?: string | null;
        created?: CreateAPIKeyResponse | null;
        copied?: boolean;
        oncreate: () => void | Promise<void>;
        oncopy: () => void | Promise<void>;
    } = $props();
</script>

<Card class="motion-panel space-y-4 p-5">
    <SectionHeader>new key</SectionHeader>
    <div class="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
        <Input
            bind:value={name}
            placeholder="key name"
            onkeydown={(e) => e.key === "Enter" && oncreate()}
        />
        <Button onclick={oncreate} disabled={creating || !name.trim()}>
            <Plus class="h-3.5 w-3.5" /> create
        </Button>
    </div>
    {#if created}
        <div class="space-y-3 rounded-sm border border-border-soft bg-surface-muted p-4">
            <div class="flex items-center justify-between gap-3">
                <p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">
                    copy this token now
                </p>
                <Button size="sm" variant="ghost" onclick={oncopy}>
                    {#if copied}
                        <Check class="h-3.5 w-3.5" /> copied
                    {:else}
                        <Copy class="h-3.5 w-3.5" /> copy
                    {/if}
                </Button>
            </div>
            <CodeBlock value={created.token} />
        </div>
    {/if}
    {#if error}
        <p class="text-sm text-destructive">{error}</p>
    {/if}
</Card>
