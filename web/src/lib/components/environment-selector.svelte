<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Environment } from '$lib/api';
	import * as Select from '$lib/components/ui/select';
	import { Select as SelectPrimitive } from 'bits-ui';
	import { ChevronDown, Layers3, Settings2 } from 'lucide-svelte';

	type Props = {
		environments?: Environment[];
		selectedEnvironment?: Environment;
	};

	let { environments = [], selectedEnvironment }: Props = $props();

	function targetHref(id: string) {
		const next = new URL(page.url);
		const path = next.pathname;
		if (!(path === '/' || path.startsWith('/flags') || path.startsWith('/api-keys'))) {
			next.pathname = '/';
		}
		next.searchParams.set('environment', id);
		return `${next.pathname}${next.search}`;
	}

	async function selectEnvironment(id: string) {
		await goto(targetHref(id));
	}
</script>

{#if selectedEnvironment && environments.length > 0}
	<Select.Root type="single" value={selectedEnvironment.id} onValueChange={selectEnvironment}>
		<SelectPrimitive.Trigger
			class="group inline-flex h-8 min-w-44 items-center gap-2 rounded-sm border border-[rgb(var(--app-accent-rgb)/0.42)] bg-[rgb(var(--app-accent-rgb)/0.1)] pl-2 pr-1.5 text-left text-app-accent shadow-[0_0_0_1px_rgb(var(--app-accent-rgb)/0.08),0_0_22px_rgb(var(--app-accent-rgb)/0.08)] transition-colors hover:border-[rgb(var(--app-accent-rgb)/0.7)] hover:bg-[rgb(var(--app-accent-rgb)/0.15)] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[rgb(var(--app-accent-rgb)/0.75)]"
			title="Environment selector"
		>
			<Layers3 class="h-3.5 w-3.5 shrink-0" />
			<span
				class="hidden text-[0.62rem] uppercase tracking-[0.14em] text-[rgb(var(--app-accent-rgb)/0.72)] sm:inline"
			>
				env
			</span>
			<span class="ml-auto max-w-28 truncate font-mono text-xs text-foreground sm:max-w-36">
				{selectedEnvironment.key}
			</span>
			<ChevronDown
				class="h-3.5 w-3.5 shrink-0 transition-transform group-data-[state=open]:rotate-180"
			/>
		</SelectPrimitive.Trigger>

		<Select.Content align="end" class="w-[var(--bits-select-anchor-width)] min-w-44 border-[rgb(var(--app-accent-rgb)/0.34)]">
			<Select.Group>
				<Select.GroupHeading
					class="px-2 py-1.5 text-[0.62rem] uppercase tracking-[0.14em] text-[rgb(var(--app-accent-rgb)/0.72)]"
				>
					environment
				</Select.GroupHeading>
				{#each environments as env (env.id)}
					<Select.Item value={env.id} label={env.key}>
						<span class="block truncate font-mono">{env.key}</span>
						<span class="block truncate text-[0.65rem] text-muted-foreground">{env.name}</span>
					</Select.Item>
				{/each}
			</Select.Group>
			<a
				href="/environments"
				class="mt-1 flex items-center gap-2 rounded-[2px] border-t border-[rgb(var(--app-accent-rgb)/0.2)] px-2 py-2 text-[0.62rem] uppercase tracking-[0.14em] text-muted-foreground transition-colors hover:text-app-accent"
			>
				<Settings2 class="h-3.5 w-3.5 shrink-0" />
				manage environments
			</a>
		</Select.Content>
	</Select.Root>
{/if}
