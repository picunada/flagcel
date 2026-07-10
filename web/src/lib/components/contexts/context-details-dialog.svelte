<script lang="ts">
    import { tick } from "svelte";
    import { Pencil, X } from "lucide-svelte";
    import Button from "$lib/components/ui/button.svelte";
    import FieldLabel from "$lib/components/ui/field-label.svelte";
    import Input from "$lib/components/ui/input.svelte";

    let {
        open = $bindable(false),
        initialName,
        initialDescription,
        submitting = false,
        error = null,
        onsave,
    }: {
        open?: boolean;
        initialName: string;
        initialDescription: string;
        submitting?: boolean;
        error?: string | null;
        onsave: (name: string, description: string) => void | Promise<void>;
    } = $props();

    let name = $state("");
    let description = $state("");
    let dialogEl: HTMLElement | undefined = $state();
    let inputEl: HTMLInputElement | null = $state(null);

    $effect(() => {
        if (!open) return;
        name = initialName;
        description = initialDescription;
        void tick().then(() => (inputEl ?? dialogEl)?.focus());
    });

    function close() {
        if (!submitting) open = false;
    }

    function submit(event: SubmitEvent) {
        event.preventDefault();
        if (!name.trim() || submitting) return;
        void onsave(name.trim(), description.trim());
    }
</script>

{#if open}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <button
            type="button"
            class="absolute inset-0 cursor-default bg-black/70 backdrop-blur-sm"
            aria-label="Close context details"
            onclick={close}
        ></button>
        <div
            class="glass-panel motion-pop relative z-10 w-full max-w-lg rounded-md p-5"
            role="dialog"
            aria-modal="true"
            aria-labelledby="context-details-title"
            tabindex="-1"
            bind:this={dialogEl}
            onkeydown={(event) => event.key === "Escape" && close()}
        >
            <form onsubmit={submit}>
                <div class="flex items-start gap-4">
                    <span class="flex h-9 w-9 items-center justify-center rounded-md border border-border-strong bg-surface-active text-foreground">
                        <Pencil class="h-4 w-4" />
                    </span>
                    <div class="min-w-0 flex-1">
                        <h2 id="context-details-title" class="text-lg font-medium">Edit context details</h2>
                        <p class="mt-1 text-sm text-muted-foreground">Rename the context or update its description.</p>
                    </div>
                    <button
                        type="button"
                        class="cursor-pointer rounded-md p-1 text-muted-foreground transition-colors hover:bg-surface-hover hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                        aria-label="Close"
                        onclick={close}
                    >
                        <X class="h-4 w-4" />
                    </button>
                </div>

                <div class="mt-6 space-y-4">
                    <label class="block space-y-2">
                        <FieldLabel>name</FieldLabel>
                        <Input bind:ref={inputEl} bind:value={name} required spellcheck={false} />
                    </label>
                    <label class="block space-y-2">
                        <FieldLabel>description · optional</FieldLabel>
                        <Input bind:value={description} />
                    </label>
                </div>

                {#if error}<p class="mt-4 text-sm text-destructive" aria-live="polite">{error}</p>{/if}

                <div class="mt-6 flex justify-end gap-2 border-t border-border-divider pt-4">
                    <Button type="button" variant="ghost" disabled={submitting} onclick={close}>cancel</Button>
                    <Button type="submit" variant="solid" disabled={!name.trim() || submitting}>
                        {submitting ? "saving…" : "save details"}
                    </Button>
                </div>
            </form>
        </div>
    </div>
{/if}
