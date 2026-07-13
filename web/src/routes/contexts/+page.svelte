<script lang="ts">
    import { Boxes } from "lucide-svelte";
    import Button from "$lib/components/ui/button.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";
</script>

<section class="motion-page min-h-full space-y-10 p-5 sm:p-7 lg:p-8">
    <PageHeader
        eyebrow="[ contexts · 0 ]"
    />

    <div class="mx-auto max-w-sm text-center">
        <Boxes class="mx-auto h-8 w-8 text-muted-foreground" aria-hidden="true" />
        <h2 class="mt-5 text-xl font-semibold">No contexts yet</h2>
        <p class="mt-2 text-sm leading-6 text-muted-foreground">
            Define a data shape to enable typed CEL rules and payload validation.
        </p>
        <Button href="/contexts/new" variant="solid" class="mt-6">new context</Button>
    </div>
</section>
