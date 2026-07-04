<script lang="ts">
    import { Plus } from "lucide-svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import SectionHeader from "$lib/components/ui/section-header.svelte";

    let {
        environmentKey = $bindable(""),
        name = $bindable(""),
        description = $bindable(""),
        creating = false,
        error = null,
        oncreate,
    }: {
        environmentKey: string;
        name: string;
        description: string;
        creating?: boolean;
        error?: string | null;
        oncreate: () => void | Promise<void>;
    } = $props();
</script>

<Card class="motion-panel space-y-4 p-5">
    <SectionHeader>new environment</SectionHeader>
    <div class="grid gap-2 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)_auto]">
        <Input
            bind:value={environmentKey}
            placeholder="key · e.g. staging"
            class="font-mono"
            onkeydown={(e) => e.key === "Enter" && oncreate()}
        />
        <Input
            bind:value={name}
            placeholder="name (optional)"
            onkeydown={(e) => e.key === "Enter" && oncreate()}
        />
        <Button onclick={oncreate} disabled={creating || !environmentKey.trim()}>
            <Plus class="h-3.5 w-3.5" /> create
        </Button>
    </div>
    <Input bind:value={description} placeholder="description (optional)" />
    <p class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
        key · lowercase letters, numbers, and hyphens
    </p>
    {#if error}
        <p class="text-sm text-destructive">{error}</p>
    {/if}
</Card>
