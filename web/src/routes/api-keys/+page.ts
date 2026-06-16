import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) =>
    runLoad(async () => {
        const api = createApi(fetch);
        const environments = await api.listEnvironments();
        const selectedEnvironment =
            environments.find((env) => env.id === url.searchParams.get("environment")) ??
            environments.find((env) => env.key === "production") ??
            environments[0];
        if (!selectedEnvironment) {
            throw new Error("No environments are configured.");
        }
        return {
            keys: await api.listAPIKeys(),
            environments,
            selectedEnvironment,
        };
    }, url.pathname);
