import { createApi } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params, url, fetch }) =>
    runLoad(async () => ({ schema: await createApi(fetch).getContext(params.id) }), url.pathname);
