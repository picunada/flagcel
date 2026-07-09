<script lang="ts">
	import { tick, type Snippet } from 'svelte';
	import type { ContextField } from '$lib/api';
	import { cn } from '$lib/utils';

	type Token = {
		text: string;
		kind: 'string' | 'operator' | 'number' | 'field' | 'identifier' | 'unknown' | 'plain';
	};

	type ValidationState = {
		message: string;
		tone: 'muted' | 'warning' | 'invalid' | 'valid';
	};

	type Props = {
		value: string;
		fields?: ContextField[];
		contextName?: string | null;
		disabled?: boolean;
		submitting?: boolean;
		onsave?: () => void | Promise<void>;
		oncancel?: () => void;
		oninput: (value: string) => void;
		aside?: Snippet;
		class?: string;
	};

	let {
		value,
		fields = [],
		contextName = null,
		disabled = false,
		submitting = false,
		onsave,
		oncancel,
		oninput,
		aside,
		class: className
	}: Props = $props();

	let textarea: HTMLTextAreaElement | null = $state(null);
	let cursor = $state(0);
	let acOpen = $state(false);
	let acIndex = $state(0);
	let suggestionStart = $state(0);
	let suggestionEnd = $state(0);

	const knownPaths = $derived(fields.map((field) => field.path));

	function tokenAt(text: string, pos: number): { start: number; value: string } {
		let start = pos;
		while (start > 0 && /[A-Za-z0-9_.]/.test(text[start - 1])) start--;
		return { start, value: text.slice(start, pos) };
	}

	const currentToken = $derived(tokenAt(value, cursor));

	const suggestions = $derived.by(() => {
		if (!fields.length) return [];
		const token = currentToken.value;
		if (!token) return [];
		const lower = token.toLowerCase();
		return fields
			.filter((field) => field.path.toLowerCase().startsWith(lower) && field.path !== token)
			.slice(0, 8);
	});

	const tokens = $derived(tokenize(value, knownPaths));
	const validation = $derived(validateExpression(value, knownPaths, Boolean(contextName)));

	$effect(() => {
		if (suggestions.length === 0) {
			acOpen = false;
			acIndex = 0;
		} else if (acIndex >= suggestions.length) {
			acIndex = 0;
		}
	});

	function tokenize(expr: string, known: string[]): Token[] {
		const re = /("[^"]*"?)|(&&|\|\||==|!=|>=|<=|>|<|!|\(|\))|(\d+(?:\.\d+)?)|([A-Za-z_][\w.]*)|(\s+)|(.)/g;
		const out: Token[] = [];
		let match: RegExpExecArray | null;

		while ((match = re.exec(expr))) {
			if (match[1]) {
				out.push({ text: match[0], kind: 'string' });
			} else if (match[2]) {
				out.push({ text: match[0], kind: 'operator' });
			} else if (match[3]) {
				out.push({ text: match[0], kind: 'number' });
			} else if (match[4]) {
				const text = match[4];
				const knownToken =
					known.length === 0 ||
					known.some((path) => path === text || path.startsWith(text) || text.startsWith(`${path}.`));
				out.push({
					text,
					kind: isKeyword(text) ? 'operator' : knownToken ? 'field' : 'unknown'
				});
			} else {
				out.push({ text: match[0], kind: 'plain' });
			}
		}

		return out.length ? out : [{ text: '', kind: 'plain' }];
	}

	function isKeyword(token: string): boolean {
		return ['true', 'false', 'null', 'in'].includes(token);
	}

	function validateExpression(expr: string, known: string[], hasContext: boolean): ValidationState {
		const trimmed = expr.trim();
		if (!trimmed) {
			return {
				message: hasContext
					? 'start typing for suggestions · chips insert at cursor'
					: 'start typing a CEL expression',
				tone: 'muted'
			};
		}

		const clause =
			'[A-Za-z_][\\w.]*\\s*(==|!=|>=|<=|>|<|in)\\s*("[^"]*"|\\d+(\\.\\d+)?|true|false|[A-Za-z_][\\w.]*)';
		const full = new RegExp(`^\\s*${clause}(\\s*(&&|\\|\\|)\\s*${clause})*\\s*$`);
		if (!full.test(trimmed)) {
			return {
				message: 'incomplete expression · expected: field == "value" [&& ...]',
				tone: 'muted'
			};
		}

		if (known.length > 0) {
			const unknown = identifiers(trimmed).find(
				(token) => !isKeyword(token) && !known.includes(token)
			);
			if (unknown) {
				return {
					message: `unknown field: ${unknown} · not in ${contextName ?? 'selected'} context`,
					tone: 'invalid'
				};
			}
		}

		return {
			message: hasContext ? `✓ valid · ${contextName} context` : '✓ valid',
			tone: 'valid'
		};
	}

	function identifiers(expr: string): string[] {
		return expr
			.replace(/"[^"]*"/g, '')
			.match(/[A-Za-z_][\w.]*/g) ?? [];
	}

	function handleCursor() {
		if (!textarea) return;
		cursor = textarea.selectionStart;
		const token = tokenAt(value, cursor);
		suggestionStart = token.start;
		suggestionEnd = cursor;
		if (suggestions.length > 0) acOpen = true;
	}

	function handleInput(e: Event) {
		const next = (e.currentTarget as HTMLTextAreaElement).value;
		oninput(next);
		cursor = (e.currentTarget as HTMLTextAreaElement).selectionStart;
		const token = tokenAt(next, cursor);
		suggestionStart = token.start;
		suggestionEnd = cursor;
		acOpen = suggestions.length > 0;
	}

	async function insertSuggestion(field: ContextField) {
		const before = value.slice(0, suggestionStart);
		const after = value.slice(suggestionEnd);
		const next = `${before}${field.path}${after}`;
		const nextCursor = before.length + field.path.length;
		oninput(next);
		acOpen = false;
		await tick();
		textarea?.focus();
		textarea?.setSelectionRange(nextCursor, nextCursor);
		cursor = nextCursor;
	}

	async function insertChip(text: string) {
		const start = textarea?.selectionStart ?? value.length;
		const end = textarea?.selectionEnd ?? value.length;
		const before = value.slice(0, start);
		const after = value.slice(end);
		const needsLeadingSpace = before.length > 0 && !/\s$/.test(before);
		const needsTrailingSpace = after.length > 0 && !/^\s/.test(after);
		const insert = `${needsLeadingSpace ? ' ' : ''}${text}${needsTrailingSpace ? ' ' : ''}`;
		const next = `${before}${insert}${after}`;
		const nextCursor = start + insert.length;
		oninput(next);
		acOpen = false;
		await tick();
		textarea?.focus();
		textarea?.setSelectionRange(nextCursor, nextCursor);
		cursor = nextCursor;
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
				void insertSuggestion(suggestions[acIndex]);
				return;
			}
			if (e.key === 'Escape') {
				e.preventDefault();
				acOpen = false;
				return;
			}
		}

		if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
			e.preventDefault();
			void onsave?.();
			return;
		}

		if (e.key === 'Escape') {
			e.preventDefault();
			oncancel?.();
		}
	}
