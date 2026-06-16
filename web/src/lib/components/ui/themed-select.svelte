<script lang="ts">
	import * as Select from '$lib/components/ui/select';
	import { cn } from '$lib/utils';

	export type SelectOption = {
		value: string;
		label: string;
		detail?: string;
	};

	type Props = {
		value: string;
		options: SelectOption[];
		onchange?: (value: string) => void;
		placeholder?: string;
		disabled?: boolean;
		class?: string;
		buttonClass?: string;
		align?: 'left' | 'right';
		mono?: boolean;
	};

	let {
		value,
		options,
		onchange,
		placeholder = 'select',
		disabled = false,
		class: className,
		buttonClass,
		align = 'left',
		mono = false
	}: Props = $props();

	const selected = $derived(options.find((option) => option.value === value));
</script>

<Select.Root
	type="single"
	{value}
	{disabled}
	onValueChange={(next) => onchange?.(next)}
>
	<Select.Trigger class={cn(mono && 'font-mono', buttonClass, className)}>
		<span class={cn('truncate', !selected && 'text-muted-foreground')}>
			{selected?.label ?? placeholder}
		</span>
	</Select.Trigger>
	<Select.Content align={align === 'right' ? 'end' : 'start'}>
		{#each options as option (option.value)}
			<Select.Item value={option.value} label={option.label}>
				<span class={cn('block truncate', mono && 'font-mono')}>{option.label}</span>
				{#if option.detail}
					<span class="block truncate text-[0.65rem] text-muted-foreground">{option.detail}</span>
				{/if}
			</Select.Item>
		{/each}
	</Select.Content>
</Select.Root>
