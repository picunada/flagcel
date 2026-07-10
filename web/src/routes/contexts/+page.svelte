<script lang="ts">
    import { Boxes } from "lucide-svelte";
    import Button from "$lib/components/ui/button.svelte";
</script>

<section class="flex min-h-full items-center justify-center p-8 text-center">
    <div class="max-w-sm">
        <Boxes class="mx-auto h-8 w-8 text-muted-foreground" aria-hidden="true" />
        <h1 class="mt-5 text-xl font-semibold">No contexts yet</h1>
        <p class="mt-2 text-sm leading-6 text-muted-foreground">
            Define a data shape to enable typed CEL rules and payload validation.
        </p>
        <Button href="/contexts/new" variant="solid" class="mt-6">new context</Button>
    </div>
</section>
