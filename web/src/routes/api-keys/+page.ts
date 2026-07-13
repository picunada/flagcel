import { createApi } from "$lib/api";
import { resolveSelectedEnvironment } from "$lib/environment";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) =>
    runLoad(async () => {
        const api = createApi(fetch);
        const environments = await api.listEnvironments();
        const selectedEnvironment = resolveSelectedEnvironment(
                environments,
                url.searchParams.get("environment"),
            );
        if (!selectedEnvironment) {
            throw new Error("No environments are configured.");
        }
        return {
            keys: await api.listAPIKeys(),
            environments,
            selectedEnvironment,
        };
    }, url.pathname);
