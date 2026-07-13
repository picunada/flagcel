import { defineConfig } from "vitepress";

const repo = "https://github.com/picunada/flagcel";
const flagcelDarkCodeTheme = {
    name: "flagcel-dark",
    type: "dark" as const,
    colors: {
        "editor.background": "#0a0a0a",
        "editor.foreground": "#ffffff",
    },
    settings: [
        {
            settings: {
                background: "#0a0a0a",
                foreground: "#ffffff",
            },
        },
        {
            scope: ["comment", "punctuation.definition.comment"],
            settings: {
                foreground: "#8e8e8e",
                fontStyle: "",
            },
        },
        {
            scope: [
                "string",
                "string.quoted",
                "string.template",
                "constant.other.symbol",
                "constant.other.key",
                "constant.other.url",
                "constant.regexp",
            ],
            settings: {
                foreground: "#8fbf7f",
            },
        },
        {
            scope: ["constant.numeric", "constant.character"],
            settings: {
                foreground: "#d0a56f",
            },
        },
        {
            scope: [
                "constant.language",
                "constant.language.boolean",
                "constant.language.null",
                "keyword",
                "storage",
                "storage.type",
                "storage.modifier",
            ],
            settings: {
                foreground: "#c889fd",
            },
        },
        {
            scope: [
                "entity.name.function",
                "entity.name.function.member",
                "support.function",
                "support.function.builtin",
                "meta.function-call",
                "variable.function",
            ],
            settings: {
                foreground: "#7cc6fe",
            },
        },
        {
            scope: ["entity.name.type", "support.type", "support.class"],
            settings: {
                foreground: "#ffffff",
            },
        },
        {
            scope: ["variable", "variable.other", "meta.definition.variable"],
            settings: {
                foreground: "#ffffff",
            },
        },
        {
            scope: [
                "variable.other.constant",
                "variable.other.enummember",
                "support.constant",
            ],
            settings: {
                foreground: "#8fbf7f",
            },
        },
        {
            scope: [
                "variable.parameter",
                "meta.object-literal.key",
                "source.json meta.structure.dictionary.json support.type.property-name.json",
            ],
            settings: {
                foreground: "#ffffff",
            },
        },
        {
            scope: ["entity.name.tag", "entity.other.attribute-name"],
            settings: {
                foreground: "#d0a56f",
            },
        },
        {
            scope: ["support.variable.property", "support.type.property-name"],
            settings: {
                foreground: "#8fbf7f",
            },
        },
        {
            scope: ["punctuation", "meta.brace", "meta.delimiter"],
            settings: {
                foreground: "#ffffff",
            },
        },
        {
            scope: [
                "source.shell support.function",
                "source.shell meta.function-call",
                "source.shell string.unquoted",
            ],
            settings: {
                foreground: "#ffffff",
            },
        },
        {
            scope: [
                "source.shell constant.other.option",
                "source.shell punctuation.definition.variable",
                "source.shell keyword.operator",
                "constant.other.option.shell",
                "punctuation.definition.variable.shell",
            ],
            settings: {
                foreground: "#ffffff",
            },
        },
        {
            scope: [
                "entity.name.command.shell",
                "entity.name.function.call.shell",
                "support.function.builtin.shell",
                "constant.numeric.shell",
            ],
            settings: {
                foreground: "#ffffff",
            },
        },
        {
            scope: ["invalid", "invalid.illegal", "support.class.error"],
            settings: {
                foreground: "#ff6b6b",
            },
        },
        {
            scope: ["markup.heading", "token.warn-token"],
            settings: {
                foreground: "#d0a56f",
            },
        },
    ],
};
export default defineConfig({
    title: "Flagcel",
    description: "Self-hosted feature flags with CEL-based targeting rules.",
    base: "/flagcel/",
    srcDir: "src",
    cleanUrls: true,
    appearance: "force-dark",
    lastUpdated: true,
    markdown: {
        headers: {
            level: [2, 3],
        },
        lineNumbers: false,
        theme: {
            // Dark-only site; VitePress still requires a light key.
            light: flagcelDarkCodeTheme,
            dark: flagcelDarkCodeTheme,
        },
    },
    head: [
        [
            "link",
            {
                rel: "icon",
                type: "image/svg+xml",
                href: "/flagcel/favicon.svg",
            },
        ],
        ["meta", { name: "theme-color", content: "#070707" }],
        [
            "meta",
            {
                name: "og:description",
                content:
                    "Self-hosted feature flags with CEL-based targeting rules.",
            },
        ],
    ],
    themeConfig: {
        logo: "/logo.svg",
        siteTitle: "flagcel",
        nav: [
            { text: "Guide", link: "/quickstart" },
            { text: "Concepts", link: "/concepts" },
            { text: "API", link: "/api" },
            { text: "SDKs", link: "/sdks/" },
        ],
        sidebar: [
            {
                text: "Get Started",
                items: [
                    { text: "Quickstart", link: "/quickstart" },
                    { text: "Concepts", link: "/concepts" },
                ],
            },
            {
                text: "Operate",
                items: [
                    { text: "Dashboard", link: "/dashboard" },
                    { text: "Deployment", link: "/deployment" },
                    { text: "Configuration", link: "/configuration" },
                    { text: "Authentication", link: "/auth" },
                    { text: "Migrations", link: "/migrations" },
                ],
            },
            {
                text: "SDKs",
                items: [
                    { text: "Overview", link: "/sdks/" },
                    { text: "Go", link: "/sdks/go" },
                    { text: "JS/TS", link: "/sdks/js" },
                    { text: "Python", link: "/sdks/python" },
                ],
            },
            {
                text: "Reference",
                items: [
                    { text: "API", link: "/api" },
                    { text: "Web UI", link: "/web-ui" },
                    { text: "Development", link: "/development" },
                ],
            },
        ],
        search: {
            provider: "local",
        },
        outline: {
            level: [2, 3],
            label: "On this page",
        },
        socialLinks: [{ icon: "github", link: repo }],
        editLink: {
            pattern: `${repo}/edit/main/docs/src/:path`,
            text: "Edit this page on GitHub",
        },
        footer: {
            message: "Released under the Apache 2.0 License.",
            copyright: "Copyright (c) Flagcel contributors",
        },
    },
});
