import { api } from "$lib/api";
import { runLoad } from "$lib/load";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ url }) =>
    runLoad(async () => ({ flags: await api.listFlags() }), url.pathname);
