<script lang="ts">
    import { beforeNavigate, goto } from "$app/navigation";
    import { browser } from "$app/environment";
    import "../app.css";
    import { page } from "$app/state";
    import { expoOut } from "svelte/easing";
    import { LogOut } from "lucide-svelte";
    import { api } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import ClickSpark from "$lib/components/svelte-bits/click-spark.svelte";
    import type { LayoutProps } from "./$types";

    let { children, data }: LayoutProps = $props();

    const auth = $derived(data.auth);
    const backendUnavailable = $derived(data.backendUnavailable ?? false);
    const backendMessage = $derived(
        data.backendMessage ??
            "Backend server is not responding. Check that Flagcel is running, then retry.",
    );

    const prefersReducedMotion =
        browser && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    // Suppress intro animations on the very first paint (mirrors the `no-intro`
    // class); everything mounted after is allowed to transition.
    let ready = $state(false);

    beforeNavigate(() => {
        if (browser) document.documentElement.classList.remove("no-intro");
    });

    const dur = (ms: number) => (ready && !prefersReducedMotion ? ms : 0);

    // Horizontal collapse: animates width + horizontal margin + opacity so the
    // glass pill grows and shrinks smoothly as groups enter or leave. The margin
    // is folded in so no gap is left behind while a group collapses.
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

    const nav = [
        { href: "/", label: "flags" },
        { href: "/contexts", label: "contexts" },
        { href: "/api-keys", label: "keys", authEnabled: true },
        { href: "/docs", label: "api", external: true, icon: "↗" },
    ];

    function isActiveNavItem(href: string) {
        const pathname = page.url.pathname;
        if (href === "/") {
            return pathname === "/" || pathname.startsWith("/flags");
        }
        return pathname === href || pathname.startsWith(`${href}/`);
    }

    $effect(() => {
        ready = true;
    });

    async function logout() {
        await api.logout();
        await goto("/login");
    }

    function retry() {
        window.location.reload();
    }
</script>

<ClickSpark
    class="min-h-screen"
    sparkColor="var(--color-app-accent)"
    sparkSize={8}
    sparkRadius={18}
    sparkCount={8}
    duration={360}
>
    <div class="page-wrapper relative flex min-h-screen flex-col">
        <header
            class="fixed inset-x-0 top-0 z-50 flex justify-center px-4 pt-4 sm:pt-5"
        >
            <div
                class="glass-pill flex items-center rounded-sm px-4 py-2 sm:px-5"
            >
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
                {#if page.url.pathname !== "/login"}
                    <div
                        class="ml-3 flex items-center gap-3 sm:ml-4 sm:gap-4"
                        transition:collapseX
                    >
                        <span
                            class="h-3 w-px bg-[rgba(255,255,255,0.18)]"
                            aria-hidden="true"
                        ></span>
                        <nav
                            class="flex items-center gap-4 text-xs uppercase tracking-[0.12em]"
                        >
                            {#each nav as item (item.href)}
                                {@const active =
                                    !item.external && isActiveNavItem(item.href)}
                                {#if !item.authEnabled || auth?.auth_enabled}
                                    <a
                                        href={item.href}
                                        target={item.external
                                            ? "_blank"
                                            : undefined}
                                        rel={item.external
                                            ? "noopener"
                                            : undefined}
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
                {#if page.url.pathname !== "/login" && auth?.authenticated}
                    <div
                        class="ml-3 flex items-center gap-3 sm:ml-4 sm:gap-4"
                        transition:collapseX
                    >
                        <span
                            class="h-3 w-px bg-[rgba(255,255,255,0.18)]"
                            aria-hidden="true"
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

        <main
            class="mx-auto w-full max-w-5xl flex-1 px-6 pb-16 pt-32 sm:pt-36"
        >
            <div class="min-w-0">
                {#if backendUnavailable}
                    <section
                        class="mx-auto max-w-xl space-y-5 text-center"
                        aria-labelledby="backend-unavailable-title"
                    >
                        <div class="glass-panel motion-panel rounded-sm p-6 sm:p-8">
                            <p
                                class="mb-3 text-xs uppercase tracking-[0.18em] text-muted-foreground"
                            >
                                backend unavailable
                            </p>
                            <h1
                                id="backend-unavailable-title"
                                class="text-balance text-3xl font-normal leading-tight sm:text-4xl"
                            >
                                Flagcel cannot reach the backend.
                            </h1>
                            <p
                                class="mx-auto mt-4 max-w-md text-sm leading-6 text-[rgba(255,255,255,0.78)]"
                            >
                                {backendMessage}
                            </p>
                            <p
                                class="mx-auto mt-3 max-w-md text-sm leading-6 text-muted-foreground"
                            >
                                Start or restart the backend server, then retry this page.
                            </p>
                            <div class="mt-6 flex justify-center">
                                <Button onclick={retry}>Retry</Button>
                            </div>
                        </div>
                    </section>
                {:else}
                    {@render children()}
                {/if}
            </div>
        </main>

        <footer
            class="py-8 text-center text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
        >
            flagcel · self-hosted feature flags with cel
        </footer>
    </div>
</ClickSpark>
