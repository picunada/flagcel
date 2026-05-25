import { api } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params, url }) =>
    runLoad(async () => {
        const flag = await api.getFlag(params.key);
        const context = flag.context_id
            ? await api.getContext(flag.context_id).catch(() => null)
            : null;
        return { flag, context };
    }, url.pathname);
