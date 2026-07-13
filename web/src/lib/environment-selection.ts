import type { Environment } from "$lib/api";

export function pickSelectedEnvironment(
    environments: Environment[],
    requestedId: string | null,
    rememberedId: string | null = null,
): Environment | undefined {
    return (
        environments.find((env) => env.id === requestedId) ??
        environments.find((env) => env.id === rememberedId) ??
        environments.find((env) => env.key === "production") ??
        environments[0]
    );
}
