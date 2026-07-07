<script lang="ts">
    import type { Flag } from "$lib/api";
    import Badge from "$lib/components/ui/badge.svelte";
    import Button from "$lib/components/ui/button.svelte";
    import { Progress } from "$lib/components/ui/progress";
    import Switch from "$lib/components/ui/switch/switch.svelte";
    import * as Tooltip from "$lib/components/ui/tooltip/index.js";
    import { cn } from "$lib/utils";
    import { formatFlagValue, valueBadgeVariant } from "$lib/values";
    import { flagEnabledTooltipLabel } from "./flag-table-tooltip";
    import { Info } from "lucide-svelte";

    type Props = {
        flags: Flag[];
        savingFlag: (key: string) => boolean;
        setFlagEnabled: (flag: Flag, enabled: boolean) => void | Promise<void>;
        flagHref: (key: string) => string;
        contextLabel: (flag: Flag) => string;
        formatUpdated: (value: string) => string;
    };

    const rolloutLayerTrackClasses = [
        "bg-rollout-stack-track-front",
        "bg-rollout-stack-layer-1",
        "bg-rollout-stack-layer-2",
        "bg-rollout-stack-layer-3",
    ];
    const rolloutLayerIndicatorClasses = [
        "bg-rollout-stack-fill-front",
        "bg-rollout-stack-fill-1",
        "bg-rollout-stack-fill-2",
        "bg-rollout-stack-fill-3",
    ];

    let {
        flags,
        savingFlag,
        setFlagEnabled,
        flagHref,
        contextLabel,
        formatUpdated,
    }: Props = $props();

    let rolloutAnchors = $state<Record<string, HTMLElement>>({});

    function rolloutAnchor(node: HTMLElement, key: string) {
        rolloutAnchors = { ...rolloutAnchors, [key]: node };

        return {
            destroy() {
                const { [key]: _removed, ...rest } = rolloutAnchors;
                rolloutAnchors = rest;
            },
        };
    }

    function rolloutPreviewLayers(flag: Flag) {
        return flag.rules.slice(0, 4).map((rule, index) => ({
            id: rule.id,
            label: `rule ${index + 1}`,
            percentage: rule.rollout.percentage,
            style: `top: calc(50% - 3px + ${index * 4}px); width: ${100 - index * 4}%; z-index: ${40 - index * 10}; scale: ${1 - index * 0.05}`,
            trackClass: rolloutLayerTrackClasses[index],
            indicatorClass: rolloutLayerIndicatorClasses[index],
        }));
    }
</script>

