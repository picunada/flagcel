import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url, fetch }) =>
    runLoad(async () => ({ contexts: await createApi(fetch).listContexts() }), url.pathname);
