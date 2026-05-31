<script lang="ts">
    import { goto } from "$app/navigation";
    import { api, APIError } from "$lib/api";
    import Button from "$lib/components/ui/button.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import Input from "$lib/components/ui/input.svelte";
    import { LogIn } from "lucide-svelte";
    import type { PageProps } from "./$types";

    let { data }: PageProps = $props();

    let email = $state("");
    let password = $state("");
    let submitting = $state(false);
    let error = $state<string | null>(null);

    function login() {
        window.location.href = "/auth/login";
    }

    async function passwordLogin() {
        if (!email.trim() || !password) return;
        submitting = true;
        error = null;
        try {
            await api.passwordLogin(email, password);
            await goto("/");
        } catch (e) {
            error = e instanceof APIError ? e.message : "Failed to sign in";
        } finally {
            submitting = false;
        }
    }
</script>

<section class="mx-auto max-w-md space-y-6">
    <header class="space-y-3 text-center">
        <p class="text-xs uppercase tracking-[0.18em] text-muted-foreground">
            flagcel admin
        </p>
        <h1 class="text-3xl font-normal leading-tight">Sign in</h1>
    </header>

    <Card class="motion-panel space-y-5 p-6 text-center">
        {#if data.auth?.mode === "oidc"}
            <p class="text-sm text-[rgba(255,255,255,0.74)]">
                Use your configured SSO provider to access flags, contexts, and API keys.
            </p>
            <Button variant="solid" class="w-full" onclick={login}>
                <LogIn class="h-3.5 w-3.5" /> continue with sso
            </Button>
        {:else}
            <form class="space-y-3 text-left" onsubmit={(e) => { e.preventDefault(); passwordLogin(); }}>
                <Input bind:value={email} type="email" autocomplete="username" placeholder="email" />
                <Input
                    bind:value={password}
                    type="password"
                    autocomplete="current-password"
                    placeholder="password"
                />
                <Button
                    variant="solid"
                    class="w-full"
                    disabled={submitting || !email.trim() || !password}
                >
                    <LogIn class="h-3.5 w-3.5" /> sign in
                </Button>
            </form>
        {/if}
        {#if error}
            <p class="text-sm text-destructive">{error}</p>
        {/if}
    </Card>
</section>
