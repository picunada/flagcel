import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import { normalizeUsage, parseUsageRange } from "$lib/usage-analytics";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) => {
    const api = createApi(fetch);
    return runLoad(
        async () => {
            const environments = await api.listEnvironments();
            const selectedEnvironment =
                environments.find((env) => env.id === url.searchParams.get("environment")) ??
                environments.find((env) => env.key === "production") ??
                environments[0];
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
