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
  created_at?: string;
  updated_at?: string;
  created_by?: string | null;
  deleted_by?: string | null;
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
  id?: string;
  name?: string;
  description?: string;
  fields: ContextField[];
  created_at?: string;
  updated_at?: string;
  created_by?: string | null;
  deleted_by?: string | null;
};

export type FlagDefinition = {
  key: string;
  description?: string;
  type: ValueType;
  enabled: boolean;
  rules: Rule[];
  default_value: FlagValue;
  context_id?: string | null;
  context_schema?: ContextSchema;
  created_at?: string;
  updated_at?: string;
  created_by?: string | null;
  deleted_by?: string | null;
};

export type Definitions = {
  flags: FlagDefinition[];
  contexts?: ContextSchema[];
};

export type EvaluationResult = {
  key?: string;
  value: unknown;
  value_type: ValueType;
  reason?: string;
  variant?: string;
  error?: string;
};

export type EvaluationEnvelope = {
  message: string;
  data: EvaluationResult;
};

export type EvaluationContextJSON = Record<string, unknown>;
