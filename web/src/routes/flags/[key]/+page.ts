import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params, url, fetch }) => {
    const api = createApi(fetch);
    return runLoad(async () => {
        const environments = await api.listEnvironments();
        const selectedEnvironment =
            environments.find((env) => env.id === url.searchParams.get("environment")) ??
            environments.find((env) => env.key === "production") ??
            environments[0];
        if (!selectedEnvironment) {
            throw new Error("No environments are configured.");
        }
        const flag = await api.getFlag(selectedEnvironment.id, params.key);
        const context = flag.context_id
            ? await api.getContext(flag.context_id).catch(() => null)
            : null;
        const history = await api
            .getFlagAudit(selectedEnvironment.id, params.key)
            .catch(() => []);
        return { environments, selectedEnvironment, flag, context, history };
    }, url.pathname);
};
