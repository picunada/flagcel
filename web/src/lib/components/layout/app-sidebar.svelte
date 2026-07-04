<script lang="ts">
    import { beforeNavigate, goto } from "$app/navigation";
    import {
        BookOpen,
        ChevronDown,
        CircleHelp,
        Flag,
        Home,
        KeyRound,
        Layers3,
        LogOut,
        Menu,
        Settings,
        Users,
        X,
    } from "lucide-svelte";
    import { api, type AuthMe, type Environment } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import EnvironmentSelector from "$lib/components/environment-selector.svelte";
    import { cn } from "$lib/utils";

    type NavItem = {
        href: string;
        match: string;
        label: string;
        icon: typeof Home;
        external?: boolean;
        authEnabled?: boolean;
        badge?: string;
    };

    let {
        auth,
        environments = [],
        selectedEnvironment,
        pathname,
        currentSearch,
    }: {
        auth?: AuthMe;
        environments?: Environment[];
        selectedEnvironment?: Environment;
        pathname: string;
        currentSearch: string;
    } = $props();

    let mobileOpen = $state(false);

    const accountName = $derived(
        auth?.user?.name?.trim() || auth?.user?.email || "Flagcel admin",
    );
    const accountEmail = $derived(auth?.user?.email || "local admin");
    const environmentQuery = $derived(
        selectedEnvironment
            ? `?environment=${encodeURIComponent(selectedEnvironment.id)}`
            : "",
    );
    const visibleEnvironments = $derived(Math.max(environments.length, 0));

    const nav = $derived<NavItem[]>([
        {
            href: `/${environmentQuery}`,
            match: "/",
            label: "Flags",
            icon: Flag,
        },
        {
            href: "/contexts",
            match: "/contexts",
            label: "Contexts",
            icon: Users,
        },
        {
            href: "/environments",
            match: "/environments",
            label: "Environments",
            icon: Layers3,
            badge: visibleEnvironments ? String(visibleEnvironments) : undefined,
        },
        {
            href: `/api-keys${environmentQuery}`,
            match: "/api-keys",
            label: "API keys",
            icon: KeyRound,
            authEnabled: true,
        },
        {
            href: "/docs",
            match: "/docs",
            label: "API docs",
            icon: BookOpen,
            external: true,
        },
    ]);

    beforeNavigate(() => {
        mobileOpen = false;
    });

    function isActiveNavItem(match: string) {
        if (match === "/") {
            return pathname === "/" || pathname.startsWith("/flags");
        }
        return pathname === match || pathname.startsWith(`${match}/`);
    }

    async function logout() {
        mobileOpen = false;
        await api.logout();
        await goto("/login");
    }

    function closeMobile() {
        mobileOpen = false;
    }

    function openMobile() {
        mobileOpen = true;
    }
</script>

