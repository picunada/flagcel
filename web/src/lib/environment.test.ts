import { describe, expect, it } from "vitest";
import { pickSelectedEnvironment } from "./environment-selection";

const environments = [
	{ id: "env-staging", key: "staging", name: "Staging", created_at: "", updated_at: "" },
	{ id: "env-prod", key: "production", name: "Production", created_at: "", updated_at: "" },
	{ id: "env-dev", key: "development", name: "Development", created_at: "", updated_at: "" },
];

describe("pickSelectedEnvironment", () => {
	it("prefers the requested environment id", () => {
		expect(pickSelectedEnvironment(environments, "env-staging", null)?.id).toBe(
			"env-staging",
		);
	});

	it("falls back to the remembered environment id", () => {
		expect(pickSelectedEnvironment(environments, null, "env-dev")?.id).toBe(
			"env-dev",
		);
	});

	it("ignores a missing requested id and uses remembered", () => {
		expect(
			pickSelectedEnvironment(environments, "missing", "env-staging")?.id,
		).toBe("env-staging");
	});

	it("falls back to production when nothing is selected", () => {
		expect(pickSelectedEnvironment(environments, null, null)?.id).toBe(
			"env-prod",
		);
	});

	it("falls back to the first environment when production is absent", () => {
		expect(
			pickSelectedEnvironment(
				[{
                    id: "env-only", key: "only", name: "Only",
                    created_at: "",
                    updated_at: ""
                }],
				null,
				null,
			)?.id,
		).toBe("env-only");
	});
});
