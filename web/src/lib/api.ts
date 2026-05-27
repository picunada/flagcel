export type Rollout = {
    percentage: number;
    bucket_by?: string;
};

export type Rule = {
    id: string;
    expression: string;
    rollout: Rollout;
};

export type Flag = {
    key: string;
    enabled: boolean;
    rules: Rule[];
    default_value: boolean;
    context_id?: string | null;
    updated_at: string;
};

export type CreateFlagRequest = {
    key: string;
    enabled?: boolean;
    rules?: CreateRuleRequest[];
    default_value?: boolean;
    context_id?: string | null;
};

export type CreateRuleRequest = {
    expression: string;
    rollout: Rollout;
};

export type UpdateRuleRequest = CreateRuleRequest;

export type EvalTrace = {
    key: string;
    enabled: boolean;
    default_value: boolean;
    value: boolean;
    reason: "disabled" | "matched_rule" | "default_no_match" | "cel_error" | string;
    error?: string;
    matched_rule?: EvalMatchedRule;
    bucket?: EvalBucket;
    rule_results: EvalRuleResult[];
};

export type EvalMatchedRule = {
    id: string;
    index: number;
    expression: string;
};

export type EvalRuleResult = {
    id: string;
    index: number;
    expression: string;
    matched: boolean;
    error?: string;
};

export type EvalBucket = {
    bucket_by: string;
    bucket_value?: string;
    bucket_number?: number;
    percentage: number;
    in_rollout: boolean;
    missing: boolean;
};

export type ContextType =
    | "string"
    | "int"
    | "double"
    | "bool"
    | "timestamp"
    | "list"
    | "map";

export type ContextField = {
    path: string;
    type: ContextType;
};

export type ContextSchema = {
    id: string;
    name: string;
    description?: string;
    fields: ContextField[];
};

export type CreateContextRequest = {
    name: string;
    description?: string;
    fields: ContextField[];
};

export type UpdateContextRequest = CreateContextRequest;

export type APIErrorCode =
    | "FLAG_NOT_FOUND"
    | "RULE_NOT_FOUND"
    | "CONTEXT_NOT_FOUND"
    | "CONTEXT_NAME_TAKEN"
    | "API_KEY_NOT_FOUND"
    | "RULE_VALIDATION_FAILED"
    | "INVALID_REQUEST"
    | "BAD_REQUEST"
    | "AUTH_NOT_CONFIGURED"
    | "UNAUTHORIZED"
    | "FORBIDDEN"
    | "INTERNAL_ERROR";

export type ValidationIssue = {
    code:
        | "parse_error"
        | "unknown_field"
        | "non_bool_expression"
        | "invalid_rollout"
        | "missing_bucket_field"
        | string;
    field: string;
    path?: string;
    message: string;
};

export type User = {
    id: string;
    email: string;
    name?: string;
    admin: boolean;
};

export type AuthMe = {
    auth_enabled: boolean;
    mode?: "oidc" | "password";
    authenticated: boolean;
    user?: User;
};

export type APIKey = {
    id: string;
    name: string;
    prefix: string;
    created_at: string;
    last_used_at?: string;
    revoked_at?: string;
};

export type CreateAPIKeyResponse = APIKey & {
    token: string;
};

export class APIError extends Error {
    code: APIErrorCode;
    status: number;
    details?: ValidationIssue[];
    constructor(
        code: APIErrorCode,
        message: string,
        status: number,
        details?: ValidationIssue[],
    ) {
        super(message);
        this.code = code;
        this.status = status;
        this.details = details;
    }
}

