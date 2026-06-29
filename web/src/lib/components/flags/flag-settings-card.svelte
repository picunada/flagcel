<script lang="ts">
    import type { Flag, FlagValue } from "$lib/api";
    import ContextPicker from "$lib/components/context-picker.svelte";
    import Badge from "$lib/components/ui/badge.svelte";
    import BoolToggle from "$lib/components/ui/bool-toggle.svelte";
    import Card from "$lib/components/ui/card.svelte";
    import ValueEditor from "$lib/components/value-editor.svelte";

    type FlagPatch = Partial<
        Pick<Flag, "enabled" | "default_value" | "context_id">
    >;

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
        updateDefaultValue,
    }: Props = $props();
</script>

<Card class="motion-panel divide-y divide-border/60">
    <div class="flex items-center justify-between gap-4 p-5">
        <div class="space-y-1">
            <p class="text-sm">type</p>
            <p class="text-xs text-muted-foreground">
                value shape returned by evaluation
            </p>
        </div>
        <Badge variant="muted">{flag.type}</Badge>
    </div>
    <div class="flex items-center justify-between gap-4 p-5">
        <div class="space-y-1">
            <p class="text-sm">enabled</p>
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
    <div class="flex flex-wrap items-start justify-between gap-4 p-5">
        <div class="space-y-1">
            <p class="text-sm">default value</p>
            <p class="text-xs text-muted-foreground">
                returned when no rule matches
            </p>
        </div>
        <div class="min-w-48 max-w-full flex-1 sm:flex-none sm:basis-80">
            <ValueEditor
                type={flag.type}
                value={flag.default_value}
                id="default-value"
                align="end"
                disabled={saving}
                onchange={updateDefaultValue}
            />
        </div>
    </div>
    <div class="flex flex-wrap items-center justify-between gap-4 p-5">
        <div class="space-y-1">
            <p class="text-sm">context</p>
            <p class="text-xs text-muted-foreground">
                selects the evaluation shape used for autocomplete in rules
            </p>
        </div>
        <div class="min-w-48">
            <ContextPicker
                value={flag.context_id ?? null}
                disabled={saving}
                onchange={(v) => patch({ context_id: v })}
            />
        </div>
    </div>
    <div
        class="flex flex-wrap items-center justify-between gap-x-6 gap-y-1 p-5 text-[0.7rem] uppercase tracking-[0.12em] text-muted-foreground"
    >
        <span>
            created {formatTimestamp(flag.created_at)}{#if createdBy}
                · <span class="font-mono lowercase text-foreground"
                    >{createdBy}</span
                >{/if}
        </span>
        <span>
            updated {formatTimestamp(flag.updated_at)}{#if updatedBy}
                · <span class="font-mono lowercase text-foreground"
                    >{updatedBy}</span
                >{/if}
        </span>
    </div>
</Card>
