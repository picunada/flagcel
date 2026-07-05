<script lang="ts">
    import { goto } from "$app/navigation";
    import type { Environment } from "$lib/api";
    import * as Select from "$lib/components/ui/select";
    import { Select as SelectPrimitive } from "bits-ui";
    import { ChevronDown, Layers3, Settings2 } from "lucide-svelte";
    import { cn } from "$lib/utils";

    type Props = {
        environments?: Environment[];
        selectedEnvironment?: Environment;
        pathname: string;
        currentSearch: string;
        class?: string;
    };

    let {
        environments = [],
        selectedEnvironment,
        pathname,
        currentSearch,
        class: className,
    }: Props = $props();

    function targetHref(id: string) {
        const next = new URL(
            `${pathname}${currentSearch}`,
            "http://flagcel.local",
        );
        const path = pathname;
        if (!(
            path === "/" ||
            path.startsWith("/flags") ||
            path.startsWith("/api-keys")
        )) {
            next.pathname = "/";
        }
        next.searchParams.set("environment", id);
        return `${next.pathname}${next.search}`;
    }

    async function selectEnvironment(id: string) {
        await goto(targetHref(id));
    }
</script>

{#if selectedEnvironment && environments.length > 0}
    <Select.Root
        type="single"
        value={selectedEnvironment.id}
        onValueChange={selectEnvironment}
    >
        <SelectPrimitive.Trigger
            class={cn(
                "ios-corners-sm group inline-flex h-10 min-w-44 items-center gap-2 border border-app-accent-border bg-app-accent-surface pl-4 pr-3 text-left text-app-accent shadow-app-accent transition-colors hover:border-app-accent-border-strong hover:bg-app-accent-surface-hover focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-app-accent-ring-strong",
                className,
            )}
            title="Environment selector"
        >
            <Layers3 class="h-3.5 w-3.5 shrink-0" />
            <span
                class="hidden text-[0.62rem] uppercase tracking-[0.14em] text-app-accent-muted sm:inline"
            >
                env
            </span>
            <span
                class="ml-auto min-w-0 flex-1 truncate text-right font-mono text-xs text-foreground"
            >
                {selectedEnvironment.key}
            </span>
            <ChevronDown
                class="h-3.5 w-3.5 shrink-0 transition-transform group-data-[state=open]:rotate-180"
            />
        </SelectPrimitive.Trigger>

        <Select.Content
            align="end"
            class="w-[var(--bits-select-anchor-width)] min-w-44 "
        >
            <Select.Group>
                <Select.GroupHeading
                    class="px-2 py-1.5 text-[0.62rem] uppercase tracking-[0.14em] text-app-accent-muted"
                >
                    environment
                </Select.GroupHeading>
                {#each environments as env (env.id)}
                    <Select.Item value={env.id} label={env.key}>
                        <span class="block truncate font-mono">{env.key}</span>
                        <span
                            class="block truncate text-[0.65rem] text-muted-foreground"
                            >{env.name}</span
                        >
                    </Select.Item>
                {/each}
            </Select.Group>
            <a
                href="/environments"
                class="mt-1 flex items-center gap-2 border-t border-app-accent-border-muted px-2 py-2 text-[0.62rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-app-accent"
            >
                <Settings2 class="h-3.5 w-3.5 shrink-0" />
                manage environments
            </a>
        </Select.Content>
    </Select.Root>
{/if}
