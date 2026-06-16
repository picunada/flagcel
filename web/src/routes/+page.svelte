<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { APIError, type ContextSchema, type Environment, type Flag } from '$lib/api';
	import Badge from '$lib/components/ui/badge.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import ThemedSelect from '$lib/components/ui/themed-select.svelte';
	import { cn } from '$lib/utils';
	import { formatFlagValue, valueBadgeVariant } from '$lib/values';
	import { LayoutGrid, ListFilter, Plus, Search, TableProperties } from 'lucide-svelte';
	import type { PageProps } from './$types';

	type StatusFilter = 'all' | 'on' | 'off';
	type SortKey = 'recent' | 'rules' | 'default' | 'status' | 'key' | 'context';
	type ViewMode = 'table' | 'cards';

	const contextFilterBaseOptions = [
		{ value: 'all', label: 'all contexts' },
		{ value: 'none', label: 'no context' }
	];
	const sortOptions = [
		{ value: 'recent', label: 'recently changed' },
		{ value: 'rules', label: 'rule count' },
		{ value: 'default', label: 'default value' },
		{ value: 'status', label: 'status' },
		{ value: 'context', label: 'context' },
		{ value: 'key', label: 'key' }
	];

	let { data }: PageProps = $props();

	const flags = $derived<Flag[]>(data.flags);
	const selectedEnvironment = $derived<Environment | undefined>(data.selectedEnvironment);
	const contexts = $derived<ContextSchema[]>(data.contexts);
	const contextById = $derived.by(() => new Map(contexts.map((ctx) => [ctx.id, ctx])));
	const contextFilterOptions = $derived([
		...contextFilterBaseOptions,
		...contexts.map((ctx) => ({ value: ctx.id, label: ctx.name }))
	]);

	let error = $state<string | null>(null);
	let query = $state('');
	let statusFilter = $state<StatusFilter>('all');
	let contextFilter = $state('all');
	let sortKey = $state<SortKey>('recent');
	let viewMode = $state<ViewMode>('table');

	const filtered = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		return flags
			.filter((flag) => {
				const statusMatches =
					statusFilter === 'all' ||
					(statusFilter === 'on' && flag.enabled) ||
					(statusFilter === 'off' && !flag.enabled);
				const contextMatches =
					contextFilter === 'all' ||
					(contextFilter === 'none' && !flag.context_id) ||
					flag.context_id === contextFilter;
				const contextName = contextLabel(flag).toLowerCase();
				const searchMatches =
					needle === '' ||
					flag.key.toLowerCase().includes(needle) ||
					contextName.includes(needle) ||
					flag.rules.some((rule) => rule.expression.toLowerCase().includes(needle));

				return statusMatches && contextMatches && searchMatches;
			})
			.sort((a, b) => compareFlags(a, b));
	});

	const activeFilters = $derived(
		Number(statusFilter !== 'all') + Number(contextFilter !== 'all') + Number(query.trim() !== '')
	);

	function compareFlags(a: Flag, b: Flag) {
		if (sortKey === 'recent') return timestamp(b) - timestamp(a) || a.key.localeCompare(b.key);
		if (sortKey === 'rules') return b.rules.length - a.rules.length || a.key.localeCompare(b.key);
		if (sortKey === 'default') {
			return formatFlagValue(a.default_value).localeCompare(formatFlagValue(b.default_value)) || a.key.localeCompare(b.key);
		}
		if (sortKey === 'status') return Number(b.enabled) - Number(a.enabled) || a.key.localeCompare(b.key);
		if (sortKey === 'context') {
			return contextLabel(a).localeCompare(contextLabel(b)) || a.key.localeCompare(b.key);
		}
		return a.key.localeCompare(b.key);
	}

	function contextLabel(flag: Flag) {
		if (!flag.context_id) return 'no context';
		return contextById.get(flag.context_id)?.name ?? 'missing context';
	}

	function timestamp(flag: Flag) {
		return Date.parse(flag.updated_at) || 0;
	}

	function formatUpdated(value: string) {
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return 'unknown';
		return new Intl.DateTimeFormat(undefined, {
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		}).format(date);
	}

	function resetFilters() {
		query = '';
		statusFilter = 'all';
		contextFilter = 'all';
	}

	function flagHref(key: string) {
		const query = selectedEnvironment ? `?environment=${encodeURIComponent(selectedEnvironment.id)}` : '';
		return `/flags/${encodeURIComponent(key)}${query}`;
	}

	function newFlagHref() {
		const query = selectedEnvironment ? `?environment=${encodeURIComponent(selectedEnvironment.id)}` : '';
		return `/flags/new${query}`;
	}

	async function refresh() {
		error = null;
		try {
			await invalidateAll();
		} catch (e) {
			error = e instanceof APIError ? e.message : 'Failed to load flags';
		}
	}
</script>

