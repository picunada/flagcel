<script setup lang="ts">
import { ref, computed } from "vue";
import { withBase } from "vitepress";
import { data as gh } from "../github.data";

const repo = "https://github.com/picunada/flagcel";

const tab = ref(0);
const tabs = ["Run", "Evaluate"];

const starsLabel = computed(() => {
    const n = gh.stars;
    if (n == null) return null;
    const count =
        n >= 1000
            ? (n / 1000).toFixed(n >= 10000 ? 0 : 1).replace(/\.0$/, "") + "k"
            : String(n);
    return `${count} ${n === 1 ? "star" : "stars"}`;
});

const cards = [
    {
        icon: "M5 19c-1 1-2 4-2 4s3-1 4-2m4-3a16 16 0 0 0 7-12c0-1 0-2-.2-2.8C17 3 16 3 15 3A16 16 0 0 0 3 10m6 5-3-3m3 3 2 4 4-6m-9-1L3 9l4 2",
        h: "Quickstart",
        d: "Install & ship a flag",
        n: "01",
        link: "/quickstart",
    },
    {
        icon: "M4 5h16l-6 7v6l-4 2v-8L4 5z",
        h: "CEL targeting",
        d: "Rule syntax & macros",
        n: "02",
        link: "/concepts",
    },
    {
        icon: "m12 2 9 5v10l-9 5-9-5V7l9-5zM3 7l9 5 9-5M12 12v10",
        h: "SDKs",
        d: "Go · TypeScript · Python",
        n: "03",
        link: "/sdks",
    },
    {
        icon: "m8 4-4 8 4 8M16 4l4 8-4 8",
        h: "API reference",
        d: "Versioned REST v1",
        n: "04",
        link: "/api",
    },
];
</script>

<template>
    <div class="fc-home">
        <div class="fc-home__inner">
            <div class="fc-home__hero">
                <!-- left column -->
                <div class="fc-home__lead">
                    <div class="fc-home__pills">
                        <span class="hh-pill"
                            ><span class="hh-dot hh-dot--green"></span>
                            Apache-2.0</span
                        >
                        <a
                            v-if="starsLabel"
                            class="hh-pill hh-pill--link"
                            :href="repo"
                            target="_blank"
                            rel="noreferrer"
                        >
                            <svg
                                width="13"
                                height="13"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                                style="opacity: 0.8"
                            >
                                <path
                                    d="M12 2l2.9 6.3 6.9.6-5.2 4.5 1.6 6.7L12 16.9 5.8 20.6l1.6-6.7L2.2 8.9l6.9-.6z"
                                />
                            </svg>
                            {{ starsLabel }}
                        </a>
                        <span v-if="gh.version" class="hh-pill">{{
                            gh.version
                        }}</span>
                    </div>

                    <h1 class="fc-home__title">Flagcel</h1>

                    <p class="fc-home__sub">
                        Self-hosted, CEL-targeted, and OpenFeature-native. These
                        docs take you from
                        <code class="hh-ic">docker run</code> to a production
                        rollout.
                    </p>

                    <div class="fc-home__cta">
                        <a
                            class="hh-cta hh-cta--primary"
                            :href="withBase('/quickstart')"
                        >
                            Start the Quickstart
                            <svg
                                width="15"
                                height="15"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2.2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path d="M5 12h13M13 6l6 6-6 6" />
                            </svg>
                        </a>
                        <a
                            class="hh-cta"
                            :href="repo"
                            target="_blank"
                            rel="noreferrer"
                        >
                            <svg
                                width="15"
                                height="15"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                            >
                                <path
                                    d="M12 2A10 10 0 0 0 8.8 21.5c.5.1.7-.2.7-.5v-1.7c-2.8.6-3.4-1.3-3.4-1.3-.5-1.2-1.1-1.5-1.1-1.5-.9-.6.1-.6.1-.6 1 .1 1.5 1 1.5 1 .9 1.5 2.3 1.1 2.9.8.1-.6.3-1.1.6-1.3-2.2-.3-4.6-1.1-4.6-5 0-1.1.4-2 1-2.7-.1-.3-.4-1.3.1-2.7 0 0 .8-.3 2.7 1a9.4 9.4 0 0 1 5 0c1.9-1.3 2.7-1 2.7-1 .5 1.4.2 2.4.1 2.7.6.7 1 1.6 1 2.7 0 3.9-2.4 4.7-4.6 5 .3.3.7 1 .7 2v3c0 .3.2.6.7.5A10 10 0 0 0 12 2z"
                                />
                            </svg>
                            GitHub
                        </a>
                    </div>
                </div>

                <!-- right column: code proof -->
                <div class="fc-home__proof">
                    <div class="hh-eyebrow">Two commands to production</div>
                    <div class="hh-proof">
                        <div class="hh-glow" aria-hidden="true"></div>
                        <div class="hh-code">
                            <div class="hh-code__bar">
                                <span class="hh-dots">
                                    <i style="background: #ff5f57"></i>
                                    <i style="background: #febc2e"></i>
                                    <i style="background: #28c840"></i>
                                </span>
                                <div class="hh-tabs" role="tablist">
                                    <button
                                        v-for="(t, i) in tabs"
                                        :key="t"
                                        role="tab"
                                        :aria-selected="i === tab"
                                        :class="{ on: i === tab }"
                                        @click="tab = i"
                                    >
                                        {{ t }}
                                    </button>
                                </div>
                                <span class="hh-lang">{{
                                    tab === 0 ? "bash" : "typescript"
                                }}</span>
                                <span class="hh-cp">
                                    <svg
                                        width="15"
                                        height="15"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="1.7"
                                    >
                                        <rect
                                            x="9"
                                            y="9"
                                            width="11"
                                            height="11"
                                            rx="2"
                                        />
                                        <path d="M5 15V5a2 2 0 0 1 2-2h10" />
                                    </svg>
                                </span>
                            </div>

                            <pre
                                v-if="tab === 0"
                                class="hh-pre"
                            ><code><span class="cm"># run the control plane — migrations included</span>
