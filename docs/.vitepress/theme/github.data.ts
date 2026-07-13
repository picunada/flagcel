import { defineLoader } from "vitepress";

const REPO = "picunada/flagcel";

export interface GitHubData {
    // null when the value could not be fetched (e.g. offline build)
    stars: number | null;
    // newest app release tag, e.g. "v0.1.0-rc.1"; null if none found
    version: string | null;
}

declare const data: GitHubData;
export { data };

interface SemVer {
    parts: [number, number, number];
    pre: string;
}

function parseTag(name: string): SemVer | null {
    // only the app's own version tags - ignore sdk-scoped tags like
    // "sdks/js/v0.1.1" (they contain a slash)
    const m = /^v(\d+)\.(\d+)\.(\d+)(?:-(.+))?$/.exec(name);
    if (!m) return null;
    return { parts: [+m[1], +m[2], +m[3]], pre: m[4] ?? "" };
}

// descending: newest first. A release outranks a pre-release of the same version.
function compare(a: SemVer, b: SemVer): number {
    for (let i = 0; i < 3; i++) {
        if (a.parts[i] !== b.parts[i]) return b.parts[i] - a.parts[i];
    }
    if (a.pre === b.pre) return 0;
    if (!a.pre) return -1;
    if (!b.pre) return 1;
    return a.pre < b.pre ? 1 : -1;
}

export default defineLoader({
    async load(): Promise<GitHubData> {
        const headers: Record<string, string> = {
            Accept: "application/vnd.github+json",
            "User-Agent": "flagcel-docs",
        };
        const token = process.env.GITHUB_TOKEN;
        if (token) headers.Authorization = `Bearer ${token}`;

        try {
            const [repoRes, tagsRes] = await Promise.all([
                fetch(`https://api.github.com/repos/${REPO}`, { headers }),
                fetch(`https://api.github.com/repos/${REPO}/tags?per_page=100`, {
                    headers,
                }),
            ]);

            const stars = repoRes.ok
                ? ((await repoRes.json()).stargazers_count ?? null)
                : null;

            let version: string | null = null;
            if (tagsRes.ok) {
                const tags: { name: string }[] = await tagsRes.json();
                const newest = tags
                    .map((t) => ({ name: t.name, ver: parseTag(t.name) }))
                    .filter((t): t is { name: string; ver: SemVer } => t.ver !== null)
                    .sort((a, b) => compare(a.ver, b.ver))[0];
                version = newest?.name ?? null;
            }

            return { stars, version };
        } catch {
            // offline / API failure - render without the pills rather than fail the build
            return { stars: null, version: null };
        }
    },
});
