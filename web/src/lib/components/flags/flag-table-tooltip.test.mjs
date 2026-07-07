import assert from "node:assert/strict";

const { flagEnabledTooltipLabel } = await import("./flag-table-tooltip.ts");

assert.equal(
    flagEnabledTooltipLabel({ enabled: true }, false),
    "Disable flag",
);
assert.equal(
    flagEnabledTooltipLabel({ enabled: false }, false),
    "Enable flag",
);
assert.equal(
    flagEnabledTooltipLabel({ enabled: true }, true),
    "Saving flag state",
);