<span class="pr">$</span> docker run <span class="nm">-p</span> 8080:8080 \
    <span class="nm">-e</span> FLAGCEL_DB_URL=<span class="st">$DATABASE_URL</span> \
    ghcr.io/flagcel/flagcel:<span class="st">latest</span>

<span class="cm"># dashboard + versioned API, one binary</span>
<span class="pr">→</span> http://localhost:8080</code></pre>

                            <pre
                                v-else
                                class="hh-pre"
                            ><code><span class="kw">const</span> client = OpenFeature.<span class="fn2">getClient</span>(<span class="st">"checkout"</span>)

<span class="kw">const</span> on = <span class="kw">await</span> client.<span class="fn2">getBooleanValue</span>(
  <span class="st">"checkout-v2"</span>, <span class="kw">false</span>, {
    targetingKey: user.id,
    plan: user.plan,
    region: geo.region,
  })</code></pre>

                            <div class="hh-status">
                                <template v-if="tab === 0">
                                    <span
                                        class="hh-status__dot"
                                        style="
                                            background: var(--green);
                                            box-shadow: 0 0 8px var(--green);
                                        "
                                    ></span>
                                    <span
                                        ><b>ready</b> · dashboard live on :8080
                                        in 1.2s</span
                                    >
                                </template>
                                <template v-else>
                                    <span
                                        class="hh-status__dot"
                                        style="
                                            background: var(--acc);
                                            box-shadow: 0 0 8px var(--acc);
                                        "
                                    ></span>
                                    <span
                                        ><b>returns true</b> · matched
                                        <code
                                            class="hh-ic"
                                            style="font-size: 11px"
                                            >plan == "pro" &amp;&amp; region ==
                                            "EU"</code
                                        ></span
                                    >
                                </template>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- browse the docs -->
            <div class="fc-home__browse">
                <div class="fc-home__browse-head">
                    <div class="hh-eyebrow">Browse the docs</div>
                    <a class="hh-link" :href="withBase('/quickstart')"
                        >All sections →</a
                    >
                </div>
                <div class="fc-home__grid">
                    <a
                        v-for="c in cards"
                        :key="c.h"
                        class="hh-card"
                        :href="withBase(c.link)"
                    >
                        <div class="hh-card__top">
                            <span class="hh-card__icon">
                                <svg
                                    width="20"
                                    height="20"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="1.7"
                                    stroke-linejoin="round"
                                    stroke-linecap="round"
                                >
                                    <path :d="c.icon" />
                                </svg>
                            </span>
                            <span class="hh-card__n">{{ c.n }}</span>
                        </div>
                        <div class="hh-card__h">
                            {{ c.h }}<span class="hh-arr">→</span>
                        </div>
                        <div class="hh-card__d">{{ c.d }}</div>
                    </a>
                </div>
            </div>
        </div>
    </div>
