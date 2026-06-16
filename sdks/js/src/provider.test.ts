import { ErrorCode, StandardResolutionReasons, type EvaluationContext } from "@openfeature/server-sdk";
import { describe, expect, it, vi } from "vitest";

import { FlagcelProvider } from "./provider.js";
import type { EvaluationResult } from "./types.js";

const logger = {
  error: vi.fn(),
  warn: vi.fn(),
  info: vi.fn(),
  debug: vi.fn(),
};

describe("FlagcelProvider", () => {
  it("sends authenticated server-side evaluation requests", async () => {
    const fetchMock = vi.fn(async (input: URL | RequestInfo, init?: RequestInit) => {
      expect(String(input)).toBe("https://flagcel.test/api/v1/eval/enabled");
      const headers = new Headers(init?.headers);
      expect(init?.method).toBe("POST");
      expect(headers.get("Authorization")).toBe("Bearer secret");
      expect(headers.get("Content-Type")).toBe("application/json");
      expect(JSON.parse(String(init?.body))).toEqual({
        context: { targetingKey: "user-123" },
      });
      return jsonResponse({
        key: "enabled",
        value: true,
        value_type: "boolean",
        reason: "matched_rule",
        variant: "targeted",
      });
    });

    const provider = new FlagcelProvider({
      endpoint: "https://flagcel.test/api/v1",
      apiKey: "secret",
      fetch: fetchMock,
    });
    await provider.initialize();

    const detail = await provider.resolveBooleanEvaluation(
      "enabled",
      false,
      { targetingKey: "user-123" },
      logger,
    );

    expect(detail).toMatchObject({
      value: true,
      reason: StandardResolutionReasons.TARGETING_MATCH,
      variant: "targeted",
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("resolves string, number, and object values from server responses", async () => {
    const provider = new FlagcelProvider({
      endpoint: "https://flagcel.test/api/v1",
      fetch: async (input: URL | RequestInfo) => {
        const key = String(input).split("/").pop();
        if (key === "title") {
          return jsonResponse({ key, value: "hello", value_type: "string", reason: "default_no_match" });
        }
        if (key === "count") {
          return jsonResponse({ key, value: 42, value_type: "number", reason: "default_no_match" });
        }
        return jsonResponse({ key, value: { plan: "pro" }, value_type: "json", reason: "default_no_match" });
      },
    });

    await expect(provider.resolveStringEvaluation("title", "", {}, logger)).resolves.toMatchObject({
      value: "hello",
      reason: StandardResolutionReasons.DEFAULT,
    });
    await expect(provider.resolveNumberEvaluation("count", 0, {}, logger)).resolves.toMatchObject({
      value: 42,
      reason: StandardResolutionReasons.DEFAULT,
    });
    await expect(provider.resolveObjectEvaluation("payload", {}, {}, logger)).resolves.toMatchObject({
      value: { plan: "pro" },
      reason: StandardResolutionReasons.DEFAULT,
    });
  });

  it("returns default on type mismatch", async () => {
    const provider = initializedProvider({ key: "title", value: "hello", value_type: "string" });

    const detail = await provider.resolveBooleanEvaluation("title", true, {}, logger);

    expect(detail).toMatchObject({
      value: true,
      reason: StandardResolutionReasons.ERROR,
      errorCode: ErrorCode.TYPE_MISMATCH,
    });
  });

  it("returns default on not found responses", async () => {
    const provider = new FlagcelProvider({
      endpoint: "https://flagcel.test/api/v1",
      fetch: async () =>
        new Response(JSON.stringify({ error: { code: "FLAG_NOT_FOUND", message: "Flag not found" } }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        }),
    });

    const detail = await provider.resolveBooleanEvaluation("missing", true, {}, logger);

    expect(detail).toMatchObject({
      value: true,
      reason: StandardResolutionReasons.ERROR,
      errorCode: ErrorCode.FLAG_NOT_FOUND,
      errorMessage: "Flag not found",
    });
  });

  it("returns default on network failures", async () => {
    const provider = new FlagcelProvider({
      endpoint: "https://flagcel.test/api/v1",
      fetch: async () => {
        throw new Error("connection refused");
      },
    });

    const detail = await provider.resolveBooleanEvaluation("enabled", false, {}, logger);

    expect(detail).toMatchObject({
      value: false,
      reason: StandardResolutionReasons.ERROR,
      errorCode: ErrorCode.GENERAL,
      errorMessage: "connection refused",
    });
  });
});

function initializedProvider(result: EvaluationResult): FlagcelProvider {
  return new FlagcelProvider({
    endpoint: "https://flagcel.test/api/v1",
    fetch: async () => jsonResponse(result),
  });
}

function jsonResponse(result: EvaluationResult): Response {
  return new Response(JSON.stringify({ message: "success", data: result }), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
