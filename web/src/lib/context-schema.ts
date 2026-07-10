import type { ContextField, ContextType } from "$lib/api";

export const contextTypes: ContextType[] = [
    "string",
    "int",
    "double",
    "bool",
    "timestamp",
    "list",
    "map",
];

export type SchemaParseError = {
    line: number;
    message: string;
};

export type SchemaParseResult = {
    fields: ContextField[];
    errors: SchemaParseError[];
    lineCount: number;
};

export type PayloadResult = {
    status: "valid" | "missing" | "mismatch" | "extra";
    path: string;
    message: string;
    field?: ContextField;
};

export type InferenceResult = {
    fields: ContextField[];
    error: string | null;
    warnings: string[];
};

const pathPattern = /^[\p{L}_][\p{L}\p{N}_]*(\.[\p{L}_][\p{L}\p{N}_]*)*$/u;

export function serializeSchema(fields: ContextField[]) {
    const width = fields.reduce((longest, field) => Math.max(longest, field.path.length), 0);
    return fields
        .map((field) => `${field.path}${" ".repeat(Math.max(1, width - field.path.length + 2))}${field.type}`)
        .join("\n");
}

export function parseSchema(text: string): SchemaParseResult {
    const fields: ContextField[] = [];
    const errors: SchemaParseError[] = [];
    const seen = new Set<string>();
    const lines = text.split("\n");

    lines.forEach((raw, index) => {
        const line = raw.trim();
        if (!line) return;
        const parts = line.split(/\s+/);
        if (parts.length !== 2) {
            errors.push({ line: index + 1, message: "Expected a path and type" });
            return;
        }
        const [path, typeValue] = parts;
        if (!pathPattern.test(path)) {
            errors.push({ line: index + 1, message: `Invalid field path “${path}”` });
            return;
        }
        if (!contextTypes.includes(typeValue as ContextType)) {
            errors.push({ line: index + 1, message: `Unknown type “${typeValue}”` });
            return;
        }
        if (seen.has(path)) {
            errors.push({ line: index + 1, message: `Duplicate field “${path}”` });
            return;
        }
        seen.add(path);
        fields.push({ path, type: typeValue as ContextType });
    });

    return { fields, errors, lineCount: lines.length };
}

export function inferPayload(text: string): InferenceResult {
    if (!text.trim()) return { fields: [], error: null, warnings: [] };
    let value: unknown;
    try {
        value = JSON.parse(text);
    } catch (error) {
        return {
            fields: [],
            error: error instanceof Error ? error.message.split("\n")[0] : "Invalid JSON",
            warnings: [],
        };
    }
    if (!isRecord(value)) {
        return { fields: [], error: "Payload must be a JSON object", warnings: [] };
    }
    const fields: ContextField[] = [];
    const warnings: string[] = [];
    flattenForInference(value, "", fields, warnings);
    return { fields, error: null, warnings };
}

export function validatePayload(text: string, fields: ContextField[]): PayloadResult[] {
    if (!text.trim()) return [];
    let value: unknown;
    try {
        value = JSON.parse(text);
    } catch (error) {
        return [{
            status: "mismatch",
            path: "JSON",
            message: error instanceof Error ? error.message.split("\n")[0] : "Invalid JSON",
        }];
    }
    if (!isRecord(value)) {
        return [{ status: "mismatch", path: "JSON", message: "Expected a JSON object" }];
    }

    const results: PayloadResult[] = fields.map((field) => {
        const fieldValue = getPath(value, field.path);
        if (fieldValue === undefined) {
            return { status: "missing", path: field.path, message: "missing from payload" };
        }
        if (fieldValue === null) {
            return { status: "mismatch", path: field.path, message: "null is not supported" };
        }
        const actual = runtimeType(fieldValue);
        if (!matchesContextType(field.type, fieldValue)) {
            return {
                status: "mismatch",
                path: field.path,
                message: `expected ${field.type}, got ${actual}`,
            };
        }
        return { status: "valid", path: field.path, message: `${field.type} · ok` };
    });

    const inferred = inferPayload(text);
    for (const field of inferred.fields) {
        const covered = fields.some(
            (schemaField) =>
                schemaField.path === field.path ||
                (schemaField.type === "map" && field.path.startsWith(`${schemaField.path}.`)),
        );
        if (!covered) {
            results.push({
                status: "extra",
                path: field.path,
                message: "in payload · not in schema",
                field,
            });
        }
    }
    return results;
}

export function samplePayload(fields: ContextField[]) {
    const root: Record<string, unknown> = {};
    for (const field of fields) setPath(root, field.path, sampleValue(field));
    return JSON.stringify(root, null, 2);
}

function flattenForInference(
    value: Record<string, unknown>,
    prefix: string,
    fields: ContextField[],
    warnings: string[],
) {
    for (const [key, child] of Object.entries(value)) {
        const path = prefix ? `${prefix}.${key}` : key;
        if (child === null) {
            warnings.push(`${path} was ignored because null has no inferable type`);
        } else if (Array.isArray(child)) {
            fields.push({ path, type: "list" });
        } else if (isRecord(child)) {
            if (Object.keys(child).length === 0) fields.push({ path, type: "map" });
            else flattenForInference(child, path, fields, warnings);
        } else if (typeof child === "boolean") {
            fields.push({ path, type: "bool" });
        } else if (typeof child === "number" && Number.isFinite(child)) {
            fields.push({ path, type: Number.isInteger(child) ? "int" : "double" });
        } else if (typeof child === "string") {
            fields.push({ path, type: "string" });
        } else {
            warnings.push(`${path} was ignored because its type is unsupported`);
        }
    }
}

function getPath(value: Record<string, unknown>, path: string) {
    return path.split(".").reduce<unknown>((current, segment) => {
        if (!isRecord(current)) return undefined;
        return current[segment];
    }, value);
}

function setPath(root: Record<string, unknown>, path: string, value: unknown) {
    const parts = path.split(".");
    let current = root;
    parts.slice(0, -1).forEach((part) => {
        if (!isRecord(current[part])) current[part] = {};
        current = current[part] as Record<string, unknown>;
    });
    current[parts.at(-1)!] = value;
}

function sampleValue(field: ContextField): unknown {
    const leaf = field.path.split(".").at(-1);
    if (field.type === "int") return 42;
    if (field.type === "double") return 42.5;
    if (field.type === "bool") return true;
    if (field.type === "timestamp") return "2026-07-10T12:00:00Z";
    if (field.type === "list") return [];
    if (field.type === "map") return {};
    const examples: Record<string, string> = {
        id: "usr_4f2a",
        role: "admin",
        country: "US",
        plan: "pro",
        path: "/checkout",
        region: "us-west-2",
    };
    return examples[leaf ?? ""] ?? "example";
}

function matchesContextType(type: ContextType, value: unknown) {
    if (type === "string") return typeof value === "string";
    if (type === "int") return typeof value === "number" && Number.isInteger(value);
    if (type === "double") return typeof value === "number" && Number.isFinite(value);
    if (type === "bool") return typeof value === "boolean";
    if (type === "timestamp") {
        return typeof value === "string" && !Number.isNaN(Date.parse(value));
    }
    if (type === "list") return Array.isArray(value);
    return isRecord(value);
}

function runtimeType(value: unknown) {
    if (Array.isArray(value)) return "list";
    if (isRecord(value)) return "map";
    if (typeof value === "number") return Number.isInteger(value) ? "int" : "double";
    return typeof value;
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return value !== null && typeof value === "object" && !Array.isArray(value);
}
