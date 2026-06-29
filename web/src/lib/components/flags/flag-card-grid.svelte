<script lang="ts">
    import type { Flag } from "$lib/api";
    import Badge from "$lib/components/ui/badge.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import { formatFlagValue } from "$lib/values";

    type Props = {
        flags: Flag[];
        flagHref: (key: string) => string;
        contextLabel: (flag: Flag) => string;
        formatUpdated: (value: string) => string;
    };

    let { flags, flagHref, contextLabel, formatUpdated }: Props = $props();
</script>

<div class="motion-list grid gap-3 p-3 sm:grid-cols-2">
    {#each flags as flag (flag.key)}
        <a href={flagHref(flag.key)} class="group block">
            <Card hoverable class="flex h-full flex-col gap-4 p-5">
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0 flex-1">
                        <p
                            class="truncate text-base font-medium group-hover:text-foreground"
                        >
                            {flag.key}
                        </p>
                        <p
                            class="mt-1 text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
                        >
                            {flag.type} · {flag.rules.length} rule{flag.rules
                                .length === 1
                                ? ""
                                : "s"} · default
                            {formatFlagValue(flag.default_value)}
                        </p>
                    </div>
                    {#if flag.enabled}
                        <Badge variant="success" dot>on</Badge>
                    {:else}
                        <Badge variant="muted" dot>off</Badge>
                    {/if}
                </div>
                <div
                    class="grid grid-cols-2 gap-2 border-y border-border py-3 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
                >
                    <div class="min-w-0">
                        <p>context</p>
                        <p
                            class="mt-1 truncate text-xs normal-case tracking-normal text-foreground/80"
                        >
                            {contextLabel(flag)}
                        </p>
                    </div>
                    <div class="text-right">
                        <p>changed</p>
                        <p
                            class="mt-1 text-xs normal-case tracking-normal text-foreground/80"
                        >
                            {formatUpdated(flag.updated_at)}
                        </p>
                    </div>
                </div>
                {#if flag.rules.length > 0}
                    <div class="mt-auto space-y-1.5">
                        {#each flag.rules.slice(0, 2) as rule (rule.id)}
                            <p
                                class="truncate border-l-2 border-border-muted pl-2.5 text-xs text-muted-foreground"
                            >
                                {rule.expression}
                            </p>
                        {/each}
                        {#if flag.rules.length > 2}
                            <p
                                class="pl-2.5 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
                            >
                                +{flag.rules.length - 2} more
                            </p>
                        {/if}
                    </div>
                {/if}
                <span
                    class="mt-auto text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors group-hover:text-foreground"
                >
                    open
                </span>
            </Card>
        </a>
    {/each}
</div>