<div class="overflow-x-auto">
    <table class="w-full min-w-[760px] border-collapse text-left">
        <thead class="border-b border-border bg-surface-subtle">
            <tr
                class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
            >
                <th class="w-[30%] px-3 py-2 font-normal">key</th>
                <th class="px-3 py-2 font-normal">status</th>
                <th class="px-3 py-2 font-normal">serving</th>
                <th class="px-3 py-2 font-normal">context</th>
                <th class="px-3 py-2 text-right font-normal">rules</th>
                <th class="px-3 py-2 font-normal">type/default</th>
                <th class="px-3 py-2 font-normal">changed</th>
                <th class="px-3 py-2 text-right font-normal">action</th>
            </tr>
        </thead>
        <tbody class="divide-y divide-border">
            {#each flags as flag (flag.key)}
                <tr class="group transition-colors hover:bg-surface-muted">
                    <td class="px-3 py-2 align-middle">
                        <a
                            href={flagHref(flag.key)}
                            class="block max-w-[20rem] truncate text-sm text-foreground underline-offset-4 hover:underline"
                        >
                            {flag.key}
                        </a>
                    </td>
                    <td class="px-3 py-2 align-middle">
                        <Tooltip.Provider>
                            <Tooltip.Root delayDuration={200}>
                                <Tooltip.Trigger>
                                    {#snippet child({ props })}
                                        <Switch
                                            {...props}
                                            id={`flag-enabled-${flag.key}`}
                                            checked={flag.enabled}
                                            disabled={savingFlag(flag.key)}
                                            aria-label={`Toggle ${flag.key}`}
                                            onCheckedChange={(
                                                checked: boolean,
                                            ) => setFlagEnabled(flag, checked)}
                                        />
                                    {/snippet}
                                </Tooltip.Trigger>
                                <Tooltip.Content
                                    side="top"
                                    align="center"
                                    sideOffset={8}
                                >
                                    {flagEnabledTooltipLabel(
                                        flag,
                                        savingFlag(flag.key),
                                    )}
                                </Tooltip.Content>
                            </Tooltip.Root>
                        </Tooltip.Provider>
                    </td>
                    <td class="px-3 py-2 align-middle">
                        {#if !flag.enabled}
                            <div class="w-40">
                                <p class="text-xs text-muted-foreground">
                                    paused
                                </p>
                            </div>
                        {:else if flag.rules.length === 1}
                            <div
                                class="grid w-40 grid-cols-[1fr_1rem] items-center gap-2"
                            >
                                <Progress
                                    value={flag.rules[0].rollout.percentage}
                                    max={100}
                                    class="h-1.5"
                                />
                                <span aria-hidden="true"></span>
                            </div>
                        {:else if flag.rules.length > 1}
                            <Tooltip.Provider>
                                <Tooltip.Root delayDuration={200}>
                                    <div
                                        class="grid w-40 grid-cols-[1fr_1rem] items-center gap-2"
                                    >
                                        <div
                                            use:rolloutAnchor={flag.key}
                                            class="progress-stack-fade relative h-6 min-w-0 rounded-sm"
                                        >
                                            {#each rolloutPreviewLayers(flag) as layer (layer.id)}
                                                <Progress
                                                    value={layer.percentage}
                                                    max={100}
                                                    aria-label={`${layer.label} rollout ${layer.percentage}%`}
                                                    class={cn(
                                                        "absolute left-1/2 h-1.5 -translate-x-1/2 rounded-full",
                                                        layer.trackClass,
                                                    )}
                                                    indicatorClass={layer.indicatorClass}
                                                    style={layer.style}
                                                />
                                            {/each}
                                        </div>
                                        <Tooltip.Trigger>
                                            {#snippet child({ props })}
                                                <button
                                                    {...props}
                                                    type="button"
                                                    class="motion-press inline-flex size-4 shrink-0 items-center justify-center text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                                                    aria-label={`Show rollout details for ${flag.key}`}
                                                >
                                                    <Info class="size-3.5" />
                                                </button>
                                            {/snippet}
                                        </Tooltip.Trigger>
                                    </div>
                                    <Tooltip.Content
                                        side="top"
                                        align="center"
                                        sideOffset={10}
                                        customAnchor={rolloutAnchors[
                                            flag.key
                                        ] ?? null}
                                        class="max-w-none p-0"
                                    >
                                        <div class="w-80">
                                            <div
                                                class="border-b border-border px-3 py-2"
                                            >
                                                <p
                                                    class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
                                                >
                                                    rollout distribution
                                                </p>
                                                <p
                                                    class="mt-1 text-xs text-foreground"
                                                >
                                                    {flag.rules.length} rules
                                                    serving this flag
                                                </p>
                                            </div>
                                            <div
                                                class="max-h-72 divide-y divide-border overflow-y-auto"
                                            >
                                                {#each flag.rules as rule, i (rule.id)}
                                                    <div
                                                        class="space-y-2 px-3 py-2.5"
                                                    >
                                                        <div
                                                            class="flex items-center justify-between gap-3"
                                                        >
                                                            <p
                                                                class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
                                                            >
                                                                rule {i + 1}
                                                            </p>
                                                            <p
                                                                class="font-mono text-xs text-app-accent"
                                                            >
                                                                {rule.rollout
                                                                    .percentage}%
                                                            </p>
                                                        </div>
                                                        <Progress
                                                            value={rule.rollout
                                                                .percentage}
                                                            max={100}
                                                            class="h-1.5"
                                                        />
                                                        <p
                                                            class="line-clamp-2 font-mono text-xs leading-5 text-foreground-soft"
                                                        >
                                                            {rule.expression}
                                                        </p>
                                                        {#if rule.rollout.bucket_by}
                                                            <p
                                                                class="truncate text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
                                                            >
                                                                bucket by
                                                                <span
                                                                    class="font-mono normal-case tracking-normal text-foreground"
                                                                    >{rule
                                                                        .rollout
                                                                        .bucket_by}</span
                                                                >
                                                            </p>
                                                        {/if}
                                                    </div>
                                                {/each}
                                            </div>
                                        </div>
                                    </Tooltip.Content>
                                </Tooltip.Root>
                            </Tooltip.Provider>
                        {/if}
                    </td>
                    <td
                        class="px-3 py-2 align-middle text-xs text-muted-foreground"
                    >
                        <span class="block max-w-[12rem] truncate"
                            >{contextLabel(flag)}</span
                        >
                    </td>
                    <td
                        class="px-3 py-2 text-right align-middle text-sm tabular-nums"
                    >
                        {flag.rules.length}
                    </td>
                    <td class="px-3 py-2 align-middle">
                        <Badge variant={valueBadgeVariant(flag.default_value)}>
                            {flag.type}: {formatFlagValue(flag.default_value)}
                        </Badge>
                    </td>
                    <td
                        class="px-3 py-2 align-middle text-xs text-muted-foreground"
                    >
                        {formatUpdated(flag.updated_at)}
                    </td>
                    <td class="px-3 py-2 text-right align-middle">
                        <Button
                            href={flagHref(flag.key)}
                            variant="ghost"
                            size="sm"
                        >
                            open
                        </Button>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</div>