<section class="space-y-8">
	<header class="space-y-3">
		<p class="text-xs uppercase tracking-[0.18em] text-muted-foreground">
			feature flags · cel-based targeting
		</p>
		<h1 class="text-balance text-3xl font-normal leading-tight sm:text-4xl">
			Roll out features<br />on your own terms.
		</h1>
		<p class="max-w-xl text-sm text-[rgba(255,255,255,0.78)] sm:text-base">
			Targeting rules written as CEL expressions, evaluated server-side, persisted in
			Postgres. No SaaS, no DSL to learn.
		</p>
	</header>

	<Card class="motion-panel overflow-hidden">
		<div class="border-b border-border p-3">
			<div class="flex flex-col gap-3">
				<div
					class="grid items-center gap-2 md:grid-cols-[minmax(14rem,1fr)_auto] lg:grid-cols-[minmax(14rem,1fr)_auto_12rem_11rem]"
				>
					<label class="relative block min-w-0">
						<span class="sr-only">Search flags</span>
						<Search class="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
						<Input bind:value={query} placeholder="search key, context, rule" class="w-full pl-8" />
					</label>

					<div class="flex h-9 shrink-0 rounded-sm border border-border p-0.5" aria-label="Status filter">
						<button
							type="button"
							aria-pressed={statusFilter === 'all'}
							class={cn(
								'h-full rounded-[2px] px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors',
								statusFilter === 'all'
									? 'bg-[rgba(255,255,255,0.1)] text-foreground'
									: 'text-muted-foreground hover:text-foreground'
							)}
							onclick={() => (statusFilter = 'all')}
						>
							all
						</button>
						<button
							type="button"
							aria-pressed={statusFilter === 'on'}
							class={cn(
								'h-full rounded-[2px] px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors',
								statusFilter === 'on'
									? 'bg-[rgba(255,255,255,0.1)] text-foreground'
									: 'text-muted-foreground hover:text-foreground'
							)}
							onclick={() => (statusFilter = 'on')}
						>
							on
						</button>
						<button
							type="button"
							aria-pressed={statusFilter === 'off'}
							class={cn(
								'h-full rounded-[2px] px-3 text-[0.65rem] uppercase tracking-[0.12em] transition-colors',
								statusFilter === 'off'
									? 'bg-[rgba(255,255,255,0.1)] text-foreground'
									: 'text-muted-foreground hover:text-foreground'
							)}
							onclick={() => (statusFilter = 'off')}
						>
							off
						</button>
					</div>

					<label class="block">
						<span class="sr-only">Context filter</span>
						<ThemedSelect value={contextFilter} options={contextFilterOptions} onchange={(v) => (contextFilter = v)} />
					</label>

					<label class="relative block">
						<span class="sr-only">Sort flags</span>
						<ListFilter class="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
						<ThemedSelect
							value={sortKey}
							options={sortOptions}
							onchange={(v) => (sortKey = v as SortKey)}
							buttonClass="pl-8 pr-2.5"
						/>
					</label>
				</div>

				<div class="flex items-center justify-between gap-2">
					<div class="hidden text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground sm:block" aria-live="polite">
						{filtered.length} shown
					</div>
					<div class="ml-auto flex items-center gap-2">
						<div class="flex h-9 rounded-sm border border-border p-0.5" aria-label="View mode">
							<button
								type="button"
								title="Table view"
								aria-label="Table view"
								aria-pressed={viewMode === 'table'}
								class={cn(
									'inline-flex h-full w-8 items-center justify-center rounded-[2px] transition-colors',
									viewMode === 'table'
										? 'bg-[rgba(255,255,255,0.1)] text-foreground'
										: 'text-muted-foreground hover:text-foreground'
								)}
								onclick={() => (viewMode = 'table')}
							>
								<TableProperties class="h-3.5 w-3.5" />
							</button>
							<button
								type="button"
								title="Card view"
								aria-label="Card view"
								aria-pressed={viewMode === 'cards'}
								class={cn(
									'inline-flex h-full w-8 items-center justify-center rounded-[2px] transition-colors',
									viewMode === 'cards'
										? 'bg-[rgba(255,255,255,0.1)] text-foreground'
										: 'text-muted-foreground hover:text-foreground'
								)}
								onclick={() => (viewMode = 'cards')}
							>
								<LayoutGrid class="h-3.5 w-3.5" />
							</button>
						</div>
						<Button href={newFlagHref()} size="default" class="h-9">
							<Plus class="h-3.5 w-3.5" /> new flag
						</Button>
					</div>
				</div>
			</div>
		</div>

		{#if error}
			<div class="p-8 text-center">
				<p class="text-sm text-destructive">{error}</p>
				<Button variant="default" class="mt-4" onclick={refresh}>retry</Button>
			</div>
		{:else if flags.length === 0}
			<div class="flex flex-col items-center gap-4 p-12 text-center">
				<p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">[ no flags yet ]</p>
				<p class="max-w-sm text-sm text-[rgba(255,255,255,0.7)]">
					Create your first flag to start routing evaluations.
				</p>
				<Button href={newFlagHref()} class="mt-2">
					<Plus class="h-3.5 w-3.5" /> new flag
				</Button>
			</div>
		{:else if filtered.length === 0}
			<div class="flex flex-col items-center gap-4 p-12 text-center">
				<p class="text-xs uppercase tracking-[0.14em] text-muted-foreground">[ no matches ]</p>
				<p class="max-w-sm text-sm text-[rgba(255,255,255,0.7)]">
					No flags match the current filters.
				</p>
				{#if activeFilters > 0}
					<Button variant="ghost" size="sm" onclick={resetFilters}>clear filters</Button>
				{/if}
			</div>
		{:else if viewMode === 'cards'}
			<div class="motion-list grid gap-3 p-3 sm:grid-cols-2">
				{#each filtered as flag (flag.key)}
					<a href={flagHref(flag.key)} class="group block">
						<Card hoverable class="flex h-full flex-col gap-4 p-5">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0 flex-1">
									<p class="truncate text-base font-medium group-hover:text-foreground">
										{flag.key}
									</p>
									<p class="mt-1 text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground">
										{flag.type} · {flag.rules.length} rule{flag.rules.length === 1 ? '' : 's'} · default
										{formatFlagValue(flag.default_value)}
									</p>
								</div>
								{#if flag.enabled}
									<Badge variant="success" dot>on</Badge>
								{:else}
									<Badge variant="muted" dot>off</Badge>
								{/if}
							</div>
							<div class="grid grid-cols-2 gap-2 border-y border-border py-3 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
								<div class="min-w-0">
									<p>context</p>
									<p class="mt-1 truncate text-xs normal-case tracking-normal text-foreground/80">
										{contextLabel(flag)}
									</p>
								</div>
								<div class="text-right">
									<p>changed</p>
									<p class="mt-1 text-xs normal-case tracking-normal text-foreground/80">
										{formatUpdated(flag.updated_at)}
									</p>
								</div>
							</div>
							{#if flag.rules.length > 0}
								<div class="mt-auto space-y-1.5">
									{#each flag.rules.slice(0, 2) as rule (rule.id)}
										<p class="truncate border-l-2 border-[rgba(255,255,255,0.08)] pl-2.5 text-xs text-muted-foreground">
											{rule.expression}
										</p>
									{/each}
									{#if flag.rules.length > 2}
										<p class="pl-2.5 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
											+{flag.rules.length - 2} more
										</p>
									{/if}
								</div>
							{/if}
							<span class="mt-auto text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors group-hover:text-foreground">
								open
							</span>
						</Card>
					</a>
				{/each}
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full min-w-[760px] border-collapse text-left">
					<thead class="border-b border-border bg-[rgba(255,255,255,0.025)]">
						<tr class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground">
							<th class="w-[30%] px-3 py-2 font-normal">key</th>
							<th class="px-3 py-2 font-normal">status</th>
							<th class="px-3 py-2 font-normal">context</th>
							<th class="px-3 py-2 text-right font-normal">rules</th>
							<th class="px-3 py-2 font-normal">type/default</th>
							<th class="px-3 py-2 font-normal">changed</th>
							<th class="px-3 py-2 text-right font-normal">action</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-border">
						{#each filtered as flag (flag.key)}
							<tr class="group transition-colors hover:bg-[rgba(255,255,255,0.035)]">
								<td class="px-3 py-2 align-middle">
									<a
										href={flagHref(flag.key)}
										class="block max-w-[20rem] truncate text-sm text-foreground underline-offset-4 hover:underline"
									>
										{flag.key}
									</a>
								</td>
								<td class="px-3 py-2 align-middle">
									{#if flag.enabled}
										<Badge variant="success" dot>on</Badge>
									{:else}
										<Badge variant="muted" dot>off</Badge>
									{/if}
								</td>
								<td class="px-3 py-2 align-middle text-xs text-muted-foreground">
									<span class="block max-w-[12rem] truncate">{contextLabel(flag)}</span>
								</td>
								<td class="px-3 py-2 text-right align-middle text-sm tabular-nums">
									{flag.rules.length}
								</td>
								<td class="px-3 py-2 align-middle">
									<Badge variant={valueBadgeVariant(flag.default_value)}>
										{flag.type}: {formatFlagValue(flag.default_value)}
									</Badge>
								</td>
								<td class="px-3 py-2 align-middle text-xs text-muted-foreground">
									{formatUpdated(flag.updated_at)}
								</td>
								<td class="px-3 py-2 text-right align-middle">
									<Button href={flagHref(flag.key)} variant="ghost" size="sm">
										open
									</Button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</Card>
</section>
