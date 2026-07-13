<script lang="ts">
    import Button from "$lib/components/ui/button.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";

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
        <PageHeader
            eyebrow="[ backend unavailable ]"
            title="Flagcel cannot reach the backend."
            description={message}
            titleId="backend-unavailable-title"
            class="items-center text-center sm:items-center sm:justify-center"
        />
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
