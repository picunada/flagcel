import { describe, expect, it } from "vitest";
import { flagEnabledTooltipLabel } from "./flag-table-tooltip";

describe("flagEnabledTooltipLabel", () => {
	it("labels disable/enable actions", () => {
		expect(flagEnabledTooltipLabel({ enabled: true }, false)).toBe(
			"Disable flag",
		);
		expect(flagEnabledTooltipLabel({ enabled: false }, false)).toBe(
			"Enable flag",
		);
	});

	it("labels the saving state", () => {
		expect(flagEnabledTooltipLabel({ enabled: true }, true)).toBe(
			"Saving flag state",
		);
	});
});