type Envelope<T> = { message: string; data: T };
type ErrorEnvelope = {
    error: { code: APIErrorCode; message: string; details?: ValidationIssue[] };
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`/api/v1${path}`, {
        ...init,
        headers: {
            "Content-Type": "application/json",
            ...(init?.headers ?? {}),
        },
    });

    if (res.status === 204) return undefined as T;

    const text = await res.text();
    const body = text ? JSON.parse(text) : null;

    if (!res.ok) {
        const err = (body as ErrorEnvelope | null)?.error;
        throw new APIError(
            err?.code ?? "INTERNAL_ERROR",
            err?.message ?? `HTTP ${res.status}`,
            res.status,
            err?.details,
        );
    }

    return (body as Envelope<T>).data;
}

export const api = {
    me: () => request<AuthMe>("/auth/me"),
    passwordLogin: async (email: string, password: string) => {
        const res = await fetch("/api/v1/auth/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email, password }),
        });
        const text = await res.text();
        const body = text ? JSON.parse(text) : null;
        if (!res.ok) {
            const err = (body as ErrorEnvelope | null)?.error;
            throw new APIError(
                err?.code ?? "INTERNAL_ERROR",
                err?.message ?? `HTTP ${res.status}`,
                res.status,
                err?.details,
            );
        }
        return (body as Envelope<AuthMe>).data;
    },
    logout: () => request<void>("/auth/logout", { method: "POST" }),

    listFlags: () => request<Flag[]>("/flags"),
    getFlag: (key: string) =>
        request<Flag>(`/flags/${encodeURIComponent(key)}`),
    createFlag: (body: CreateFlagRequest) =>
        request<Flag>("/flags", { method: "POST", body: JSON.stringify(body) }),
    deleteFlag: (key: string) =>
        request<void>(`/flags/${encodeURIComponent(key)}`, {
            method: "DELETE",
        }),

    listRules: (key: string) =>
        request<Rule[]>(`/flags/${encodeURIComponent(key)}/rules`),
    createRule: (key: string, body: CreateRuleRequest) =>
        request<Rule>(`/flags/${encodeURIComponent(key)}/rules`, {
            method: "POST",
            body: JSON.stringify(body),
        }),
    updateRule: (key: string, id: string, body: UpdateRuleRequest) =>
        request<Rule>(`/flags/${encodeURIComponent(key)}/rules/${id}`, {
            method: "PUT",
            body: JSON.stringify(body),
        }),
    deleteRule: (key: string, id: string) =>
        request<void>(`/flags/${encodeURIComponent(key)}/rules/${id}`, {
            method: "DELETE",
        }),
    reorderRules: (key: string, ruleIds: string[]) =>
        request<void>(`/flags/${encodeURIComponent(key)}/rules/reorder`, {
            method: "POST",
            body: JSON.stringify({ rule_ids: ruleIds }),
        }),
    evaluateFlag: (key: string, context: Record<string, unknown>) =>
        request<EvalTrace>(`/flags/${encodeURIComponent(key)}/evaluate`, {
            method: "POST",
            body: JSON.stringify({ context }),
        }),

    listContexts: () => request<ContextSchema[]>("/contexts"),
    getContext: (id: string) =>
        request<ContextSchema>(`/contexts/${encodeURIComponent(id)}`),
    createContext: (body: CreateContextRequest) =>
        request<ContextSchema>("/contexts", {
            method: "POST",
            body: JSON.stringify(body),
        }),
    updateContext: (id: string, body: UpdateContextRequest) =>
        request<ContextSchema>(`/contexts/${encodeURIComponent(id)}`, {
            method: "PUT",
            body: JSON.stringify(body),
        }),
    deleteContext: (id: string) =>
        request<void>(`/contexts/${encodeURIComponent(id)}`, {
            method: "DELETE",
        }),

    listAPIKeys: () => request<APIKey[]>("/api-keys"),
    createAPIKey: (name: string) =>
        request<CreateAPIKeyResponse>("/api-keys", {
            method: "POST",
            body: JSON.stringify({ name }),
        }),
    revokeAPIKey: (id: string) =>
        request<void>(`/api-keys/${encodeURIComponent(id)}`, {
            method: "DELETE",
        }),
};
