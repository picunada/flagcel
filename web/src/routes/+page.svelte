<script lang="ts">
	import { onMount } from 'svelte';
	import { api, APIError, type Flag } from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import Badge from '$lib/components/ui/badge.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import SectionHeader from '$lib/components/ui/section-header.svelte';
	import { Plus, Search } from 'lucide-svelte';

	let flags = $state<Flag[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let query = $state('');

	const filtered = $derived(
		flags.filter((f) => f.key.toLowerCase().includes(query.toLowerCase()))
	);

	onMount(load);

	async function load() {
		loading = true;
		error = null;
		try {
			flags = await api.listFlags();
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to load flags';
		} finally {
			loading = false;
		}
	}
</script>

<section class="space-y-12">
	<header class="space-y-3">
		<p class="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">
			feature flags · cel-based targeting
		</p>
		<h1
			class="text-balance font-mono text-3xl font-normal leading-tight sm:text-4xl"
		>
			Roll out features<br />on your own terms.
		</h1>
		<p class="max-w-xl text-sm text-[rgba(255,255,255,0.78)] sm:text-base">
			Targeting rules written as CEL expressions, evaluated server-side, persisted in
			Postgres. No SaaS, no DSL to learn.
		</p>
	</header>

	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
		<SectionHeader>flags · {flags.length}</SectionHeader>
		<div class="flex items-center gap-2">
			<div class="relative">
				<Search
					class="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground"
				/>
				<Input bind:value={query} placeholder="search…" class="w-56 pl-8" />
			</div>
			<Button href="/flags/new">
				<Plus class="h-3.5 w-3.5" /> new flag
			</Button>
		</div>
	</div>

	{#if loading}
		<div class="grid gap-3 sm:grid-cols-2">
			{#each Array(4) as _, i (i)}
				<Card class="h-32 animate-pulse" />
			{/each}
		</div>
	{:else if error}
		<Card class="p-8 text-center">
			<p class="text-sm text-destructive">{error}</p>
			<Button variant="default" class="mt-4" onclick={load}>retry</Button>
		</Card>
	{:else if filtered.length === 0}
		<Card class="flex flex-col items-center gap-4 p-12 text-center">
			<p class="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
				{flags.length === 0 ? '[ no flags yet ]' : '[ no matches ]'}
			</p>
			<p class="max-w-sm text-sm text-[rgba(255,255,255,0.7)]">
				{flags.length === 0
					? 'Create your first flag to start rolling out features.'
					: 'Try a different search term.'}
			</p>
			{#if flags.length === 0}
				<Button href="/flags/new" class="mt-2">
					<Plus class="h-3.5 w-3.5" /> new flag
				</Button>
			{/if}
		</Card>
	{:else}
		<div class="grid gap-3 sm:grid-cols-2">
			{#each filtered as flag (flag.key)}
				<a href="/flags/{encodeURIComponent(flag.key)}" class="group block">
					<Card hoverable class="flex h-full flex-col gap-4 p-5">
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0 flex-1">
								<p
									class="truncate font-mono text-base font-medium group-hover:text-foreground"
								>
									{flag.key}
								</p>
								<p
									class="mt-1 font-mono text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
								>
									{flag.rules.length} rule{flag.rules.length === 1 ? '' : 's'} · default
									{String(flag.default_value)}
								</p>
							</div>
							{#if flag.enabled}
								<Badge variant="success" dot>on</Badge>
							{:else}
								<Badge variant="muted" dot>off</Badge>
							{/if}
						</div>
						{#if flag.rules.length > 0}
							<div class="mt-auto space-y-1.5">
								{#each flag.rules.slice(0, 2) as rule (rule.id)}
									<p
										class="truncate border-l-2 border-[rgba(255,255,255,0.08)] pl-2.5 font-mono text-xs text-muted-foreground"
									>
										{rule.expression}
									</p>
								{/each}
								{#if flag.rules.length > 2}
									<p
										class="pl-2.5 font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
									>
										+{flag.rules.length - 2} more
									</p>
								{/if}
							</div>
						{/if}
						<span
							class="mt-auto font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors group-hover:text-foreground"
						>
							open →
						</span>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</section>
