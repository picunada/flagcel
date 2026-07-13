import { createApi } from "$lib/api";
import { resolveSelectedEnvironment } from "$lib/environment";
import { runLoad } from "$lib/load";
import { normalizeUsage, parseUsageRange } from "$lib/usage-analytics";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params, url, fetch }) => {
    const api = createApi(fetch);
    return runLoad(async () => {
        const environments = await api.listEnvironments();
        const selectedEnvironment = resolveSelectedEnvironment(
                environments,
                url.searchParams.get("environment"),
            );
        if (!selectedEnvironment) {
            throw new Error("No environments are configured.");
        }
        const usageRange = parseUsageRange(url.searchParams.get("usageRange"));
        const flag = await api.getFlag(selectedEnvironment.id, params.key);
        const context = flag.context_id
            ? await api.getContext(flag.context_id).catch(() => null)
            : null;
        const history = await api
            .getFlagAudit(selectedEnvironment.id, params.key)
            .catch(() => []);
        const usage = await api
            .getFlagUsage(selectedEnvironment.id, params.key, usageRange)
            .catch(() => null);
        return {
            environments,
            selectedEnvironment,
            flag,
            context,
            history,
            usage: normalizeUsage(usage),
            usageRange,
        };
    }, url.pathname);
};
