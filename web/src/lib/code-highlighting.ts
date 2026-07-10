const schemaTypes = new Set(["string", "int", "double", "bool", "timestamp", "list", "map"]);

export function formatJSON(text: string) {
    if (!text.trim()) return text;
    try {
        return JSON.stringify(JSON.parse(text), null, 2);
    } catch {
        return text;
    }
}

export function highlightCode(text: string, syntax: "schema" | "json") {
    return syntax === "schema" ? highlightSchema(text) : highlightJSON(text);
}

function highlightSchema(text: string) {
    return text
        .split("\n")
        .map((line) => {
            const match = line.match(/^(\s*)(\S+)(\s+)(\S+)(.*)$/);
            if (!match) return escapeHTML(line);
            const [, before, path, gap, type, after] = match;
            const typeColor = schemaTypes.has(type)
                ? "var(--color-valid)"
                : "var(--color-destructive)";
            return (
                escapeHTML(before) +
                token(path, "var(--color-foreground)") +
                escapeHTML(gap) +
                token(type, typeColor) +
                token(after, "var(--color-muted-foreground)")
            );
        })
        .join("\n");
}

function highlightJSON(text: string) {
    const pattern = /"(?:\\.|[^"\\])*"(?=\s*:)|"(?:\\.|[^"\\])*"|\b(?:true|false|null)\b|-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b/g;
    let output = "";
    let cursor = 0;
    for (const match of text.matchAll(pattern)) {
        const index = match.index ?? 0;
        const value = match[0];
        output += escapeHTML(text.slice(cursor, index));
        if (value.startsWith('"')) {
            const isKey = /^\s*:/.test(text.slice(index + value.length));
            output += token(
                value,
                isKey ? "var(--color-app-accent-muted)" : "var(--color-foreground-soft)",
            );
        } else if (value === "true" || value === "false") {
            output += token(value, "var(--color-valid-muted)");
        } else if (value === "null") {
            output += token(value, "var(--color-destructive)");
        } else {
            output += token(value, "var(--color-cel-warning)");
        }
        cursor = index + value.length;
    }
    return output + escapeHTML(text.slice(cursor));
}

function token(value: string, color: string) {
    return `<span style="color:${color}">${escapeHTML(value)}</span>`;
}

function escapeHTML(value: string) {
    return value
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}
