import { redirect } from "@sveltejs/kit";
import { APIError } from "$lib/api";

export async function runLoad<T>(fn: () => Promise<T>, currentPath: string): Promise<T> {
    try {
        return await fn();
    } catch (e) {
        if (e instanceof APIError && e.status === 401 && currentPath !== "/login") {
            throw redirect(307, "/login");
        }
        throw e;
    }
}