</template>

<style>
.fc-home {
    --acc: #f97316;
    --acc-2: #fdba74;
    --acc-dim: rgba(249, 115, 22, 0.14);
    --green: #3ecf8e;
    --bg: #0b0b0d;
    --bg-2: #101013;
    --surface: #16161a;
    --surface-2: #1b1b20;
    --line: rgba(255, 255, 255, 0.07);
    --line-2: rgba(255, 255, 255, 0.11);
    --tx: #f4f3f1;
    --tx-2: #b6b5bd;
    --tx-3: #76767f;
    --mono: var(--vp-font-family-mono);
    --sans: var(--vp-font-family-base);
}

.fc-home__inner {
    max-width: 1180px;
    margin: 0 auto;
    padding: 64px 48px 64px;
}

.fc-home__hero {
    display: grid;
    grid-template-columns: 1.02fr 1fr;
    gap: 64px;
    align-items: center;
}

/* ---- pills ---- */
.fc-home__pills {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
}

.hh-pill {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    font-family: var(--mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--tx-2);
    border: 1px solid var(--line);
    background: var(--bg-2);
    border-radius: 20px;
    padding: 5px 12px;
}

.hh-pill--link {
    text-decoration: none;
    color: var(--tx-2);
    transition:
        border-color 0.18s ease-out,
        color 0.18s ease-out;
}

.hh-pill--link:hover {
    border-color: var(--line-2);
    color: var(--tx);
}

.hh-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
}

.hh-dot--green {
    background: var(--green);
}

/* ---- headline ---- */
.fc-home__title {
    margin: 22px 0 0;
    font-family: var(--mono);
    font-weight: 700;
    font-size: 50px;
    line-height: 1.07;
    letter-spacing: -0.022em;
    color: var(--tx);
}

.fc-home__sub {
    margin: 20px 0 0;
    max-width: 440px;
    font-size: 18px;
    line-height: 1.6;
    color: var(--tx-2);
}

.hh-ic {
    font-family: var(--mono);
    font-size: 0.86em;
    background: var(--surface-2);
    border: 1px solid var(--line);
    padding: 1px 6px;
    border-radius: 6px;
    color: var(--acc-2);
}

/* ---- CTAs ---- */
.fc-home__cta {
    display: flex;
    gap: 12px;
    margin-top: 28px;
    flex-wrap: wrap;
}

.hh-cta {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    height: 46px;
    padding: 0 22px;
    border-radius: 11px;
    font-family: var(--mono);
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.01em;
    text-decoration: none;
    cursor: pointer;
    background: var(--surface);
    color: var(--tx);
    border: 1px solid var(--line-2);
    transition:
        border-color 0.18s ease-out,
        background-color 0.18s ease-out,
        color 0.18s ease-out,
        transform 0.12s ease-out;
}

.hh-cta:hover {
    border-color: var(--line-2);
    background: var(--surface-2);
}

.hh-cta:active {
    transform: translateY(0) scale(0.98);
}

.hh-cta--primary {
    background: var(--acc);
    color: #1a0f06;
    border-color: var(--acc);
}

.hh-cta--primary:hover {
    background: var(--acc-2);
    border-color: var(--acc-2);
    color: #1a0f06;
}

/* ---- eyebrow ---- */
.hh-eyebrow {
    font-family: var(--mono);
    font-size: 12px;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: var(--acc);
    font-weight: 500;
}

/* ---- code proof ---- */
.fc-home__proof .hh-eyebrow {
    margin-bottom: 14px;
}

.hh-proof {
    position: relative;
}

.hh-glow {
    position: absolute;
    inset: -14% -10%;
    background: radial-gradient(
        58% 56% at 72% 26%,
        rgba(249, 115, 22, 0.18),
        transparent 70%
    );
    filter: blur(22px);
    pointer-events: none;
}

.hh-code {
    position: relative;
    border: 1px solid var(--line);
    border-radius: 12px;
    overflow: hidden;
    background: #0e0e11;
    box-shadow:
        0 28px 64px -22px rgba(0, 0, 0, 0.75),
        0 0 0 1px var(--line);
}

.hh-code__bar {
    display: flex;
    align-items: center;
    gap: 9px;
    height: 40px;
    padding: 0 14px;
    border-bottom: 1px solid var(--line);
    background: var(--bg-2);
}

.hh-dots {
    display: flex;
    gap: 6px;
}

.hh-dots i {
    width: 11px;
    height: 11px;
    border-radius: 50%;
    opacity: 0.85;
}

