<script lang="ts">
    import "../app.css";
    import AppFooter from "$lib/components/layout/app-footer.svelte";
    import AppSidebar from "$lib/components/layout/app-sidebar.svelte";
    import BackendUnavailablePanel from "$lib/components/layout/backend-unavailable-panel.svelte";
    import ClickSpark from "$lib/components/svelte-bits/click-spark.svelte";
    import type { LayoutProps } from "./$types";

    let { children, data }: LayoutProps = $props();

    const auth = $derived(data.auth);
    const environments = $derived(data.environments ?? []);
    const selectedEnvironment = $derived(data.selectedEnvironment);
    const backendUnavailable = $derived(data.backendUnavailable ?? false);
    const backendMessage = $derived(
        data.backendMessage ??
            "Backend server is not responding. Check that Flagcel is running, then retry.",
    );
    const pathname = $derived(data.currentPathname);
    const currentSearch = $derived(data.currentSearch);
    const appShell = $derived(pathname !== "/login" && auth?.authenticated);
</script>

<ClickSpark
    class={appShell ? "min-h-dvh" : "min-h-screen"}
    sparkColor="var(--color-app-accent)"
    sparkSize={8}
    sparkRadius={18}
    sparkCount={8}
    duration={360}
>
    <div
        class={appShell
            ? "page-wrapper relative flex flex-col"
            : "page-wrapper relative flex min-h-screen flex-col"}
    >
        {#if appShell}
            <div class="h-dvh bg-sidebar p-2">
                <div
                    class="mx-auto flex h-full min-h-0 w-full flex-col lg:flex-row"
                >
                    <AppSidebar
                        {auth}
                        {environments}
                        {selectedEnvironment}
                        {pathname}
                        {currentSearch}
                    />

                    <main class="app-frame min-h-0 min-w-0 flex-1 bg-background">
                        <div
                            class={pathname.startsWith("/contexts")
                                ? "h-full overflow-hidden"
                                : "h-full overflow-y-auto px-5 py-8 sm:px-7 lg:px-8 lg:py-10"}
                        >
                            <div
                                class={pathname.startsWith("/contexts")
                                    ? "h-full min-w-0"
                                    : "min-w-0"}
                            >
                                {@render children()}
                            </div>
                        </div>
                    </main>
                </div>
            </div>
        {:else}
            <main class="mx-auto w-full flex-1 px-6 pb-16 pt-24 sm:pt-28">
                <div class="min-w-0">
                    {#if backendUnavailable}
                        <BackendUnavailablePanel message={backendMessage} />
                    {:else}
                        {@render children()}
                    {/if}
                </div>
            </main>

            <AppFooter />
        {/if}
    </div>
</ClickSpark>
