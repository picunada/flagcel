import { describe, expect, it } from "vitest";
import type { ContextField } from "$lib/api";
import {
	inferPayload,
	parseSchema,
	samplePayload,
	serializeSchema,
	validatePayload,
} from "./context-schema";

const fields: ContextField[] = [
	{ path: "user.id", type: "string" },
	{ path: "user.age", type: "int" },
	{ path: "score", type: "double" },
];

describe("context schema", () => {
	it("round-trips schema serialization", () => {
		expect(parseSchema(serializeSchema(fields)).fields).toEqual(fields);
	});

	it("reports schema parse errors", () => {
		expect(
			parseSchema("bad-path string\nuser.id nope\nuser.id string\nuser.id string")
				.errors,
		).toHaveLength(3);
	});

	it("infers fields from a payload", () => {
		const inferred = inferPayload(
			JSON.stringify({
				user: { id: "u1", age: 42, ratio: 0.5, active: true },
				tags: [],
				metadata: {},
				missing: null,
			}),
		);
		expect(inferred.fields).toEqual([
			{ path: "user.id", type: "string" },
			{ path: "user.age", type: "int" },
			{ path: "user.ratio", type: "double" },
			{ path: "user.active", type: "bool" },
			{ path: "tags", type: "list" },
			{ path: "metadata", type: "map" },
		]);
		expect(inferred.warnings).toHaveLength(1);
	});

	it("validates payloads against a schema", () => {
		const results = validatePayload(
			JSON.stringify({ user: { id: "u1", age: 42 }, score: 4, extra: true }),
			fields,
		);
		expect(results.map(({ status, path }) => ({ status, path }))).toEqual([
			{ status: "valid", path: "user.id" },
			{ status: "valid", path: "user.age" },
			{ status: "valid", path: "score" },
			{ status: "extra", path: "extra" },
		]);
	});

	it("builds a sample payload", () => {
		const sample = JSON.parse(samplePayload(fields));
		expect(sample.user.id).toBe("usr_4f2a");
		expect(sample.user.age).toBe(42);
	});
});