</script>

{#snippet editor()}
	<div class="relative h-full min-w-0">
		<div
			aria-hidden="true"
			class="pointer-events-none absolute inset-0 z-20 overflow-hidden whitespace-pre-wrap break-words rounded-sm border border-transparent px-3 py-2 font-mono text-sm leading-6"
		>
			{#each tokens as token}
				<span
					class={cn(
						token.kind === 'string' && 'text-valid',
						token.kind === 'operator' && 'text-cel-warning',
						token.kind === 'number' && 'text-cel-warning',
						token.kind === 'field' && 'text-valid',
						token.kind === 'unknown' && 'text-destructive',
						(token.kind === 'identifier' || token.kind === 'plain') && 'text-foreground'
					)}
				>{token.text}</span>
			{/each}
		</div>
		<textarea
			id="expression"
			bind:this={textarea}
			value={value}
			oninput={handleInput}
			onkeyup={handleCursor}
			onclick={handleCursor}
			onkeydown={handleKeydown}
			onblur={() => setTimeout(() => (acOpen = false), 120)}
			rows="2"
			required
			disabled={disabled || submitting}
			aria-invalid={validation.tone === 'invalid'}
			placeholder={'user.country == "US" && user.membership == "pro"'}
			spellcheck="false"
			autocomplete="off"
			autocapitalize="off"
			class={cn(
				'ios-corners-sm relative z-10 flex h-full min-h-16 w-full resize-y border bg-transparent px-3 py-2 font-mono text-sm leading-6 text-transparent caret-foreground transition-colors [-webkit-text-fill-color:transparent] placeholder:text-muted-foreground/50 placeholder:[-webkit-text-fill-color:var(--color-muted-foreground)] focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50',
				validation.tone === 'valid' && 'border-valid/60 focus-visible:border-valid',
				validation.tone === 'warning' && 'border-cel-warning/45 focus-visible:border-cel-warning',
				validation.tone === 'invalid' && 'border-destructive/60 focus-visible:border-destructive',
				validation.tone === 'muted' && 'border-input focus-visible:border-border-hover'
			)}
		></textarea>
		{#if acOpen && suggestions.length > 0}
			<div
				role="listbox"
				class="glass-panel motion-pop absolute left-0 right-0 top-full z-30 mt-1 max-h-64 overflow-y-auto rounded-sm py-1 shadow-lg"
			>
				{#each suggestions as field, i (field.path)}
					<button
						type="button"
						role="option"
						aria-selected={i === acIndex}
						tabindex="-1"
						onmousedown={(e) => {
							e.preventDefault();
							void insertSuggestion(field);
						}}
						onmouseenter={() => (acIndex = i)}
						class={cn(
							'flex w-full cursor-pointer items-center justify-between gap-3 px-3 py-1.5 text-left font-mono text-xs transition-colors',
							i === acIndex
								? 'bg-surface-active text-foreground'
								: 'text-muted-foreground hover:text-foreground'
						)}
					>
						<span class="truncate">{field.path}</span>
						<span class="text-[0.65rem] uppercase tracking-[0.12em] text-muted-foreground/70">
							{field.type}
						</span>
					</button>
				{/each}
			</div>
		{/if}
	</div>
{/snippet}

<div class={cn('space-y-2', className)}>
	{#if aside}
		<div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(14rem,18rem)] lg:items-stretch">
			<div class="flex min-w-0 flex-col gap-2">
				<div class="h-4 shrink-0 text-[0.65rem] uppercase leading-4 tracking-[0.14em] text-muted-foreground">
					expression
				</div>
				<div class="min-h-0 flex-1">
					{@render editor()}
				</div>
			</div>
			{@render aside()}
		</div>
	{:else}
		{@render editor()}
	{/if}

	<p
		aria-live="polite"
		class={cn(
			'font-mono text-[0.65rem] tracking-[0.04em]',
			validation.tone === 'valid' && 'text-valid',
			validation.tone === 'warning' && 'text-cel-warning',
			validation.tone === 'invalid' && 'text-destructive',
			validation.tone === 'muted' && 'text-muted-foreground'
		)}
	>
		{validation.message}
	</p>

	<div class="flex flex-wrap gap-1.5">
		{#each fields as field (field.path)}
			<button
				type="button"
				class="ios-corners-sm cursor-pointer border border-border-control px-2.5 py-1 font-mono text-[0.65rem] text-foreground-soft transition-colors hover:border-border-hover hover:bg-surface-hover hover:text-foreground"
				onclick={() => insertChip(field.path)}
				disabled={disabled || submitting}
			>
				{field.path}
			</button>
		{/each}
		{#each ['== ""', '!= ""', '&&'] as chip}
			<button
				type="button"
				class="ios-corners-sm cursor-pointer border border-border-control px-2.5 py-1 font-mono text-[0.65rem] text-foreground-soft transition-colors hover:border-border-hover hover:bg-surface-hover hover:text-foreground"
				onclick={() => insertChip(chip)}
				disabled={disabled || submitting}
			>
				{chip}
			</button>
		{/each}
	</div>
</div>
