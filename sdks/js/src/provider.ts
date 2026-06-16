import {
  ErrorCode,
  OpenFeatureEventEmitter,
  StandardResolutionReasons,
  type EvaluationContext,
  type JsonValue,
  type Logger,
  type Provider,
  type ResolutionDetails,
} from "@openfeature/server-sdk";

import { EvalClient, EvalClientError, type Fetch } from "./eval-client.js";
import type { EvaluationContextJSON, EvaluationResult, ValueType } from "./types.js";

export type FlagcelProviderOptions = {
  endpoint: string | URL;
  apiKey?: string;
  fetch?: Fetch;
};

export class FlagcelProvider implements Provider {
  readonly runsOn = "server";
  readonly metadata = { name: "flagcel" };
  readonly events = new OpenFeatureEventEmitter();

  readonly #client: EvalClient;

  constructor(options: FlagcelProviderOptions) {
    this.#client = new EvalClient(options.endpoint, options.apiKey, options.fetch);
  }

  async initialize(): Promise<void> {
    return;
  }

  async onClose(): Promise<void> {
    return;
  }

  async resolveBooleanEvaluation(
    flagKey: string,
    defaultValue: boolean,
    context: EvaluationContext,
    _logger: Logger,
  ): Promise<ResolutionDetails<boolean>> {
    const result = await this.#evaluate(flagKey, context);
    if (!result.ok) {
      return errorDetails(defaultValue, result.errorCode, result.errorMessage);
    }
    if (result.value.value_type !== "boolean" || typeof result.value.value !== "boolean") {
      return typeMismatchDetails(defaultValue, flagKey, "boolean", result.value);
    }
    return successDetails(result.value.value, result.value);
  }

  async resolveStringEvaluation(
    flagKey: string,
    defaultValue: string,
    context: EvaluationContext,
    _logger: Logger,
  ): Promise<ResolutionDetails<string>> {
    const result = await this.#evaluate(flagKey, context);
    if (!result.ok) {
      return errorDetails(defaultValue, result.errorCode, result.errorMessage);
    }
    if (result.value.value_type !== "string" || typeof result.value.value !== "string") {
      return typeMismatchDetails(defaultValue, flagKey, "string", result.value);
    }
    return successDetails(result.value.value, result.value);
  }

  async resolveNumberEvaluation(
    flagKey: string,
    defaultValue: number,
    context: EvaluationContext,
    _logger: Logger,
  ): Promise<ResolutionDetails<number>> {
    const result = await this.#evaluate(flagKey, context);
    if (!result.ok) {
      return errorDetails(defaultValue, result.errorCode, result.errorMessage);
    }
    if (result.value.value_type !== "number" || typeof result.value.value !== "number") {
      return typeMismatchDetails(defaultValue, flagKey, "number", result.value);
    }
    return successDetails(result.value.value, result.value);
  }

  async resolveObjectEvaluation<T extends JsonValue>(
    flagKey: string,
    defaultValue: T,
    context: EvaluationContext,
    _logger: Logger,
  ): Promise<ResolutionDetails<T>> {
    const result = await this.#evaluate(flagKey, context);
    if (!result.ok) {
      return errorDetails(defaultValue, result.errorCode, result.errorMessage);
    }
    if (result.value.value_type !== "json") {
      return typeMismatchDetails(defaultValue, flagKey, "json", result.value);
    }
    return successDetails(result.value.value as T, result.value);
  }

  async #evaluate(
    flagKey: string,
    context: EvaluationContext,
  ): Promise<
    | { ok: true; value: EvaluationResult }
    | { ok: false; errorCode: ErrorCode; errorMessage: string }
  > {
    let result: EvaluationResult;
    try {
      result = await this.#client.evaluate(flagKey, contextToJSON(context));
    } catch (error: unknown) {
      return {
        ok: false,
        errorCode:
          error instanceof EvalClientError && error.status === 404
            ? ErrorCode.FLAG_NOT_FOUND
            : ErrorCode.GENERAL,
        errorMessage: normalizeError(error).message,
      };
    }
    if (result.error) {
      return {
        ok: false,
        errorCode: result.reason === "not_found" ? ErrorCode.FLAG_NOT_FOUND : ErrorCode.GENERAL,
        errorMessage: result.error,
      };
    }
    return { ok: true, value: result };
  }
}

function successDetails<T extends JsonValue | boolean | number | string>(
  value: T,
  result: EvaluationResult,
): ResolutionDetails<T> {
  return {
    value,
    reason: mapReason(result.reason ?? ""),
    variant: result.variant,
    flagMetadata: {
      flagcelReason: result.reason ?? "",
      valueType: result.value_type,
    },
  };
}

function errorDetails<T>(
  defaultValue: T,
  errorCode: ErrorCode,
  errorMessage: string,
): ResolutionDetails<T> {
  return {
    value: defaultValue,
    reason: StandardResolutionReasons.ERROR,
    errorCode,
    errorMessage,
  };
}

function typeMismatchDetails<T>(
  defaultValue: T,
  flagKey: string,
  expected: ValueType,
  result: EvaluationResult,
): ResolutionDetails<T> {
  return errorDetails(
    defaultValue,
    ErrorCode.TYPE_MISMATCH,
    `${flagKey}: expected ${expected}, got ${result.value_type}`,
  );
}

function mapReason(reason: string): ResolutionDetails<unknown>["reason"] {
  switch (reason) {
    case "matched_rule":
      return StandardResolutionReasons.TARGETING_MATCH;
    case "default_no_match":
      return StandardResolutionReasons.DEFAULT;
    case "disabled":
      return StandardResolutionReasons.DISABLED;
    case "not_found":
    case "cel_error":
    case "error":
      return StandardResolutionReasons.ERROR;
    default:
      return reason as ResolutionDetails<unknown>["reason"];
  }
}

function contextToJSON(context: EvaluationContext | undefined): EvaluationContextJSON {
  if (!context) {
    return {};
  }
  return Object.fromEntries(
    Object.entries(context).filter(([, value]) => value !== undefined),
  ) as EvaluationContextJSON;
}

function normalizeError(error: unknown): Error {
  return error instanceof Error ? error : new Error(String(error));
}
