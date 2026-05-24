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

export type ContextType =
	| 'string'
	| 'int'
	| 'double'
	| 'bool'
	| 'timestamp'
	| 'list'
	| 'map';

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
	| 'FLAG_NOT_FOUND'
	| 'RULE_NOT_FOUND'
	| 'CONTEXT_NOT_FOUND'
	| 'CONTEXT_NAME_TAKEN'
	| 'INVALID_REQUEST'
	| 'BAD_REQUEST'
	| 'INTERNAL_ERROR';

export class APIError extends Error {
	code: APIErrorCode;
	status: number;
	constructor(code: APIErrorCode, message: string, status: number) {
		super(message);
		this.code = code;
		this.status = status;
	}
}

type Envelope<T> = { message: string; data: T };
type ErrorEnvelope = { error: { code: APIErrorCode; message: string } };

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const res = await fetch(`/api/v1${path}`, {
		...init,
		headers: {
			'Content-Type': 'application/json',
			...(init?.headers ?? {})
		}
	});

	if (res.status === 204) return undefined as T;

	const text = await res.text();
	const body = text ? JSON.parse(text) : null;

	if (!res.ok) {
		const err = (body as ErrorEnvelope | null)?.error;
		throw new APIError(
			err?.code ?? 'INTERNAL_ERROR',
			err?.message ?? `HTTP ${res.status}`,
			res.status
		);
	}

	return (body as Envelope<T>).data;
}

export const api = {
	listFlags: () => request<Flag[]>('/flags'),
	getFlag: (key: string) => request<Flag>(`/flags/${encodeURIComponent(key)}`),
	createFlag: (body: CreateFlagRequest) =>
		request<Flag>('/flags', { method: 'POST', body: JSON.stringify(body) }),
	deleteFlag: (key: string) =>
		request<void>(`/flags/${encodeURIComponent(key)}`, { method: 'DELETE' }),

	listRules: (key: string) => request<Rule[]>(`/flags/${encodeURIComponent(key)}/rules`),
	createRule: (key: string, body: CreateRuleRequest) =>
		request<Rule>(`/flags/${encodeURIComponent(key)}/rules`, {
			method: 'POST',
			body: JSON.stringify(body)
		}),
	updateRule: (key: string, id: string, body: UpdateRuleRequest) =>
		request<Rule>(`/flags/${encodeURIComponent(key)}/rules/${id}`, {
			method: 'PUT',
			body: JSON.stringify(body)
		}),
	deleteRule: (key: string, id: string) =>
		request<void>(`/flags/${encodeURIComponent(key)}/rules/${id}`, { method: 'DELETE' }),
	reorderRules: (key: string, ruleIds: string[]) =>
		request<void>(`/flags/${encodeURIComponent(key)}/rules/reorder`, {
			method: 'POST',
			body: JSON.stringify({ rule_ids: ruleIds })
		}),

	listContexts: () => request<ContextSchema[]>('/contexts'),
	getContext: (id: string) => request<ContextSchema>(`/contexts/${encodeURIComponent(id)}`),
	createContext: (body: CreateContextRequest) =>
		request<ContextSchema>('/contexts', { method: 'POST', body: JSON.stringify(body) }),
	updateContext: (id: string, body: UpdateContextRequest) =>
		request<ContextSchema>(`/contexts/${encodeURIComponent(id)}`, {
			method: 'PUT',
			body: JSON.stringify(body)
		}),
	deleteContext: (id: string) =>
		request<void>(`/contexts/${encodeURIComponent(id)}`, { method: 'DELETE' })
};
