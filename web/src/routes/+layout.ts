import { redirect } from "@sveltejs/kit";
import { createApi, APIError, type AuthMe } from "$lib/api";
import type { LayoutLoad } from "./$types";

export const ssr = false;
export const prerender = false;
export const trailingSlash = "never";

export const load: LayoutLoad = async ({ url, fetch }) => {
    let auth: AuthMe;
    try {
        auth = await createApi(fetch).me();
    } catch (e) {
        if (e instanceof APIError && e.status === 401) {
            if (url.pathname !== "/login") throw redirect(307, "/login");
            return { auth: { auth_enabled: true, authenticated: false } satisfies AuthMe };
        }
        throw e;
    }

    if (auth.auth_enabled && !auth.authenticated && url.pathname !== "/login") {
        throw redirect(307, "/login");
    }
    if (auth.authenticated && url.pathname === "/login") {
        throw redirect(307, "/");
    }

    return { auth };
};
