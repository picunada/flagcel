import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) =>
    runLoad(async () => {
        const api = createApi(fetch);
        return {
            environments: await api.listEnvironments(),
        };
    }, url.pathname);
