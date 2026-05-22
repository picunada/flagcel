package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openapiYAML []byte

const swaggerUI = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Flagcel API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    :root {
      --bg: #0f0f0f;
      --surface: rgba(15, 15, 15, 0.9);
      --surface-2: rgba(255, 255, 255, 0.06);
      --border: rgba(255, 255, 255, 0.12);
      --fg: #ffffff;
      --fg-dim: #8e8e8e;
      --accent: #ffffff;
      --radius: 0.25rem;
      --font-mono: ui-monospace, SFMono-Regular, 'JetBrains Mono', Menlo, Monaco, Consolas, monospace;
    }
    .swagger-ui *,
    .swagger-ui *::before,
    .swagger-ui *::after { border-radius: var(--radius) !important; }
    html, body {
      background: var(--bg);
      color: var(--fg);
      font-family: var(--font-mono);
      font-weight: 400;
      letter-spacing: 0.02em;
      font-feature-settings: 'tnum' 1, 'cv11' 1;
    }
    .swagger-ui, .swagger-ui * {
      font-family: var(--font-mono) !important;
      letter-spacing: 0.02em;
      font-feature-settings: 'tnum' 1, 'cv11' 1;
    }
    body, .swagger-ui, .swagger-ui .info .title, .swagger-ui .opblock-tag,
    .swagger-ui .opblock .opblock-summary-operation-id,
    .swagger-ui .opblock .opblock-summary-path,
    .swagger-ui .opblock .opblock-summary-description,
    .swagger-ui .opblock-description-wrapper p,
    .swagger-ui .opblock-external-docs-wrapper p,
    .swagger-ui .opblock-title_normal p,
    .swagger-ui table thead tr td, .swagger-ui table thead tr th,
    .swagger-ui .response-col_status, .swagger-ui .response-col_links,
    .swagger-ui .parameter__name, .swagger-ui .parameter__type,
    .swagger-ui .parameter__in, .swagger-ui label,
    .swagger-ui .tab li, .swagger-ui .scheme-container,
    .swagger-ui section.models h4, .swagger-ui section.models h5,
    .swagger-ui .model, .swagger-ui .model-title,
    .swagger-ui .prop-type, .swagger-ui .prop-format,
    .swagger-ui .info li, .swagger-ui .info p, .swagger-ui .info table,
    .swagger-ui .markdown p, .swagger-ui .renderedMarkdown p,
    .swagger-ui .btn { color: var(--fg) !important; }
    .swagger-ui .info .title small, .swagger-ui .parameter__type,
    .swagger-ui .response-col_description__inner div.markdown,
    .swagger-ui .response-col_description__inner div.renderedMarkdown { color: var(--fg-dim) !important; }
    .swagger-ui .topbar { background: var(--surface); border-bottom: 1px solid var(--border); }
    .swagger-ui .scheme-container { background: var(--surface); box-shadow: none; border: 1px solid var(--border); }
    .swagger-ui .opblock { background: var(--surface); border: 1px solid var(--border); box-shadow: none; }
    .swagger-ui .opblock .opblock-summary { border-bottom: 1px solid var(--border); }
    .swagger-ui .opblock .opblock-section-header { background: var(--surface-2); box-shadow: none; border-bottom: 1px solid var(--border); }
    .swagger-ui .opblock .opblock-section-header h4, .swagger-ui .opblock .opblock-section-header > label { color: var(--fg) !important; }
    .swagger-ui section.models { background: var(--surface); border: 1px solid var(--border); }
    .swagger-ui section.models .model-container { background: var(--surface-2); }
    .swagger-ui .model-box { background: var(--surface-2); }
    .swagger-ui input[type=text], .swagger-ui input[type=password],
    .swagger-ui input[type=email], .swagger-ui input[type=search],
    .swagger-ui textarea, .swagger-ui select {
      background: var(--surface-2) !important; color: var(--fg) !important;
      border: 1px solid var(--border) !important; padding: 6px 12px;
    }
    .swagger-ui .btn { background: var(--surface-2); border: 1px solid var(--border); box-shadow: none; }
    .swagger-ui .btn:hover { background: var(--border); }
    .swagger-ui .btn.authorize { color: var(--fg) !important; border-color: var(--border); }
    .swagger-ui .btn.authorize svg { fill: var(--fg); }
    .swagger-ui .btn.execute { background: var(--accent); color: var(--bg) !important; border-color: var(--accent); }
    .swagger-ui .highlight-code, .swagger-ui .microlight,
    .swagger-ui .opblock-body pre.microlight { background: #000 !important; color: var(--fg) !important; }
    .swagger-ui table tbody tr td { background: transparent; border-color: var(--border); }
    .swagger-ui .opblock-tag { border-bottom: 1px solid var(--border); }
    .swagger-ui .opblock-tag:hover { background: var(--surface); }
    .swagger-ui svg:not(.opblock-summary-method svg) { fill: var(--fg); }
    .swagger-ui .arrow { fill: var(--fg); }
    .swagger-ui .dialog-ux .modal-ux { background: var(--surface); border: 1px solid var(--border); }
    .swagger-ui .dialog-ux .modal-ux-header, .swagger-ui .dialog-ux .modal-ux-content { color: var(--fg); }
    .swagger-ui .info a, .swagger-ui a { color: var(--fg); }
    .swagger-ui .response-control-media-type__title,
    .swagger-ui .response-control-media-type__accept-message,
    .swagger-ui .responses-inner h4, .swagger-ui .responses-inner h5 { color: var(--fg) !important; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/openapi.yaml",
      dom_id: "#swagger-ui",
      deepLinking: true,
    });
  </script>
</body>
</html>`

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", serveSpec)
	mux.HandleFunc("GET /docs", serveUI)
}

func serveSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Write(openapiYAML)
}

func serveUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerUI))
}
