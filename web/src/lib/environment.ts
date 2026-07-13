import { browser } from "$app/environment";
import type { Environment } from "$lib/api";
import { pickSelectedEnvironment } from "$lib/environment-selection";

const STORAGE_KEY = "flagcel.selectedEnvironmentId";

export function rememberEnvironmentId(id: string | null | undefined) {
    if (!browser || !id) return;
    try {
        localStorage.setItem(STORAGE_KEY, id);
    } catch {
        // Ignore quota / private-mode failures.
    }
}

export function readRememberedEnvironmentId(): string | null {
    if (!browser) return null;
    try {
        return localStorage.getItem(STORAGE_KEY);
    } catch {
        return null;
    }
}

export function resolveSelectedEnvironment(
    environments: Environment[],
    requestedId: string | null,
): Environment | undefined {
    const selected = pickSelectedEnvironment(
        environments,
        requestedId,
        readRememberedEnvironmentId(),
    );
    if (selected) rememberEnvironmentId(selected.id);
    return selected;
}

export { pickSelectedEnvironment };
