import { createApi } from "$lib/api";
import { resolveSelectedEnvironment } from "$lib/environment";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) => {
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
        return { environments, selectedEnvironment };
    }, url.pathname);
};
