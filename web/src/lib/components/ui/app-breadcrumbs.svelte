<script lang="ts">
	import * as Breadcrumb from '$lib/components/ui/breadcrumb';
	import { cn } from '$lib/utils';

	export type BreadcrumbCrumb = {
		label: string;
		href?: string;
	};

	type Props = {
		items: BreadcrumbCrumb[];
		class?: string;
	};

	let { items, class: className }: Props = $props();
</script>

{#if items.length > 0}
	<Breadcrumb.Root class={cn(className)}>
		<Breadcrumb.List>
			{#each items as item, index (item.href ?? item.label)}
				{#if index > 0}
					<Breadcrumb.Separator />
				{/if}
				<Breadcrumb.Item>
					{#if item.href && index < items.length - 1}
						<Breadcrumb.Link href={item.href}>{item.label}</Breadcrumb.Link>
					{:else}
						<Breadcrumb.Page class={index < items.length - 1 ? 'text-muted-foreground' : undefined}>
							{item.label}
						</Breadcrumb.Page>
					{/if}
				</Breadcrumb.Item>
			{/each}
		</Breadcrumb.List>
	</Breadcrumb.Root>
{/if}
