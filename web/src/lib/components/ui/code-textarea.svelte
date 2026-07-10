<script lang="ts">
    import type { HTMLTextareaAttributes } from "svelte/elements";
    import { highlightCode } from "$lib/code-highlighting";
    import { cn } from "$lib/utils";

    type Props = {
        value?: string;
        syntax: "schema" | "json";
        class?: string;
        editorClass?: string;
    } & Omit<HTMLTextareaAttributes, "class" | "value">;

    let {
        value = $bindable(""),
        syntax,
        class: className,
        editorClass,
        ...rest
    }: Props = $props();

    const highlighted = $derived(highlightCode(value, syntax));
    const editorStyles =
        "box-border m-0 whitespace-pre-wrap break-words px-4 py-3 font-mono text-xs leading-6 [overflow-wrap:anywhere]";
</script>

<div
    class={cn(
        "relative min-w-0 overflow-hidden transition-colors focus-within:bg-surface-subtle",
        className,
    )}
>
    <pre
        class={cn(
            editorStyles,
            "pointer-events-none min-h-[inherit] text-muted-foreground",
            editorClass,
        )}
        aria-hidden="true"
    ><code>{@html highlighted}{value.endsWith("\n") ? " " : ""}</code></pre>
    <textarea
        bind:value
        class={cn(
            editorStyles,
            "absolute inset-0 h-full w-full resize-none overflow-hidden border-0 bg-transparent text-transparent caret-foreground outline-none",
            "placeholder:text-muted-foreground/60 selection:bg-app-accent/25",
            editorClass,
        )}
        {...rest}
    ></textarea>
</div>
