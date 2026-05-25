<script lang="ts">
	import { onMount } from 'svelte';
	import { api, APIError, type ContextSchema } from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import SectionHeader from '$lib/components/ui/section-header.svelte';
	import { Plus } from 'lucide-svelte';

	let contexts = $state<ContextSchema[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(load);

	async function load() {
		loading = true;
		error = null;
		try {
			contexts = await api.listContexts();
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to load contexts';
		} finally {
			loading = false;
		}
	}
</script>

<section class="space-y-12">
	<header class="space-y-3">
		<p class="font-mono text-xs uppercase tracking-[0.18em] text-muted-foreground">
			evaluation contexts · field shapes
		</p>
		<h1 class="text-balance font-mono text-3xl font-normal leading-tight sm:text-4xl">
			Describe what your<br />rules can reach.
		</h1>
		<p class="max-w-xl text-sm text-[rgba(255,255,255,0.78)] sm:text-base">
			Name the fields each flag's expressions can use. Selected at flag level — drives
			autocomplete in the rule editor.
		</p>
	</header>

	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
		<SectionHeader>contexts · {contexts.length}</SectionHeader>
		<Button href="/contexts/new">
			<Plus class="h-3.5 w-3.5" /> new context
		</Button>
	</div>

	{#if loading}
		<div class="grid gap-3 sm:grid-cols-2">
			{#each Array(2) as _, i (i)}
				<Card class="h-32 animate-pulse" />
			{/each}
		</div>
	{:else if error}
		<Card class="motion-panel p-8 text-center">
			<p class="text-sm text-destructive">{error}</p>
			<Button class="mt-4" onclick={load}>retry</Button>
		</Card>
	{:else if contexts.length === 0}
		<Card class="motion-panel flex flex-col items-center gap-4 p-12 text-center">
			<p class="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
				[ no contexts yet ]
			</p>
			<p class="max-w-sm text-sm text-[rgba(255,255,255,0.7)]">
				Define one to enable autocomplete on rules.
			</p>
			<Button href="/contexts/new" class="mt-2">
				<Plus class="h-3.5 w-3.5" /> new context
			</Button>
		</Card>
	{:else}
		<div class="motion-list grid gap-3 sm:grid-cols-2">
			{#each contexts as ctx (ctx.id)}
				<a href="/contexts/{encodeURIComponent(ctx.id)}" class="group block">
					<Card hoverable class="flex h-full flex-col gap-3 p-5">
						<p class="truncate font-mono text-base font-medium">{ctx.name}</p>
						{#if ctx.description}
							<p class="line-clamp-2 text-sm text-[rgba(255,255,255,0.7)]">
								{ctx.description}
							</p>
						{/if}
						<p
							class="mt-auto font-mono text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
						>
							{ctx.fields.length} field{ctx.fields.length === 1 ? '' : 's'}
						</p>
						{#if ctx.fields.length > 0}
							<div class="space-y-1">
								{#each ctx.fields.slice(0, 3) as f (f.path)}
									<p
										class="truncate border-l-2 border-[rgba(255,255,255,0.08)] pl-2.5 font-mono text-xs text-muted-foreground"
									>
										{f.path}
										<span class="text-muted-foreground/60">· {f.type}</span>
									</p>
								{/each}
								{#if ctx.fields.length > 3}
									<p
										class="pl-2.5 font-mono text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
									>
										+{ctx.fields.length - 3} more
									</p>
								{/if}
							</div>
						{/if}
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</section>
