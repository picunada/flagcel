import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent }) => {
    const { contexts } = await parent();
    if (contexts.length > 0) {
        throw redirect(307, `/contexts/${encodeURIComponent(contexts[0].id)}`);
    }
    return {};
};
