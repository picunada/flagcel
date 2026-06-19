import { defineConfig } from "vitepress";

const repo = "https://github.com/picunada/flagcel";
const flagcelDarkCodeTheme = {
    name: "flagcel-dark",
    type: "dark" as const,
    colors: {
        "editor.background": "#0e0e11",
        "editor.foreground": "#f4f3f1",
    },
    settings: [
        {
            settings: {
                background: "#0e0e11",
                foreground: "#f4f3f1",
            },
        },
        {
            scope: ["comment", "punctuation.definition.comment"],
            settings: {
                foreground: "#76757d",
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
                foreground: "#3dcd8c",
            },
        },
        {
            scope: ["constant.numeric", "constant.character"],
            settings: {
                foreground: "#f0a868",
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
                foreground: "#f4f3f1",
            },
        },
        {
            scope: ["variable", "variable.other", "meta.definition.variable"],
            settings: {
                foreground: "#f4f3f1",
            },
        },
        {
            scope: [
                "variable.other.constant",
                "variable.other.enummember",
                "support.constant",
            ],
            settings: {
                foreground: "#3dcd8c",
            },
        },
        {
            scope: [
                "variable.parameter",
                "meta.object-literal.key",
                "source.json meta.structure.dictionary.json support.type.property-name.json",
            ],
            settings: {
                foreground: "#f4f3f1",
            },
        },
        {
            scope: ["entity.name.tag", "entity.other.attribute-name"],
            settings: {
                foreground: "#f0a868",
            },
        },
        {
            scope: ["support.variable.property", "support.type.property-name"],
            settings: {
                foreground: "#3dcd8c",
            },
        },
        {
            scope: ["punctuation", "meta.brace", "meta.delimiter"],
            settings: {
                foreground: "#f4f3f1",
            },
        },
        {
            scope: [
                "source.shell support.function",
                "source.shell meta.function-call",
                "source.shell string.unquoted",
            ],
            settings: {
                foreground: "#f4f3f1",
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
                foreground: "#f97316",
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
                foreground: "#f97316",
            },
        },
        {
            scope: ["invalid", "invalid.illegal", "support.class.error"],
            settings: {
                foreground: "#f07178",
            },
        },
        {
            scope: ["markup.heading", "token.warn-token"],
            settings: {
                foreground: "#f0a868",
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
            scope: [
                "constant.numeric",
                "constant.language",
                "constant.character",
            ],
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
            scope: [
                "entity.name.function",
                "support.function",
                "meta.function-call",
            ],
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
        headers: {
            level: [2, 3],
        },
        lineNumbers: false,
        theme: {
            light: flagcelLightCodeTheme,
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
        logo: "/logo.svg",
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
                    { text: "Quickstart", link: "/quickstart" },
                    { text: "Concepts", link: "/concepts" },
                ],
            },
            {
                text: "Operate",
                items: [
                    { text: "Deployment", link: "/deployment" },
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
