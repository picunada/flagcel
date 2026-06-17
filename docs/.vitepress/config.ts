import { defineConfig } from "vitepress";

const repo = "https://github.com/picunada/flagcel";
const flagcelDarkCodeTheme = {
    name: "flagcel-dark",
    type: "dark" as const,
    colors: {
        "editor.background": "#1b1b1f",
        "editor.foreground": "#a6accd",
    },
    settings: [
        {
            settings: {
                background: "#1b1b1f",
                foreground: "#a6accd",
            },
        },
        {
            scope: ["comment", "punctuation.definition.comment"],
            settings: {
                foreground: "#767c9d",
                fontStyle: "italic",
            },
        },
        {
            scope: ["string", "constant.other.symbol", "constant.other.key"],
            settings: {
                foreground: "#ffb86b",
            },
        },
        {
            scope: ["constant.numeric", "constant.language", "constant.character"],
            settings: {
                foreground: "#ff9f43",
            },
        },
        {
            scope: ["keyword", "storage", "storage.type"],
            settings: {
                foreground: "#a6accd",
            },
        },
        {
            scope: ["entity.name.function", "support.function", "meta.function-call"],
            settings: {
                foreground: "#add7ff",
            },
        },
        {
            scope: ["entity.name.type", "support.type", "support.class"],
            settings: {
                foreground: "#91b4d5",
            },
        },
        {
            scope: ["variable", "variable.other", "meta.definition.variable"],
            settings: {
                foreground: "#e4f0fb",
            },
        },
        {
            scope: [
                "variable.parameter",
                "meta.object-literal.key",
                "source.json meta.structure.dictionary.json support.type.property-name.json",
            ],
            settings: {
                foreground: "#e4f0fb",
            },
        },
        {
            scope: [
                "entity.name.tag",
                "keyword.control.module",
                "keyword.control.import",
                "keyword.control.export",
                "keyword.control.default",
                "keyword.operator.new",
                "keyword.control.new",
            ],
            settings: {
                foreground: "#ff9f43",
            },
        },
        {
            scope: ["entity.other.attribute-name", "support.variable.property"],
            settings: {
                foreground: "#91b4d5",
            },
        },
        {
            scope: ["punctuation", "meta.brace", "meta.delimiter"],
            settings: {
                foreground: "#a6accd",
            },
        },
        {
            scope: [
                "invalid",
                "invalid.illegal",
                "support.class.error",
                "keyword.control.trycatch",
                "constant.language.null",
                "constant.language.boolean.false",
                "constant.language.undefined",
            ],
            settings: {
                foreground: "#d0679d",
            },
        },
        {
            scope: ["markup.heading", "token.warn-token"],
            settings: {
                foreground: "#fffac2",
            },
        },
    ],
};
const flagcelLightCodeTheme = {
    name: "flagcel-light",
    type: "light" as const,
    colors: {
        "editor.background": "#f6f6f7",
        "editor.foreground": "#171717",
    },
    settings: [
        {
            settings: {
                background: "#f6f6f7",
                foreground: "#171717",
            },
        },
        {
            scope: ["comment", "punctuation.definition.comment"],
            settings: {
                foreground: "#6f748f",
                fontStyle: "italic",
            },
        },
        {
            scope: ["string", "constant.other.symbol", "constant.other.key"],
            settings: {
                foreground: "#c2410c",
            },
        },
        {
            scope: ["constant.numeric", "constant.language", "constant.character"],
            settings: {
                foreground: "#f97316",
            },
        },
        {
            scope: ["keyword", "storage", "storage.type"],
            settings: {
                foreground: "#5f6680",
            },
        },
        {
            scope: ["entity.name.function", "support.function", "meta.function-call"],
            settings: {
                foreground: "#2563eb",
            },
        },
        {
            scope: ["entity.name.type", "support.type", "support.class"],
            settings: {
                foreground: "#3b668a",
            },
        },
        {
            scope: ["variable", "variable.other", "meta.definition.variable"],
            settings: {
                foreground: "#171717",
            },
        },
        {
            scope: [
                "variable.parameter",
                "meta.object-literal.key",
                "source.json meta.structure.dictionary.json support.type.property-name.json",
            ],
            settings: {
                foreground: "#111827",
            },
        },
        {
            scope: [
                "entity.name.tag",
                "keyword.control.module",
                "keyword.control.import",
                "keyword.control.export",
                "keyword.control.default",
                "keyword.operator.new",
                "keyword.control.new",
            ],
            settings: {
                foreground: "#f97316",
            },
        },
        {
            scope: ["entity.other.attribute-name", "support.variable.property"],
            settings: {
                foreground: "#3b668a",
            },
        },
        {
            scope: ["punctuation", "meta.brace", "meta.delimiter"],
            settings: {
                foreground: "#6f748f",
            },
        },
        {
            scope: [
                "invalid",
                "invalid.illegal",
                "support.class.error",
                "keyword.control.trycatch",
                "constant.language.null",
                "constant.language.boolean.false",
                "constant.language.undefined",
            ],
            settings: {
                foreground: "#be185d",
            },
        },
        {
            scope: ["markup.heading", "token.warn-token"],
            settings: {
                foreground: "#a16207",
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
    appearance: "dark",
    lastUpdated: true,
    markdown: {
        lineNumbers: false,
        theme: {
            light: flagcelLightCodeTheme,
            dark: flagcelDarkCodeTheme,
        },
    },
    head: [
        ["meta", { name: "theme-color", content: "#0f0f0f" }],
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
        siteTitle: "flagcel",
        nav: [
            { text: "Guide", link: "/quickstart" },
            { text: "Concepts", link: "/concepts" },
            { text: "API", link: "/api" },
            { text: "SDKs", link: "/sdks" },
        ],
        sidebar: [
            {
                text: "Get Started",
                items: [
                    { text: "Overview", link: "/" },
                    { text: "Quickstart", link: "/quickstart" },
                    { text: "Concepts", link: "/concepts" },
                ],
            },
            {
                text: "Operate",
                items: [
                    { text: "Configuration", link: "/configuration" },
                    { text: "Authentication", link: "/auth" },
                    { text: "Migrations", link: "/migrations" },
                ],
            },
            {
                text: "Reference",
                items: [
                    { text: "API", link: "/api" },
                    { text: "SDKs", link: "/sdks" },
                    { text: "Web UI", link: "/web-ui" },
                    { text: "Development", link: "/development" },
                ],
            },
        ],
        search: {
            provider: "local",
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
