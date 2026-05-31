import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) => {
    const api = createApi(fetch);
    return runLoad(
        async () => ({
            flags: await api.listFlags(),
            contexts: await api.listContexts(),
        }),
        url.pathname,
    );
};
