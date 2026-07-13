<script lang="ts">
    import Button from "$lib/components/ui/button.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";
    import "../app.css";

    let { status, error }: { status: number; error: App.Error } = $props();
    const message = $derived(
        error?.message ||
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
        <PageHeader
            eyebrow={isBackendError ? "[ backend unavailable ]" : `[ error ${status} ]`}
            title={isBackendError ? "Flagcel cannot reach the backend." : "Flagcel hit an error."}
            description={message}
            titleId="error-title"
        />
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
