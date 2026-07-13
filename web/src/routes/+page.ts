import { createApi } from "$lib/api";
import { resolveSelectedEnvironment } from "$lib/environment";
import { runLoad } from "$lib/load";
import { normalizeUsage, parseUsageRange } from "$lib/usage-analytics";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) => {
    const api = createApi(fetch);
    return runLoad(
        async () => {
            const environments = await api.listEnvironments();
            const selectedEnvironment = resolveSelectedEnvironment(
                environments,
                url.searchParams.get("environment"),
            );
            const usageRange = parseUsageRange(url.searchParams.get("usageRange"));
            return {
                environments,
                selectedEnvironment,
                flags: selectedEnvironment ? await api.listFlags(selectedEnvironment.id) : [],
                environmentUsage: selectedEnvironment
                    ? normalizeUsage(
                          await api
                              .getEnvironmentUsage(selectedEnvironment.id, usageRange)
                              .catch(() => null),
                      )
                    : normalizeUsage(null),
                usageRange,
                contexts: await api.listContexts(),
            };
        },
        url.pathname,
    );
};
