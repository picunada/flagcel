import { api } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params, url }) =>
    runLoad(async () => ({ schema: await api.getContext(params.id) }), url.pathname);
