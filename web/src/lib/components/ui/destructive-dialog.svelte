<script lang="ts">
	import { tick } from 'svelte';
	import { AlertTriangle, X } from 'lucide-svelte';
	import Button from '$lib/components/ui/button.svelte';

	type Props = {
		open?: boolean;
		title: string;
		description: string;
		details?: string | null;
		confirmationValue?: string | null;
		actionLabel?: string;
		submitting?: boolean;
		error?: string | null;
		onconfirm: () => void | Promise<void>;
		oncancel?: () => void;
	};

	let {
		open = $bindable(false),
		title,
		description,
		details = null,
		confirmationValue = null,
		actionLabel = 'confirm',
		submitting = false,
		error = null,
		onconfirm,
		oncancel
	}: Props = $props();

	let typed = $state('');
	let dialogEl: HTMLElement | undefined = $state();
	let inputEl: HTMLInputElement | undefined = $state();

	const requiresTypedConfirmation = $derived(Boolean(confirmationValue));
	const canConfirm = $derived(
		!submitting && (!requiresTypedConfirmation || typed === confirmationValue)
	);

	$effect(() => {
		if (!open) return;
		typed = '';
		void tick().then(() => {
			(inputEl ?? dialogEl)?.focus();
		});
	});

	function close() {
		if (submitting) return;
		open = false;
		typed = '';
		oncancel?.();
	}

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!canConfirm) return;
		void onconfirm();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			close();
		}
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4">
		<button
			type="button"
			class="absolute inset-0 cursor-default bg-black/70 backdrop-blur-sm"
			aria-label="cancel destructive action"
			disabled={submitting}
			onclick={close}
		></button>

		<div
			class="glass-panel motion-pop relative z-10 w-full max-w-lg rounded-sm p-5"
			role="dialog"
			aria-modal="true"
			aria-labelledby="destructive-dialog-title"
			aria-describedby="destructive-dialog-description"
			tabindex="-1"
			bind:this={dialogEl}
			onkeydown={handleKeydown}
		>
			<form onsubmit={handleSubmit}>
				<div class="flex items-start gap-4">
					<div
						class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-sm border border-[rgba(255,107,107,0.3)] bg-[rgba(255,107,107,0.06)] text-destructive"
					>
						<AlertTriangle class="h-4 w-4" />
					</div>
					<div class="min-w-0 flex-1 space-y-2">
						<h2 id="destructive-dialog-title" class="font-mono text-lg font-normal">
							{title}
						</h2>
						<p id="destructive-dialog-description" class="text-sm text-muted-foreground">
							{description}
						</p>
					</div>
					<button
						type="button"
						class="motion-press rounded-sm p-1 text-muted-foreground transition-colors hover:bg-[rgba(255,255,255,0.04)] hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-40"
						aria-label="cancel"
						disabled={submitting}
						onclick={close}
					>
						<X class="h-4 w-4" />
					</button>
				</div>

				{#if details}
					<pre
						class="mt-4 max-h-36 overflow-auto whitespace-pre-wrap break-words border-l-2 border-destructive/40 bg-[rgba(255,255,255,0.02)] py-2 pl-3 font-mono text-xs text-[rgba(255,255,255,0.82)]">{details}</pre>
				{/if}

				{#if confirmationValue}
					<label class="mt-5 block space-y-2">
						<span class="block text-xs text-muted-foreground">
							Type <code class="font-mono text-foreground">{confirmationValue}</code> to confirm.
						</span>
						<input
							bind:this={inputEl}
							bind:value={typed}
							disabled={submitting}
							autocomplete="off"
							spellcheck={false}
							class="flex h-9 w-full rounded-sm border border-input bg-transparent px-3 py-1 font-mono text-sm transition-all duration-200 ease-out placeholder:text-muted-foreground focus-visible:border-[rgba(255,255,255,0.36)] focus-visible:bg-[rgba(255,255,255,0.025)] focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
						/>
					</label>
				{/if}

				{#if error}
					<p class="mt-4 text-sm text-destructive" aria-live="polite">{error}</p>
				{/if}

				<div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
					<Button type="button" variant="ghost" disabled={submitting} onclick={close}>
						cancel
					</Button>
					<Button type="submit" variant="destructive" disabled={!canConfirm}>
						{#if submitting}
							working
						{:else}
							{actionLabel}
						{/if}
					</Button>
				</div>
			</form>
		</div>
	</div>
{/if}
