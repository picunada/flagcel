<script lang="ts">
    import { goto, invalidateAll } from "$app/navigation";
    import { untrack } from "svelte";
    import {
        api,
        APIError,
        type ContextSchema,
        type Environment,
        type Flag,
        type FlagUsage,
    } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import FlagCardGrid from "$lib/components/flags/flag-card-grid.svelte";
    import FlagDashboardToolbar from "$lib/components/flags/flag-dashboard-toolbar.svelte";
    import FlagListEmptyState from "$lib/components/flags/flag-list-empty-state.svelte";
    import FlagTable from "$lib/components/flags/flag-table.svelte";
    import UsageAnalyticsCharts from "$lib/components/flags/usage-analytics-charts.svelte";
    import UsageRangeControl from "$lib/components/flags/usage-range-control.svelte";
    import { normalizeUsage, type UsageRangeHours } from "$lib/usage-analytics";
    import { formatFlagValue } from "$lib/values";
    import AppBreadcrumbs from "$lib/components/ui/app-breadcrumbs.svelte";
    import PageHeader from "$lib/components/ui/page-header.svelte";
    import type { PageProps } from "./$types";

    type StatusFilter = "all" | "on" | "off";
    type SortKey =
        "recent" | "rules" | "default" | "status" | "key" | "context";
    type ViewMode = "table" | "cards";

    const contextFilterBaseOptions = [
        { value: "all", label: "all contexts" },
        { value: "none", label: "no context" },
    ];
    const sortOptions = [
        { value: "recent", label: "recently changed" },
        { value: "rules", label: "rule count" },
        { value: "default", label: "default value" },
        { value: "status", label: "status" },
        { value: "context", label: "context" },
        { value: "key", label: "key" },
    ];
    let { data }: PageProps = $props();

    let flags = $state<Flag[]>([]);
    const selectedEnvironment = $derived<Environment | undefined>(
        data.selectedEnvironment,
    );
    const contexts = $derived<ContextSchema[]>(data.contexts);
    const contextById = $derived.by(
        () => new Map(contexts.map((ctx) => [ctx.id, ctx])),
    );
    const contextFilterOptions = $derived([
        ...contextFilterBaseOptions,
        ...contexts.map((ctx) => ({ value: ctx.id, label: ctx.name })),
    ]);

    let error = $state<string | null>(null);
    let environmentUsage = $state<FlagUsage>(
        untrack(() => normalizeUsage(data.environmentUsage)),
    );
    let usageRange = $state<UsageRangeHours>(untrack(() => data.usageRange));
    let query = $state("");
    let statusFilter = $state<StatusFilter>("all");
    let contextFilter = $state("all");
    let sortKey = $state<SortKey>("recent");
    let viewMode = $state<ViewMode>("table");

    $effect(() => {
        flags = data.flags.map(copyFlag);
        environmentUsage = normalizeUsage(data.environmentUsage);
        usageRange = data.usageRange;
    });

    const filtered = $derived.by(() => {
        const needle = query.trim().toLowerCase();
        return flags
            .filter((flag) => {
                const statusMatches =
                    statusFilter === "all" ||
                    (statusFilter === "on" && flag.enabled) ||
                    (statusFilter === "off" && !flag.enabled);
                const contextMatches =
                    contextFilter === "all" ||
                    (contextFilter === "none" && !flag.context_id) ||
                    flag.context_id === contextFilter;
                const contextName = contextLabel(flag).toLowerCase();
                const searchMatches =
                    needle === "" ||
                    flag.key.toLowerCase().includes(needle) ||
                    contextName.includes(needle) ||
                    flag.rules.some((rule) =>
                        rule.expression.toLowerCase().includes(needle),
                    );

                return statusMatches && contextMatches && searchMatches;
            })
            .sort((a, b) => compareFlags(a, b));
    });

    const activeFilters = $derived(
        Number(statusFilter !== "all") +
            Number(contextFilter !== "all") +
            Number(query.trim() !== ""),
    );

    function compareFlags(a: Flag, b: Flag) {
        if (sortKey === "recent")
            return timestamp(b) - timestamp(a) || a.key.localeCompare(b.key);
        if (sortKey === "rules")
            return (
                b.rules.length - a.rules.length || a.key.localeCompare(b.key)
            );
        if (sortKey === "default") {
            return (
                formatFlagValue(a.default_value).localeCompare(
                    formatFlagValue(b.default_value),
                ) || a.key.localeCompare(b.key)
            );
        }
        if (sortKey === "status")
            return (
                Number(b.enabled) - Number(a.enabled) ||
                a.key.localeCompare(b.key)
            );
        if (sortKey === "context") {
            return (
                contextLabel(a).localeCompare(contextLabel(b)) ||
                a.key.localeCompare(b.key)
            );
        }
        return a.key.localeCompare(b.key);
    }

    function copyFlag(flag: Flag): Flag {
        return {
            ...flag,
            rules: flag.rules.map((rule) => ({
                ...rule,
                rollout: { ...rule.rollout },
            })),
        };
    }

    function contextLabel(flag: Flag) {
        if (!flag.context_id) return "no context";
        return contextById.get(flag.context_id)?.name ?? "missing context";
    }

    function timestamp(flag: Flag) {
        return Date.parse(flag.updated_at) || 0;
    }

    function formatUpdated(value: string) {
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return "unknown";
        return new Intl.DateTimeFormat(undefined, {
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "2-digit",
        }).format(date);
    }

    function resetFilters() {
        query = "";
        statusFilter = "all";
        contextFilter = "all";
    }

    function flagHref(key: string) {
        const query = selectedEnvironment
            ? `?environment=${encodeURIComponent(selectedEnvironment.id)}`
            : "";
        return `/flags/${encodeURIComponent(key)}${query}`;
    }

    function newFlagHref() {
        const query = selectedEnvironment
            ? `?environment=${encodeURIComponent(selectedEnvironment.id)}`
            : "";
        return `/flags/new${query}`;
    }

    function setUsageRange(next: UsageRangeHours) {
        const nextUrl = new URL(window.location.href);
        nextUrl.searchParams.set("usageRange", String(next));
        goto(`${nextUrl.pathname}${nextUrl.search}`, {
            keepFocus: true,
            noScroll: true,
        });
    }

    async function refresh() {
        error = null;
        try {
            await invalidateAll();
        } catch (e) {
            error = e instanceof APIError ? e.message : "Failed to load flags";
        }
    }

    // Flag enabled state
    let savingFlagKeys = $state<string[]>([]);

    function savingFlag(key: string) {
        return savingFlagKeys.includes(key);
    }

    async function setFlagEnabled(flag: Flag, enabled: boolean) {
        const environment = selectedEnvironment;
        if (!environment || savingFlag(flag.key) || flag.enabled == enabled)
            return;

        const previous = flag.enabled;
        flags = flags.map((candidate) =>
            candidate.key === flag.key ? { ...candidate, enabled } : candidate,
        );
        savingFlagKeys.push(flag.key);
        error = null;

        try {
            await api.createFlag(environment.id, {
                key: flag.key,
                description: flag.description,
                type: flag.type,
                enabled,
                default_value: flag.default_value,
                context_id: flag.context_id,
                rules: flag.rules.map((rule) => ({
                    id: rule.id,
                    description: rule.description,
                    expression: rule.expression,
                    rollout: rule.rollout,
                    value: rule.value,
                })),
            });
        } catch (e) {
            flags = flags.map((candidate) =>
                candidate.key === flag.key
                    ? { ...candidate, enabled: previous }
                    : candidate,
            );
            error = e instanceof APIError ? e.message : "Failed to save flag";
        } finally {
            savingFlagKeys = savingFlagKeys.filter((key) => key !== flag.key);
        }
    }
