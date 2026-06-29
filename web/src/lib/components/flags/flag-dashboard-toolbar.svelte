<script lang="ts">
    import Button from "$lib/components/ui/button.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import ThemedSelect from "$lib/components/ui/themed-select.svelte";
    import { cn } from "$lib/utils";
    import {
        LayoutGrid,
        ListFilter,
        Plus,
        Search,
        TableProperties,
    } from "lucide-svelte";

    type StatusFilter = "all" | "on" | "off";
    type SortKey =
        | "recent"
        | "rules"
        | "default"
        | "status"
        | "key"
        | "context";
    type ViewMode = "table" | "cards";
    type SelectOption = { value: string; label: string };

    type Props = {
        query: string;
        statusFilter: StatusFilter;
        contextFilter: string;
        sortKey: SortKey;
        viewMode: ViewMode;
        contextFilterOptions: SelectOption[];
        sortOptions: SelectOption[];
        shownCount: number;
        newFlagHref: string;
    };

    let {
        query = $bindable(),
        statusFilter = $bindable(),
        contextFilter = $bindable(),
        sortKey = $bindable(),
        viewMode = $bindable(),
        contextFilterOptions,
        sortOptions,
        shownCount,
        newFlagHref,
    }: Props = $props();
</script>

<div class="border-b border-border p-3">
    <div class="flex flex-col gap-3">
        <div
            class="grid items-center gap-2 md:grid-cols-[minmax(14rem,1fr)_auto] lg:grid-cols-[minmax(14rem,1fr)_auto_12rem_11rem]"
        >
            <label class="relative block min-w-0">
                <span class="sr-only">Search flags</span>
                <Search
                    class="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
                />
                <Input
                    bind:value={query}
                    placeholder="search key, context, rule"
                    class="w-full pl-8"
                />
            </label>

            <div
                class="flex h-9 shrink-0 rounded-sm border border-border p-0.5"
                aria-label="Status filter"
            >
                <button
                    type="button"
                    aria-pressed={statusFilter === "all"}
                    class={cn(
                        "h-full rounded-[2px] px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors",
                        statusFilter === "all"
                            ? "bg-surface-selected text-foreground"
                            : "text-muted-foreground hover:text-foreground",
                    )}
                    onclick={() => (statusFilter = "all")}
                >
                    all
                </button>
                <button
                    type="button"
                    aria-pressed={statusFilter === "on"}
                    class={cn(
                        "h-full rounded-[2px] px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors",
                        statusFilter === "on"
                            ? "bg-surface-selected text-foreground"
                            : "text-muted-foreground hover:text-foreground",
                    )}
                    onclick={() => (statusFilter = "on")}
                >
                    on
                </button>
                <button
                    type="button"
                    aria-pressed={statusFilter === "off"}
                    class={cn(
                        "h-full rounded-[2px] px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors",
                        statusFilter === "off"
                            ? "bg-surface-selected text-foreground"
                            : "text-muted-foreground hover:text-foreground",
                    )}
                    onclick={() => (statusFilter = "off")}
                >
                    off
                </button>
            </div>

            <label class="block">
                <span class="sr-only">Context filter</span>
                <ThemedSelect
                    value={contextFilter}
                    options={contextFilterOptions}
                    onchange={(v) => (contextFilter = v)}
                />
            </label>

            <label class="relative block">
                <span class="sr-only">Sort flags</span>
                <ListFilter
                    class="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
                />
                <ThemedSelect
                    value={sortKey}
                    options={sortOptions}
                    onchange={(v) => (sortKey = v as SortKey)}
                    buttonClass="pl-8 pr-2.5"
                />
            </label>
        </div>

        <div class="flex items-center justify-between gap-2">
            <div
                class="hidden text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground sm:block"
                aria-live="polite"
            >
                {shownCount} shown
            </div>
            <div class="ml-auto flex items-center gap-2">
                <div
                    class="flex h-9 rounded-sm border border-border p-0.5"
                    aria-label="View mode"
                >
                    <button
                        type="button"
                        title="Table view"
                        aria-label="Table view"
                        aria-pressed={viewMode === "table"}
                        class={cn(
                            "inline-flex h-full w-8 items-center justify-center rounded-[2px] transition-colors",
                            viewMode === "table"
                                ? "bg-surface-selected text-foreground"
                                : "text-muted-foreground hover:text-foreground",
                        )}
                        onclick={() => (viewMode = "table")}
                    >
                        <TableProperties class="h-3.5 w-3.5" />
                    </button>
                    <button
                        type="button"
                        title="Card view"
                        aria-label="Card view"
                        aria-pressed={viewMode === "cards"}
                        class={cn(
                            "inline-flex h-full w-8 items-center justify-center rounded-[2px] transition-colors",
                            viewMode === "cards"
                                ? "bg-surface-selected text-foreground"
                                : "text-muted-foreground hover:text-foreground",
                        )}
                        onclick={() => (viewMode = "cards")}
                    >
                        <LayoutGrid class="h-3.5 w-3.5" />
                    </button>
                </div>
                <Button href={newFlagHref} size="default" class="h-9">
                    <Plus class="h-3.5 w-3.5" /> new flag
                </Button>
            </div>
        </div>
    </div>
</div>
