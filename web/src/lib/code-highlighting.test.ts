import { describe, expect, it } from "vitest";
import { formatJSON, highlightCode } from "./code-highlighting";

describe("code highlighting", () => {
	it("pretty-prints valid JSON", () => {
		expect(formatJSON('{"user":{"id":"u1"},"active":true}')).toBe(`{
  "user": {
    "id": "u1"
  },
  "active": true
}`);
	});

	it("returns invalid JSON unchanged", () => {
		expect(formatJSON("{")).toBe("{");
	});

	it("highlights schema and JSON tokens", () => {
		expect(highlightCode("user.id string", "schema")).toMatch(
			/color:var\(--color-valid\)/,
		);
		expect(highlightCode('{"count":42,"ok":true}', "json")).toMatch(
			/color:var\(--color-cel-warning\)/,
		);
	});

	it("escapes HTML in highlighted JSON", () => {
		expect(highlightCode('{"value":"<script>"}', "json")).not.toMatch(
			/<script>/,
		);
	});
});
