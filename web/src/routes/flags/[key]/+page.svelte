<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api, APIError, type Flag } from '$lib/api';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import BoolToggle from '$lib/components/ui/bool-toggle.svelte';
	import SectionHeader from '$lib/components/ui/section-header.svelte';
	import { Trash2 } from 'lucide-svelte';

	const key = $derived(page.params.key);

	let flag = $state<Flag | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let saving = $state(false);

	$effect(() => {
		key;
		load();
	});

	async function load() {
		loading = true;
		error = null;
		try {
			flag = await api.getFlag(key);
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to load flag';
		} finally {
			loading = false;
		}
	}

	async function patch(updates: Partial<Pick<Flag, 'enabled' | 'default_value'>>) {
		if (!flag) return;
		const prev = { enabled: flag.enabled, default_value: flag.default_value };
		flag = { ...flag, ...updates };
		saving = true;
		error = null;
		try {
			await api.createFlag({
				key: flag.key,
				enabled: flag.enabled,
				default_value: flag.default_value,
				rules: flag.rules.map((r) => ({ expression: r.expression, rollout: r.rollout }))
			});
		} catch (e) {
			flag = { ...flag, ...prev };
			error = e instanceof APIError ? e.message : 'Failed to save flag';
		} finally {
			saving = false;
		}
	}

	async function remove() {
		if (!flag) return;
		if (!confirm(`Delete flag "${flag.key}"? This removes all of its rules.`)) return;
		try {
			await api.deleteFlag(flag.key);
			await goto('/');
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to delete flag';
		}
	}
</script>

<div class="space-y-10">
	<a
		href="/"
		class="inline-flex items-center gap-1.5 font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-foreground"
	>
		← all flags
	</a>

	{#if loading}
		<Card class="h-40 animate-pulse" />
	{:else if error && !flag}
		<Card class="p-8 text-center">
			<p class="text-sm text-destructive">{error}</p>
			<Button class="mt-4" onclick={load}>retry</Button>
		</Card>
	{:else if flag}
		<header class="flex flex-wrap items-start justify-between gap-4">
			<div class="space-y-3">
				<p
					class="font-mono text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground"
				>
					[ flag ]
				</p>
				<h1 class="font-mono text-3xl font-normal tracking-tight sm:text-4xl">
					{flag.key}
				</h1>
			</div>
			<Button variant="destructive" onclick={remove}>
				<Trash2 class="h-3.5 w-3.5" /> delete
			</Button>
		</header>

		<Card class="divide-y divide-border/60">
			<div class="flex items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="font-mono text-sm">enabled</p>
					<p class="text-xs text-muted-foreground">
						when off, the default value is returned for every request
					</p>
				</div>
				<BoolToggle
					value={flag.enabled}
					disabled={saving}
					onchange={(v) => patch({ enabled: v })}
				/>
			</div>
			<div class="flex items-center justify-between gap-4 p-5">
				<div class="space-y-1">
					<p class="font-mono text-sm">default value</p>
					<p class="text-xs text-muted-foreground">returned when no rule matches</p>
				</div>
				<BoolToggle
					value={flag.default_value}
					disabled={saving}
					onchange={(v) => patch({ default_value: v })}
				/>
			</div>
		</Card>

		{#if error}
			<p class="font-mono text-xs text-destructive">{error}</p>
		{/if}

		<section class="space-y-4">
			<SectionHeader>rules · evaluated top-to-bottom</SectionHeader>
			{#if flag.rules.length === 0}
				<Card class="p-8 text-center">
					<p
						class="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground"
					>
						[ no rules ]
					</p>
					<p class="mt-3 text-sm text-[rgba(255,255,255,0.7)]">
						Requests fall through to the default value.
					</p>
				</Card>
			{:else}
				<div class="space-y-2">
					{#each flag.rules as rule, i (rule.id)}
						<Card class="p-5">
							<div class="flex items-start gap-4">
								<div
									class="font-mono text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground"
								>
									#{String(i + 1).padStart(2, '0')}
								</div>
								<div class="min-w-0 flex-1 space-y-3">
									<pre
										class="overflow-x-auto border-l-2 border-success/40 bg-[rgba(255,255,255,0.02)] py-2 pl-3 font-mono text-sm text-foreground">{rule.expression}</pre>
									<div
										class="flex flex-wrap items-center gap-4 font-mono text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
									>
										<span>
											rollout
											<span class="text-foreground">{rule.rollout.percentage}%</span>
										</span>
										{#if rule.rollout.bucket_by}
											<span>
												bucket by
												<span class="text-foreground">{rule.rollout.bucket_by}</span>
											</span>
										{/if}
									</div>
								</div>
							</div>
						</Card>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</div>