{#snippet sidebarContent()}
    <div class="flex h-full min-h-0 flex-col overflow-y-auto bg-sidebar">
        <div class="space-y-7 px-5 pb-5 pt-5">
            <div class="flex items-center gap-3">
                <a
                    href="/"
                    class="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-[0.55rem] bg-surface-active text-foreground transition-colors hover:bg-surface-selected"
                    aria-label="Flagcel home"
                >
                    <Flag class="h-5 w-5" strokeWidth={2.5} />
                    <span
                        class="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border border-sidebar bg-app-accent"
                        aria-hidden="true"
                    ></span>
                </a>
                <div class="min-w-0 flex-1">
                    <div class="truncate text-lg font-semibold leading-5 text-foreground">
                        Flagcel
                    </div>
                    <div class="mt-1 truncate text-xs text-muted-foreground">
                        {accountEmail}
                    </div>
                </div>
                <Button
                    variant="ghost"
                    size="icon"
                    class="hidden h-8 w-8 shrink-0 text-muted-foreground lg:inline-flex"
                    title="Workspace menu"
                >
                    <ChevronDown class="h-4 w-4" />
                </Button>
                <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0 text-muted-foreground lg:hidden"
                    title="Close menu"
                    onclick={closeMobile}
                >
                    <X class="h-4 w-4" />
                </Button>
            </div>

            {#if selectedEnvironment && environments.length > 0}
                <div class="space-y-3">
                    <div class="flex items-center justify-between gap-3">
                        <p class="text-xs font-medium text-muted-foreground">
                            Environment
                        </p>
                        <span
                            class="rounded-sm bg-app-accent-surface px-2 py-0.5 text-[0.65rem] font-medium text-app-accent"
                        >
                            active
                        </span>
                    </div>
                    <EnvironmentSelector
                        {environments}
                        {selectedEnvironment}
                        {pathname}
                        {currentSearch}
                        class="w-full"
                    />
                </div>
            {/if}
        </div>

        <div class="h-px bg-border-divider"></div>

        <nav class="space-y-1 px-4 py-5" aria-label="Primary navigation">
            {#each nav as item (item.href)}
                {@const Icon = item.icon}
                {@const active = !item.external && isActiveNavItem(item.match)}
                {#if !item.authEnabled || auth?.auth_enabled}
                    <a
                        href={item.href}
                        target={item.external ? "_blank" : undefined}
                        rel={item.external ? "noopener" : undefined}
                        class={cn(
                            "group flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors",
                            active
                                ? "bg-surface-selected text-foreground"
                                : "text-muted-foreground hover:bg-surface-hover hover:text-foreground",
                        )}
                    >
                        <Icon class="h-4 w-4 shrink-0" />
                        <span class="min-w-0 flex-1 truncate">{item.label}</span>
                        {#if item.badge}
                            <span
                                class="rounded-sm bg-app-accent-surface px-1.5 py-0.5 text-[0.65rem] leading-none text-app-accent"
                            >
                                {item.badge}
                            </span>
                        {/if}
                    </a>
                {/if}
            {/each}
        </nav>

        <div class="mt-auto space-y-1 px-4 pb-5">
            <div class="h-px bg-border-divider"></div>
            <a
                href="/environments"
                class="mt-4 flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-surface-hover hover:text-foreground"
            >
                <Settings class="h-4 w-4 shrink-0" />
                <span class="min-w-0 flex-1 truncate">Settings</span>
            </a>
            <a
                href="/docs"
                target="_blank"
                rel="noopener"
                class="flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-surface-hover hover:text-foreground"
            >
                <CircleHelp class="h-4 w-4 shrink-0" />
                <span class="min-w-0 flex-1 truncate">Help center</span>
            </a>
            {#if auth?.authenticated}
                <button
                    type="button"
                    class="flex h-10 w-full items-center gap-3 rounded-md px-3 text-left text-sm font-medium text-muted-foreground transition-colors hover:bg-surface-hover hover:text-foreground"
                    onclick={logout}
                >
                    <LogOut class="h-4 w-4 shrink-0" />
                    <span class="min-w-0 flex-1 truncate">Sign out</span>
                </button>
            {/if}
        </div>
    </div>
{/snippet}

<div class="w-full lg:hidden">
    <div
        class="flex h-16 items-center gap-3 border-b border-border-divider bg-sidebar px-4"
    >
        <Button
            variant="ghost"
            size="icon"
            class="h-9 w-9 text-muted-foreground"
            title="Open menu"
            onclick={openMobile}
        >
            <Menu class="h-4 w-4" />
        </Button>
        <a href="/" class="flex min-w-0 flex-1 items-center gap-3">
            <span
                class="relative flex h-8 w-8 shrink-0 items-center justify-center rounded-[0.45rem] bg-surface-active text-foreground"
            >
                <Flag class="h-4 w-4" strokeWidth={2.5} />
                <span
                    class="absolute -bottom-0.5 -right-0.5 h-2 w-2 rounded-full border border-sidebar bg-app-accent"
                    aria-hidden="true"
                ></span>
            </span>
            <span class="truncate text-base font-semibold">Flagcel</span>
        </a>
        <span class="truncate text-xs text-muted-foreground">{accountName}</span>
    </div>

    {#if mobileOpen}
        <div class="fixed inset-0 z-50 bg-black/60" role="presentation" onclick={closeMobile}></div>
        <aside
            class="fixed bottom-0 left-0 top-0 z-50 w-[min(20rem,calc(100vw-2rem))] border-r border-border-divider shadow-drawer"
            aria-label="Mobile navigation"
        >
            {@render sidebarContent()}
        </aside>
    {/if}
</div>

<aside
    class="hidden h-full w-72 shrink-0 overflow-hidden lg:block"
    aria-label="Application navigation"
>
    {@render sidebarContent()}
</aside>
