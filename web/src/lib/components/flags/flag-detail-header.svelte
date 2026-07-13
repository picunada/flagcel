<script lang="ts">
	import type { Flag, FlagValue } from '$lib/api';
	import ContextPicker from '$lib/components/context-picker.svelte';
	import Badge from '$lib/components/ui/badge.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import PageHeader from '$lib/components/ui/page-header.svelte';
	import * as Popover from '$lib/components/ui/popover';
	import ValueEditor from '$lib/components/value-editor.svelte';
	import { cn } from '$lib/utils';
	import { formatFlagValue } from '$lib/values';
	import { Power } from 'lucide-svelte';

	type FlagPatch = Partial<Pick<Flag, 'enabled' | 'default_value' | 'context_id'>>;

	type Props = {
		flag: Flag;
		saving: boolean;
		createdBy: string | null;
		updatedBy: string | null;
		formatTimestamp: (iso: string) => string;
		patch: (updates: FlagPatch) => void | Promise<void>;
		updateDefaultValue: (value: FlagValue) => void | Promise<void>;
	};

	let {
		flag,
		saving,
		createdBy,
		updatedBy,
		formatTimestamp,
		patch,
		updateDefaultValue
	}: Props = $props();

	const cellClass =
		'flex h-9 items-center gap-2 px-3 text-left transition-colors duration-200';
</script>

{#snippet flagType()}
	<Badge
		variant="muted"
		class="border border-border-control bg-transparent px-2 py-0.5 font-mono text-[0.6rem] leading-4 tracking-[0.14em]"
	>
		{flag.type}
	</Badge>
{/snippet}

<section class="min-w-0 space-y-4">
	<PageHeader
		eyebrow="[ flag ]"
		title={flag.key}
		titleClass="break-words font-mono"
		titleAfter={flagType}
	/>

	<div
		class="ios-corners-md flex flex-wrap items-center gap-0 overflow-hidden border border-border-control bg-surface-faint p-1"
	>
		<Popover.Root>
			<Popover.Trigger
				class={cn(
					cellClass,
					'ios-corners-sm cursor-pointer hover:bg-surface-hover focus-visible:bg-surface-hover focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50',
					saving && 'opacity-60'
				)}
				disabled={saving}
			>
				<span class="shrink-0 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
					>default</span
				>
				<span class="max-w-40 truncate font-mono text-xs text-foreground">
					{formatFlagValue(flag.default_value)}
				</span>
			</Popover.Trigger>
			<Popover.Content class="w-[min(28rem,calc(100vw-2rem))] p-4">
				<div class="space-y-3">
					<p class="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-muted-foreground">
						default value · {flag.type}
					</p>
					<ValueEditor
						type={flag.type}
						value={flag.default_value}
						id="default-value"
						disabled={saving}
						onchange={updateDefaultValue}
					/>
				</div>
			</Popover.Content>
		</Popover.Root>

		<div class="mx-0.5 h-5 w-px shrink-0 self-center bg-border-control" aria-hidden="true"></div>

		<div class={cn(cellClass, 'ios-corners-sm min-w-0')}>
			<span class="shrink-0 text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground"
				>context</span
			>
			<ContextPicker
				value={flag.context_id ?? null}
				disabled={saving}
				compact
				onchange={(v) => patch({ context_id: v })}
				class="flex h-9 min-w-0 max-w-56 items-center"
				buttonClass="h-9 w-auto min-w-0 border-0 bg-transparent px-0 text-xs shadow-none hover:border-transparent hover:bg-transparent focus-visible:border-transparent focus-visible:ring-0"
			/>
		</div>

		<div class="mx-0.5 h-5 w-px shrink-0 self-center bg-border-control" aria-hidden="true"></div>

		<div class="flex h-9 shrink-0 items-center px-1.5">
			<Button
				size="sm"
				variant={flag.enabled ? 'solid' : 'default'}
				disabled={saving}
				onclick={() => patch({ enabled: !flag.enabled })}
				class="h-7 font-mono"
			>
				<Power class="h-3 w-3" />
				{flag.enabled ? 'enabled' : 'disabled'}
			</Button>
		</div>

		<div class="mx-0.5 h-5 w-px shrink-0 self-center bg-border-control" aria-hidden="true"></div>

		<p
			class="flex h-9 min-w-0 flex-1 items-center overflow-hidden px-3 text-[0.65rem] uppercase leading-4 tracking-[0.08em] text-muted-foreground"
			title={[
				`created ${formatTimestamp(flag.created_at)}`,
				createdBy,
				`updated ${formatTimestamp(flag.updated_at)}`,
				updatedBy
			]
				.filter(Boolean)
				.join(' · ')}
		>
			<span class="truncate">
				created {formatTimestamp(flag.created_at)}
				{#if createdBy}
					· <span class="font-mono lowercase text-foreground">{createdBy}</span>
				{/if}
				· updated {formatTimestamp(flag.updated_at)}
				{#if updatedBy}
					· <span class="font-mono lowercase text-foreground">{updatedBy}</span>
				{/if}
			</span>
		</p>
	</div>
</section>
