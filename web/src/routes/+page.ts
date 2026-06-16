import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
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
            return {
                environments,
                selectedEnvironment,
                flags: selectedEnvironment ? await api.listFlags(selectedEnvironment.id) : [],
                contexts: await api.listContexts(),
            };
        },
        url.pathname,
    );
};
