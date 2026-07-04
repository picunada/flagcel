<script lang="ts">
    import Button from "$lib/components/ui/button.svelte";

    let { message }: { message: string } = $props();

    function retry() {
        window.location.reload();
    }
</script>

<section
    class="mx-auto max-w-xl space-y-5 text-center"
    aria-labelledby="backend-unavailable-title"
>
    <div class="glass-panel motion-panel rounded-sm p-6 sm:p-8">
        <p class="mb-3 text-xs uppercase tracking-[0.18em] text-muted-foreground">
            backend unavailable
        </p>
        <h1
            id="backend-unavailable-title"
            class="text-balance text-3xl font-normal leading-tight sm:text-4xl"
        >
            Flagcel cannot reach the backend.
        </h1>
        <p class="mx-auto mt-4 max-w-md text-sm leading-6 text-foreground-soft">
            {message}
        </p>
        <p
            class="mx-auto mt-3 max-w-md text-sm leading-6 text-muted-foreground"
        >
            Start or restart the backend server, then retry this page.
        </p>
        <div class="mt-6 flex justify-center">
            <Button onclick={retry}>Retry</Button>
        </div>
    </div>
</section>