</script>

<section class="space-y-4">
    <div class="space-y-3">
        <AppBreadcrumbs items={[{ label: "flags" }]} />
        <!-- <PageHeader eyebrow="feature flags · targeting rules" /> -->
    </div>

    {#if flags.length > 0}
        <div class="space-y-3">
            <div class="flex items-center justify-between gap-3">
                <div>
                    <p
                        class="text-xs uppercase tracking-[0.14em] text-muted-foreground"
                    >
                        usage analytics
                    </p>
                </div>
                <UsageRangeControl value={usageRange} onchange={setUsageRange} />
            </div>
            <UsageAnalyticsCharts
                usage={environmentUsage}
                range={usageRange}
                variant="dashboard"
            />
        </div>
    {/if}

    <Card class="motion-panel overflow-hidden">
        <FlagDashboardToolbar
            bind:query
            bind:statusFilter
            bind:contextFilter
            bind:sortKey
            bind:viewMode
            {contextFilterOptions}
            {sortOptions}
            shownCount={filtered.length}
            newFlagHref={newFlagHref()}
        />

        {#if error}
            <div class="p-8 text-center">
                <p class="text-sm text-destructive">{error}</p>
                <Button variant="default" class="mt-4" onclick={refresh}
                    >retry</Button
                >
            </div>
        {:else if flags.length === 0}
            <FlagListEmptyState
                kind="empty"
                newFlagHref={newFlagHref()}
                onreset={resetFilters}
            />
        {:else if filtered.length === 0}
            <FlagListEmptyState
                kind="filtered"
                {activeFilters}
                newFlagHref={newFlagHref()}
                onreset={resetFilters}
            />
        {:else if viewMode === "cards"}
            <FlagCardGrid
                flags={filtered}
                {flagHref}
                {contextLabel}
                {formatUpdated}
            />
        {:else}
            <FlagTable
                flags={filtered}
                {savingFlag}
                {setFlagEnabled}
                {flagHref}
                {contextLabel}
                {formatUpdated}
            />
        {/if}
    </Card>
</section>