.hh-tabs {
    display: flex;
    gap: 2px;
    margin-left: 6px;
}

.hh-tabs button {
    font-family: var(--mono);
    font-size: 12.5px;
    color: var(--tx-3);
    background: none;
    border: none;
    padding: 5px 11px;
    border-radius: 7px;
    cursor: pointer;
    transition:
        color 0.15s,
        background 0.15s;
}

.hh-tabs button.on {
    color: var(--tx);
    background: var(--surface-2);
}

.hh-lang {
    margin-left: auto;
    font-family: var(--mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--tx-3);
}

.hh-cp {
    color: var(--tx-3);
    display: grid;
    place-items: center;
    margin-left: 12px;
    cursor: pointer;
}

.hh-pre {
    margin: 0;
    padding: 18px;
    min-height: 196px;
    overflow: auto;
    font-family: var(--mono);
    font-size: 13.5px;
    line-height: 1.75;
    color: var(--tx);
}

.hh-pre code {
    font-family: var(--mono);
    background: none;
    border: none;
    padding: 0;
    color: inherit;
}

.hh-pre .cm {
    color: var(--tx-3);
}
.hh-pre .kw {
    color: #c98aff;
}
.hh-pre .st {
    color: var(--green);
}
.hh-pre .fn2 {
    color: #7cc7ff;
}
.hh-pre .nm {
    color: #f0a868;
}
.hh-pre .pr {
    color: var(--acc);
}

.hh-status {
    border-top: 1px solid var(--line);
    padding: 11px 16px;
    display: flex;
    align-items: center;
    gap: 9px;
    font-family: var(--mono);
    font-size: 12px;
    color: var(--tx-2);
    background: var(--bg-2);
}

.hh-status b {
    color: var(--tx);
    font-weight: 600;
}

.hh-status__dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    flex: none;
}

/* ---- browse the docs ---- */
.fc-home__browse {
    margin-top: 64px;
}

.fc-home__browse-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 18px;
}

.hh-link {
    font-family: var(--mono);
    font-size: 13px;
    color: var(--acc-2);
    text-decoration: none;
    border-bottom: 1px solid rgba(253, 186, 116, 0.3);
}

.fc-home__grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
}

.hh-card {
    display: block;
    border: 1px solid var(--line);
    border-radius: 12px;
    padding: 18px;
    background: var(--bg-2);
    text-decoration: none;
    transition:
        border-color 0.15s,
        transform 0.15s,
        background 0.15s,
        box-shadow 0.15s;
}

.hh-card:hover {
    border-color: var(--line-2);
    transform: translateY(-2px);
    background: var(--surface);
    box-shadow: 0 12px 28px -16px rgba(0, 0, 0, 0.7);
}

.hh-card__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.hh-card__icon {
    width: 38px;
    height: 38px;
    border-radius: 10px;
    display: grid;
    place-items: center;
    background: var(--acc-dim);
    border: 1px solid rgba(249, 115, 22, 0.22);
    color: var(--acc);
}

.hh-card__n {
    font-family: var(--mono);
    font-size: 12px;
    color: var(--tx-3);
}

.hh-card__h {
    margin-top: 15px;
    font-family: var(--mono);
    font-weight: 600;
    font-size: 15px;
    color: var(--tx);
    display: flex;
    align-items: center;
    gap: 8px;
}

.hh-arr {
    color: var(--acc);
}

.hh-card__d {
    margin-top: 5px;
    font-size: 13.5px;
    color: var(--tx-2);
    line-height: 1.5;
}

.fc-home a:focus-visible,
.fc-home button:focus-visible {
    outline: 2px solid var(--acc);
    outline-offset: 3px;
    border-radius: 9px;
}

/* ---- responsive ---- */
@media (max-width: 1100px) {
    .fc-home__hero {
        gap: 44px;
    }
}

@media (max-width: 959px) {
    .fc-home__inner {
        padding: 48px 32px 56px;
    }

    .fc-home__hero {
        grid-template-columns: 1fr;
        gap: 48px;
    }

    .fc-home__title {
        font-size: 42px;
    }

    .fc-home__grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 640px) {
    .fc-home__inner {
        padding: 36px 20px 48px;
    }

    .fc-home__title {
        font-size: 34px;
    }

    .fc-home__title br {
        display: none;
    }

    .fc-home__grid {
        grid-template-columns: 1fr;
    }
}
</style>
