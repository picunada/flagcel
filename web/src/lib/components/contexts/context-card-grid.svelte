<script lang="ts">
    import type { ContextSchema } from "$lib/api";
    import Card from "$lib/components/ui/card.svelte";

    let { contexts }: { contexts: ContextSchema[] } = $props();
</script>

<div class="motion-list grid gap-3 sm:grid-cols-2">
    {#each contexts as ctx (ctx.id)}
        <a href="/contexts/{encodeURIComponent(ctx.id)}" class="group block">
            <Card hoverable class="flex h-full flex-col gap-3 p-5">
                <p class="truncate text-base font-medium">{ctx.name}</p>
                {#if ctx.description}
                    <p class="line-clamp-2 text-sm text-foreground-softer">
                        {ctx.description}
                    </p>
                {/if}
                <p
                    class="mt-auto text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
                >
                    {ctx.fields.length} field{ctx.fields.length === 1 ? "" : "s"}
                </p>
                {#if ctx.fields.length > 0}
                    <div class="space-y-1">
                        {#each ctx.fields.slice(0, 3) as f (f.path)}
                            <p
                                class="truncate border-l-2 border-border-muted pl-2.5 font-mono text-xs text-muted-foreground"
                            >
                                {f.path}
                                <span class="text-muted-foreground/60">· {f.type}</span>
                            </p>
                        {/each}
                        {#if ctx.fields.length > 3}
                            <p
                                class="pl-2.5 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
                            >
                                +{ctx.fields.length - 3} more
                            </p>
                        {/if}
                    </div>
                {/if}
            </Card>
        </a>
    {/each}
</div>
