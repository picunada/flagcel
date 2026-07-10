<script lang="ts">
    import { goto } from "$app/navigation";
    import { Plus, Search, Users } from "lucide-svelte";
    import Button from "$lib/components/ui/button.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import { cn } from "$lib/utils";
    import type { LayoutProps } from "./$types";

    let { data, children }: LayoutProps = $props();
    let query = $state("");

    const contexts = $derived(data.contexts);
    const references = $derived(data.contextReferences);
    const pathname = $derived(data.currentPathname);
    const selectedID = $derived.by(() => {
        if (!pathname.startsWith("/contexts/") || pathname === "/contexts/new") return "";
        return decodeURIComponent(pathname.slice("/contexts/".length));
    });
    const filtered = $derived.by(() => {
        const needle = query.trim().toLowerCase();
        return contexts.filter(
            (context) =>
                !needle ||
                context.name.toLowerCase().includes(needle) ||
                context.description?.toLowerCase().includes(needle),
        );
    });

    function contextReferences(id: string) {
        return references.filter((reference) => reference.context_id === id);
    }

    function ruleCount(id: string) {
        return contextReferences(id).reduce((total, reference) => total + reference.rule_count, 0);
    }

    function selectContext(event: Event) {
        const id = (event.currentTarget as HTMLSelectElement).value;
        if (id === "new") goto("/contexts/new");
        else if (id) goto(`/contexts/${encodeURIComponent(id)}`);
    }
</script>

<div class="flex h-full min-h-0 flex-col xl:flex-row">
    <aside
        class="hidden w-[17.5rem] shrink-0 flex-col border-r border-border-divider bg-sidebar/40 xl:flex"
        aria-label="Contexts"
    >
        <div class="space-y-4 border-b border-border-divider px-5 pb-4 pt-6">
            <div>
                <p class="font-mono text-[0.65rem] uppercase tracking-[0.2em] text-muted-foreground">
                    [ contexts · {contexts.length} ]
                </p>
                <p class="mt-2 text-xs text-foreground-softer">
                    Data shapes available to flag rules
                </p>
            </div>
            <div class="relative">
                <Search
                    class="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
                    aria-hidden="true"
                />
                <Input
                    bind:value={query}
                    aria-label="Filter contexts"
                    placeholder="filter contexts…"
                    class="pl-9 font-mono text-xs"
                />
            </div>
        </div>

        <nav class="motion-list min-h-0 flex-1 space-y-1.5 overflow-y-auto p-3" aria-label="Context schemas">
            {#each filtered as context (context.id)}
                {@const usage = contextReferences(context.id)}
                {@const active = context.id === selectedID}
                <a
                    href="/contexts/{encodeURIComponent(context.id)}"
                    aria-current={active ? "page" : undefined}
                    class={cn(
                        "ios-corners group block cursor-pointer border px-3.5 py-3 transition-colors duration-200",
                        active
                            ? "border-border-strong bg-surface-active"
                            : "border-transparent hover:border-border-hover hover:bg-surface-hover",
                    )}
                >
                    <span class="flex items-center justify-between gap-3">
                        <span class="min-w-0 truncate font-mono text-sm font-medium text-foreground">
                            {context.name}
                        </span>
                        <span
                            class={cn(
                                "shrink-0 font-mono text-[0.55rem] uppercase tracking-[0.12em]",
                                usage.length ? "text-valid" : "text-muted-foreground",
                            )}
                        >
                            {usage.length ? "● in use" : "unused"}
                        </span>
                    </span>
                    <span class="mt-1.5 block font-mono text-[0.6rem] uppercase tracking-[0.08em] text-muted-foreground">
                        {context.fields.length} fields · {usage.length} flags{usage.length
                            ? ` · ${ruleCount(context.id)} rules`
                            : ""}
                    </span>
                </a>
            {:else}
                <p class="px-3 py-8 text-center text-xs text-muted-foreground">
                    No matching contexts
                </p>
            {/each}
        </nav>

        <div class="border-t border-border-divider p-3">
            <Button href="/contexts/new" variant="solid" class="w-full">
                <Plus class="h-3.5 w-3.5" /> new context
            </Button>
        </div>
    </aside>

    <div class="flex min-h-0 min-w-0 flex-1 flex-col">
        <div class="flex items-center gap-2 border-b border-border-divider bg-sidebar/30 p-3 xl:hidden">
            <div class="relative min-w-0 flex-1">
                <Users
                    class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                    aria-hidden="true"
                />
                <select
                    class="ios-corners-sm h-10 w-full cursor-pointer appearance-none border border-input bg-background pl-10 pr-8 text-sm text-foreground outline-none transition-colors focus-visible:border-border-hover focus-visible:ring-1 focus-visible:ring-ring"
                    value={pathname === "/contexts/new" ? "new" : selectedID}
                    aria-label="Selected context"
                    onchange={selectContext}
                >
                    {#if !selectedID && pathname !== "/contexts/new"}<option value="">contexts</option>{/if}
                    {#each contexts as context (context.id)}
                        <option value={context.id}>{context.name}</option>
                    {/each}
                    <option value="new">New context</option>
                </select>
            </div>
            <Button href="/contexts/new" size="icon" variant="solid" aria-label="New context">
                <Plus class="h-4 w-4" />
            </Button>
        </div>

        <div class="min-h-0 min-w-0 flex-1 overflow-y-auto">
            {@render children()}
        </div>
    </div>
</div>
