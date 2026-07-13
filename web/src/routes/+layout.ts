import { redirect } from "@sveltejs/kit";
import { createApi, APIError, type AuthMe } from "$lib/api";
import { resolveSelectedEnvironment } from "$lib/environment";
import type { LayoutLoad } from "./$types";

export const ssr = false;
export const prerender = false;
export const trailingSlash = "never";

const unauthenticatedAuth: AuthMe = {
    auth_enabled: false,
    authenticated: false,
};

export const load: LayoutLoad = async ({ url, fetch }) => {
    const route = { currentPathname: url.pathname, currentSearch: url.search };
	let auth: AuthMe;
	const api = createApi(fetch);
	try {
		auth = await api.me();
    } catch (e) {
        if (e instanceof APIError && e.status === 401) {
            if (url.pathname !== "/login") throw redirect(307, "/login");
            return {
                ...route,
                auth: { auth_enabled: true, authenticated: false } satisfies AuthMe,
            };
        }
        if (e instanceof APIError && (e.status === 0 || e.status >= 500)) {
            return {
                ...route,
                auth: unauthenticatedAuth,
                backendUnavailable: true,
                backendMessage: e.message,
            };
        }
        throw e;
    }

    if (auth.auth_enabled && !auth.authenticated && url.pathname !== "/login") {
        throw redirect(307, "/login");
    }
    if (auth.authenticated && url.pathname === "/login") {
        throw redirect(307, "/");
    }

	if (!auth.authenticated || url.pathname === "/login") {
		return { ...route, auth };
	}

	const environments = await api.listEnvironments();
	const selectedEnvironment = resolveSelectedEnvironment(
		environments,
		url.searchParams.get("environment"),
	);

	return { ...route, auth, environments, selectedEnvironment };
};
