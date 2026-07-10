import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) =>
    runLoad(async () => {
        const api = createApi(fetch);
        const environments = await api.listEnvironments();
        const flagsByEnvironment = new Map(
            await Promise.all(
                environments.map(async (environment) => [
                    environment.id,
                    await api.listFlags(environment.id),
                ] as const),
            ),
        );
        const defaultEnvironment = environments.find((environment) => environment.key === "production");
        const defaultFlags = defaultEnvironment
            ? flagsByEnvironment.get(defaultEnvironment.id) ?? []
            : [];
        const defaultByKey = new Map(defaultFlags.map((flag) => [flag.key, flag]));

        return {
            environments,
            environmentMetrics: environments.map((environment) => {
                const flags = flagsByEnvironment.get(environment.id) ?? [];
                return {
                    environment_id: environment.id,
                    flag_count: flags.length,
                    override_count:
                        environment.id === defaultEnvironment?.id
                            ? 0
                            : flags.filter((flag) => {
                                  const baseline = defaultByKey.get(flag.key);
                                  return !baseline || targetingSignature(flag) !== targetingSignature(baseline);
                              }).length,
                };
            }),
        };
    }, url.pathname);

function targetingSignature(flag: {
    enabled: boolean;
    context_id?: string | null;
    rules: unknown[];
}) {
    return JSON.stringify({
        enabled: flag.enabled,
        context_id: flag.context_id ?? null,
        rules: flag.rules,
    });
}
