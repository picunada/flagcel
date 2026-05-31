export type ValueType = "boolean" | "string" | "number" | "json";
export type FlagValue = boolean | string | number | null | Record<string, unknown> | unknown[];

export type Rollout = {
    percentage: number;
    bucket_by?: string;
};

export type Rule = {
    id: string;
    description?: string;
    expression: string;
    rollout: Rollout;
    value: FlagValue;
    created_at: string;
    updated_at: string;
    created_by?: string | null;
    deleted_by?: string | null;
};

export type Flag = {
    key: string;
    description?: string;
    type: ValueType;
    enabled: boolean;
    rules: Rule[];
    default_value: FlagValue;
    context_id?: string | null;
    created_at: string;
    updated_at: string;
    created_by?: string | null;
    deleted_by?: string | null;
};

export type CreateFlagRequest = {
    key: string;
    description?: string;
    type?: ValueType;
    enabled?: boolean;
    rules?: CreateRuleRequest[];
    default_value?: FlagValue;
    context_id?: string | null;
};

export type CreateRuleRequest = {
    description?: string;
    expression: string;
    rollout: Rollout;
    value?: FlagValue;
};

export type UpdateRuleRequest = CreateRuleRequest;

export type EvalTrace = {
    key: string;
    value_type: ValueType;
    enabled: boolean;
    default_value: FlagValue;
    value: FlagValue;
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
    value: FlagValue;
};

export type EvalRuleResult = {
    id: string;
    index: number;
    expression: string;
    value: FlagValue;
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
    created_at: string;
    updated_at: string;
    created_by?: string | null;
    deleted_by?: string | null;
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
        | "invalid_value_type"
        | "invalid_value"
        | string;
    field: string;
    path?: string;
    message: string;
};

export type User = {
    id: string;
    email: string;
    name?: string;
    description?: string;
    admin: boolean;
    created_at: string;
    updated_at: string;
    created_by?: string | null;
    deleted_by?: string | null;
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
    description?: string;
    prefix: string;
    created_at: string;
    updated_at: string;
    last_used_at?: string;
    revoked_at?: string;
    created_by?: string | null;
    deleted_by?: string | null;
};

export type CreateAPIKeyRequest = {
    name: string;
    description?: string;
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

type Fetch = typeof globalThis.fetch;

const offlineMessage =
    "Backend server is not responding. Check that Flagcel is running, then retry.";

async function readResponseBody(res: Response) {
    const text = await res.text();
    if (!text) return null;

    try {
        return JSON.parse(text) as unknown;
    } catch {
        if (!res.ok) return null;
        throw new APIError(
            "INTERNAL_ERROR",
            "Backend returned an invalid response.",
            res.status,
        );
    }
}

/**
 * Build an API client bound to a specific `fetch`. Inside `load` functions pass
 * the `fetch` SvelteKit provides; elsewhere `api` (bound to the global fetch) is fine.
 */
export function createApi(fetchFn: Fetch = fetch) {
    async function request<T>(path: string, init?: RequestInit): Promise<T> {
        let res: Response;
        try {
            res = await fetchFn(`/api/v1${path}`, {
                ...init,
                headers: {
                    "Content-Type": "application/json",
                    ...(init?.headers ?? {}),
                },
            });
        } catch {
            throw new APIError("INTERNAL_ERROR", offlineMessage, 0);
        }

        if (res.status === 204) return undefined as T;

        const body = await readResponseBody(res);

        if (!res.ok) {
            const err = (body as ErrorEnvelope | null)?.error;
            throw new APIError(
                err?.code ?? "INTERNAL_ERROR",
                err?.message ?? (res.status >= 500 ? offlineMessage : `HTTP ${res.status}`),
                res.status,
                err?.details,
            );
        }

        return (body as Envelope<T>).data;
    }

    return {
        me: () => request<AuthMe>("/auth/me"),
        passwordLogin: async (email: string, password: string) => {
            let res: Response;
            try {
                res = await fetchFn("/api/v1/auth/login", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ email, password }),
                });
            } catch {
                throw new APIError("INTERNAL_ERROR", offlineMessage, 0);
            }
            const body = await readResponseBody(res);
            if (!res.ok) {
                const err = (body as ErrorEnvelope | null)?.error;
                throw new APIError(
                    err?.code ?? "INTERNAL_ERROR",
                    err?.message ?? (res.status >= 500 ? offlineMessage : `HTTP ${res.status}`),
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
        createAPIKey: (name: string, description = "") =>
            request<CreateAPIKeyResponse>("/api-keys", {
                method: "POST",
                body: JSON.stringify({ name, description }),
            }),
        revokeAPIKey: (id: string) =>
            request<void>(`/api-keys/${encodeURIComponent(id)}`, {
                method: "DELETE",
            }),
    };
}

export type Api = ReturnType<typeof createApi>;

export const api = createApi();
