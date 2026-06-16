import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) => {
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
        return { environments, selectedEnvironment };
    }, url.pathname);
};
