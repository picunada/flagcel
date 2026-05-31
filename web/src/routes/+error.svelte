<script lang="ts">
    import { page } from "$app/state";
    import Button from "$lib/components/ui/button.svelte";
    import "../app.css";

    const status = $derived(page.status);
    const message = $derived(
        page.error?.message ||
            "Something went wrong while loading Flagcel.",
    );
    const isBackendError = $derived(
        status === 0 ||
            status >= 500 ||
            message.toLowerCase().includes("backend server"),
    );

    function retry() {
        window.location.reload();
    }
</script>

<svelte:head>
    <title>{isBackendError ? "Backend unavailable" : "Error"} · Flagcel</title>
</svelte:head>

<main
    class="flex min-h-screen items-center justify-center bg-background px-6 py-12 text-foreground"
>
    <section
        class="glass-panel motion-panel w-full max-w-xl rounded-sm p-6 sm:p-8"
        aria-labelledby="error-title"
    >
        <p class="mb-3 text-xs uppercase tracking-[0.18em] text-muted-foreground">
            {isBackendError ? "backend unavailable" : `error ${status}`}
        </p>
        <h1
            id="error-title"
            class="text-balance text-3xl font-normal leading-tight sm:text-4xl"
        >
            {isBackendError ? "Flagcel cannot reach the backend." : "Flagcel hit an error."}
        </h1>
        <p class="mt-4 text-sm leading-6 text-[rgba(255,255,255,0.78)]">
            {message}
        </p>
        {#if isBackendError}
            <p class="mt-3 text-sm leading-6 text-muted-foreground">
                Start or restart the backend server, then retry this page.
            </p>
        {/if}

        <div class="mt-6 flex flex-wrap gap-3">
            <Button onclick={retry}>Retry</Button>
            <Button href="/" variant="ghost">Go to flags</Button>
        </div>
    </section>
</main>
