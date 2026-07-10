import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { LayoutLoad } from "./$types";

export const load: LayoutLoad = ({ url, fetch }) =>
    runLoad(async () => {
        const api = createApi(fetch);
        const [contexts, contextReferences] = await Promise.all([
            api.listContexts(),
            api.listContextReferences(),
        ]);
        return { contexts, contextReferences };
    }, url.pathname);
