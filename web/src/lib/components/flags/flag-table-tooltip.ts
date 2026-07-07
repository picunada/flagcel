import type { Flag } from "$lib/api";

export function flagEnabledTooltipLabel(
    flag: Pick<Flag, "enabled">,
    saving: boolean,
) {
    if (saving) return "Saving flag state";
    return flag.enabled ? "Disable flag" : "Enable flag";
}
