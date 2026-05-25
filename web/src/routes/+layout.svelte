<script lang="ts">
    import "../app.css";
    import { page } from "$app/state";
    import type { Snippet } from "svelte";
    import ClickSpark from "$lib/components/svelte-bits/click-spark.svelte";

    let { children }: { children: Snippet } = $props();

    const nav = [
        { href: "/", label: "flags" },
        { href: "/contexts", label: "contexts" },
        { href: "/docs", label: "api", external: true, icon: "↗" },
    ];

    function isActiveNavItem(href: string) {
        const pathname = page.url.pathname;
        if (href === "/") {
            return pathname === "/" || pathname.startsWith("/flags");
        }
        return pathname === href || pathname.startsWith(`${href}/`);
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
                class="glass-pill flex items-center gap-3 rounded-sm px-4 py-2 sm:gap-4 sm:px-5"
            >
                <a
                    href="/"
                    class="flex items-center gap-2 font-mono text-xs font-medium uppercase tracking-[0.12em]"
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
                <span
                    class="h-3 w-px bg-[rgba(255,255,255,0.18)]"
                    aria-hidden="true"
                ></span>
                <nav
                    class="flex items-center gap-4 font-mono text-xs uppercase tracking-[0.12em]"
                >
                    {#each nav as item (item.href)}
                        {@const active =
                            !item.external && isActiveNavItem(item.href)}
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
                    {/each}
                </nav>
            </div>
        </header>

        <main
            class="motion-page mx-auto w-full max-w-5xl flex-1 px-6 pb-16 pt-32 sm:pt-36"
        >
            {@render children()}
        </main>

        <footer
            class="py-8 text-center font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground"
        >
            flagcel · self-hosted feature flags with cel
        </footer>
    </div>
</ClickSpark>
