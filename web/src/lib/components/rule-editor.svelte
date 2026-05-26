<script lang="ts">
	import { untrack, tick } from 'svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import { cn } from '$lib/utils';
	import type { ContextField, ContextSchema, CreateRuleRequest, Rule } from '$lib/api';

	type Props = {
		rule?: Rule;
		context?: ContextSchema | null;
		submitting?: boolean;
		error?: string | null;
		onsave: (data: CreateRuleRequest) => void | Promise<void>;
		oncancel?: () => void;
		submitLabel?: string;
		class?: string;
	};

	let {
		rule,
		context = null,
		submitting = false,
		error = null,
		onsave,
		oncancel,
		submitLabel = 'save rule',
		class: className
	}: Props = $props();

	let expression = $state(untrack(() => rule?.expression ?? ''));
	let percentage = $state(untrack(() => String(rule?.rollout.percentage ?? 100)));
	let bucketBy = $state(untrack(() => rule?.rollout.bucket_by ?? ''));

	const parsedPercentage = $derived(Number(percentage));
	const canSubmit = $derived(
		expression.trim().length > 0 &&
			Number.isFinite(parsedPercentage) &&
			parsedPercentage >= 0 &&
			parsedPercentage <= 100 &&
			!submitting
	);

	// --- autocomplete ---
	let textarea: HTMLTextAreaElement | null = $state(null);
	let cursor = $state(0);
	let acOpen = $state(false);
	let acIndex = $state(0);

	const candidates = $derived<ContextField[]>(context?.fields ?? []);

	function tokenAt(text: string, pos: number): { start: number; value: string } {
		let start = pos;
		while (start > 0 && /[A-Za-z0-9_.]/.test(text[start - 1])) start--;
		return { start, value: text.slice(start, pos) };
	}

	const currentToken = $derived(tokenAt(expression, cursor));

	const suggestions = $derived.by(() => {
		if (!candidates.length) return [];
		const t = currentToken.value;
		if (!t) return [];
		const lower = t.toLowerCase();
		return candidates
			.filter((f) => f.path.toLowerCase().startsWith(lower) && f.path !== t)
			.slice(0, 8);
	});

	$effect(() => {
		if (suggestions.length === 0) {
			acOpen = false;
			acIndex = 0;
		} else if (acIndex >= suggestions.length) {
			acIndex = 0;
		}
	});

	function handleCursor() {
		if (!textarea) return;
		cursor = textarea.selectionStart;
		if (suggestions.length > 0) acOpen = true;
	}

	function handleInput() {
		handleCursor();
	}

	async function insert(s: ContextField) {
		if (!textarea) return;
		const { start } = currentToken;
		const end = cursor;
		const before = expression.slice(0, start);
		const after = expression.slice(end);
		const next = before + s.path + after;
		expression = next;
		acOpen = false;
		await tick();
		const newPos = (before + s.path).length;
		textarea.focus();
		textarea.setSelectionRange(newPos, newPos);
		cursor = newPos;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (acOpen && suggestions.length > 0) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				acIndex = (acIndex + 1) % suggestions.length;
				return;
			}
			if (e.key === 'ArrowUp') {
				e.preventDefault();
				acIndex = (acIndex - 1 + suggestions.length) % suggestions.length;
				return;
			}
			if (e.key === 'Enter' || e.key === 'Tab') {
				e.preventDefault();
				insert(suggestions[acIndex]);
				return;
			}
			if (e.key === 'Escape') {
				e.preventDefault();
				acOpen = false;
				return;
			}
		}
		// open on Ctrl+Space even if closed
		if ((e.ctrlKey || e.metaKey) && e.key === ' ') {
			e.preventDefault();
			handleCursor();
			if (suggestions.length > 0) acOpen = true;
		}
	}

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		if (!canSubmit) return;
		await onsave({
			expression: expression.trim(),
			rollout: {
				percentage: Math.round(parsedPercentage),
				...(bucketBy.trim() ? { bucket_by: bucketBy.trim() } : {})
			}
		});
	}
</script>

<form onsubmit={submit} class={cn('space-y-5', className)}>
	<div class="space-y-2">
		<label
			for="expression"
			class="text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
		>
			expression · cel
			{#if context}
				<span class="text-muted-foreground/70">· {context.name}</span>
			{/if}
		</label>
		<div class="relative">
			<textarea
				id="expression"
				bind:this={textarea}
				bind:value={expression}
				oninput={handleInput}
				onkeyup={handleCursor}
				onclick={handleCursor}
				onkeydown={handleKeydown}
				onblur={() => setTimeout(() => (acOpen = false), 120)}
				rows="3"
				required
				placeholder={'user.country == "US" && user.plan == "pro"'}
				class="flex w-full rounded-sm border border-input bg-transparent px-3 py-2 font-mono text-sm transition-colors placeholder:text-muted-foreground/60 focus-visible:outline-none focus-visible:border-[rgba(255,255,255,0.36)] disabled:cursor-not-allowed disabled:opacity-50"
			></textarea>
			{#if acOpen && suggestions.length > 0}
				<div
					role="listbox"
					class="glass-panel motion-pop absolute left-0 right-0 top-full z-20 mt-1 max-h-64 overflow-y-auto rounded-sm py-1 shadow-lg"
				>
					{#each suggestions as s, i (s.path)}
						<button
							type="button"
							role="option"
							aria-selected={i === acIndex}
							tabindex="-1"
							onmousedown={(e) => {
								e.preventDefault();
								insert(s);
							}}
							onmouseenter={() => (acIndex = i)}
							class={cn(
								'flex w-full items-center justify-between gap-3 px-3 py-1.5 text-left font-mono text-xs transition-colors',
								i === acIndex
									? 'bg-[rgba(255,255,255,0.06)] text-foreground'
									: 'text-muted-foreground hover:text-foreground'
							)}
						>
							<span class="truncate">{s.path}</span>
							<span class="text-muted-foreground/70 text-[0.65rem] uppercase tracking-[0.12em]">
								{s.type}
							</span>
						</button>
					{/each}
				</div>
			{/if}
		</div>
		{#if context && candidates.length === 0}
			<p class="text-[0.65rem] text-muted-foreground">
				context "{context.name}" has no fields yet
			</p>
		{:else if context}
			<p class="text-[0.65rem] text-muted-foreground">
				start typing for suggestions · ↑/↓ navigate · enter to insert
			</p>
		{/if}
	</div>

	<div class="grid gap-4 sm:grid-cols-2">
		<div class="space-y-2">
			<label
				for="percentage"
				class="text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
			>
				rollout %
			</label>
			<Input
				id="percentage"
				type="number"
				min="0"
				max="100"
				step="1"
				required
				bind:value={percentage}
			/>
		</div>

		<div class="space-y-2">
			<label
				for="bucket-by"
				class="text-[0.7rem] uppercase tracking-[0.14em] text-muted-foreground"
			>
				bucket by · optional
			</label>
			<Input id="bucket-by" bind:value={bucketBy} placeholder="user.id" />
		</div>
	</div>

	{#if error}
		<p class="text-xs text-destructive">{error}</p>
	{/if}

	<div class="flex justify-end gap-2 border-t border-border/60 pt-4">
		{#if oncancel}
			<Button variant="ghost" type="button" onclick={oncancel} disabled={submitting}>
				cancel
			</Button>
		{/if}
		<Button variant="solid" type="submit" disabled={!canSubmit}>
			{submitting ? 'saving…' : submitLabel}
		</Button>
	</div>
</form>
