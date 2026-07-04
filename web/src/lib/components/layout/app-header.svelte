<script lang="ts">
    import { browser } from "$app/environment";
    import { goto, beforeNavigate } from "$app/navigation";
    import { expoOut } from "svelte/easing";
    import { LogOut } from "lucide-svelte";
    import { api, type AuthMe, type Environment } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import EnvironmentSelector from "$lib/components/environment-selector.svelte";

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

    const prefersReducedMotion =
        browser &&
        window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    let ready = $state(false);

    beforeNavigate(() => {
        if (browser) document.documentElement.classList.remove("no-intro");
    });

    const dur = (ms: number) => (ready && !prefersReducedMotion ? ms : 0);

    function collapseX(node: HTMLElement, { duration = 260 } = {}) {
        const { width } = node.getBoundingClientRect();
        const { marginLeft, marginRight } = getComputedStyle(node);
        const ml = parseFloat(marginLeft);
        const mr = parseFloat(marginRight);
        return {
            duration: dur(duration),
            easing: expoOut,
            css: (t: number) =>
                `overflow:hidden;white-space:nowrap;opacity:${t};width:${t * width}px;margin-left:${t * ml}px;margin-right:${t * mr}px`,
        };
    }

    const environmentQuery = $derived(
        selectedEnvironment
            ? `?environment=${encodeURIComponent(selectedEnvironment.id)}`
            : "",
    );

    const nav = $derived([
        { href: `/${environmentQuery}`, match: "/", label: "flags" },
        { href: "/contexts", match: "/contexts", label: "contexts" },
        { href: "/environments", match: "/environments", label: "envs" },
        {
            href: `/api-keys${environmentQuery}`,
            match: "/api-keys",
            label: "keys",
            authEnabled: true,
        },
        {
            href: "/docs",
            match: "/docs",
            label: "api",
            external: true,
            icon: "↗",
        },
    ]);

    function isActiveNavItem(match: string) {
        if (match === "/") {
            return pathname === "/" || pathname.startsWith("/flags");
        }
        return pathname === match || pathname.startsWith(`${match}/`);
    }

    $effect(() => {
        ready = true;
    });

    async function logout() {
        await api.logout();
        await goto("/login");
    }
</script>

<header class="fixed inset-x-0 top-0 z-50 flex justify-center px-4 pt-4 sm:pt-5">
    <div class="glass-pill flex items-center rounded-sm px-4 py-2 sm:px-5">
        <a
            href="/"
            class="flex h-6 items-center gap-2 font-mono text-xs font-medium uppercase tracking-[0.12em]"
        >
            <span
                class="inline-flex h-4 w-4 items-center justify-center text-success"
                aria-hidden="true"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="h-3.5 w-3.5"
                >
                    <path
                        d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"
                    />
                    <line x1="4" y1="22" x2="4" y2="15" />
                </svg>
            </span>
            <span>flagcel</span>
        </a>
        {#if pathname !== "/login"}
            <div
                class="ml-3 flex items-center gap-3 sm:ml-4 sm:gap-4"
                transition:collapseX
            >
                <span class="h-3 w-px bg-border-divider" aria-hidden="true"
                ></span>
                <nav
                    class="flex items-center gap-4 text-xs uppercase tracking-[0.12em]"
                >
                    {#each nav as item (item.href)}
                        {@const active =
                            !item.external && isActiveNavItem(item.match)}
                        {#if !item.authEnabled || auth?.auth_enabled}
                            <a
                                href={item.href}
                                target={item.external ? "_blank" : undefined}
                                rel={item.external ? "noopener" : undefined}
                                class="inline-flex items-baseline gap-1 transition-colors {active
                                    ? 'text-foreground'
                                    : 'text-muted-foreground hover:text-foreground'}"
                            >
                                {item.label}{#if item.icon}<span
                                        class="text-muted-foreground text-[0.85em]"
                                        >{item.icon}</span
                                    >{/if}
                            </a>
                        {/if}
                    {/each}
                </nav>
            </div>
        {/if}
        {#if pathname !== "/login" && auth?.authenticated}
            <div
                class="ml-3 flex items-center gap-3 sm:ml-4"
                transition:collapseX
            >
                <span class="h-3 w-px bg-border-divider" aria-hidden="true"
                ></span>
                <EnvironmentSelector
                    {environments}
                    {selectedEnvironment}
                    {pathname}
                    {currentSearch}
                />
            </div>
        {/if}
        {#if pathname !== "/login" && auth?.authenticated}
            <div
                class="ml-3 flex items-center gap-3 sm:ml-4 sm:gap-4"
                transition:collapseX
            >
                <span class="h-3 w-px bg-border-divider" aria-hidden="true"
                ></span>
                <Button
                    variant="ghost"
                    size="sm"
                    class="-mr-2 h-6 px-2"
                    title="Sign out"
                    onclick={logout}
                >
                    <LogOut class="h-3.5 w-3.5" />
                </Button>
            </div>
        {/if}
    </div>
</header>
