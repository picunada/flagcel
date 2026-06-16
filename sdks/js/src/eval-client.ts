import type { EvaluationContextJSON, EvaluationEnvelope, EvaluationResult } from "./types.js";

export type Fetch = typeof globalThis.fetch;

export class EvalClient {
  readonly #endpoint: URL;
  readonly #apiKey: string;
  readonly #fetch: Fetch;

  constructor(endpoint: string | URL, apiKey = "", fetchFn: Fetch = fetch) {
    this.#endpoint = new URL(endpoint);
    this.#apiKey = apiKey;
    this.#fetch = fetchFn;
  }

  async evaluate(flagKey: string, context: EvaluationContextJSON): Promise<EvaluationResult> {
    const url = new URL(
      joinPath(this.#endpoint.pathname, `/eval/${encodeURIComponent(flagKey)}`),
      this.#endpoint,
    );
    const headers = new Headers({ Accept: "application/json", "Content-Type": "application/json" });
    if (this.#apiKey) {
      headers.set("Authorization", `Bearer ${this.#apiKey}`);
    }

    const response = await this.#fetch(url, {
      method: "POST",
      headers,
      body: JSON.stringify({ context }),
    });
    if (!response.ok) {
      const message = await readErrorMessage(response);
      throw new EvalClientError(message, response.status);
    }

    const envelope = (await response.json()) as EvaluationEnvelope;
    return envelope.data;
  }
}

export class EvalClientError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "EvalClientError";
  }
}

async function readErrorMessage(response: Response): Promise<string> {
  const fallback = `evaluate flag failed: ${response.status}`;
  const body = await response.text().catch(() => "");
  if (!body.trim()) {
    return fallback;
  }
  try {
    const parsed = JSON.parse(body) as { error?: { message?: string; code?: string } };
    return parsed.error?.message || parsed.error?.code || `${fallback} ${body.trim()}`;
  } catch {
    return `${fallback} ${body.trim()}`;
  }
}

function joinPath(basePath: string, path: string): string {
  const trimmed = basePath.replace(/\/+$/, "");
  return trimmed === "" ? path : `${trimmed}${path}`;
}
