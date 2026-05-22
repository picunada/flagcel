// Package web embeds the SvelteKit static build and serves it as a SPA.
//
// The build/ directory is produced by `pnpm --dir web build`. Only a
// `.gitkeep` sentinel is tracked in git so `go build` always works; if no
// real build is present the handler serves a friendly placeholder page.
package web

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:build
var buildFS embed.FS

const placeholderHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"/><title>Flagcel</title>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<style>
body{font-family:ui-sans-serif,system-ui,-apple-system,sans-serif;background:#0a0a0c;color:#e7e7ea;display:grid;place-items:center;min-height:100vh;margin:0;padding:1.5rem}
main{max-width:32rem;text-align:center}
h1{font-weight:600;letter-spacing:-.02em}
code{background:#1a1a1e;padding:.15rem .4rem;border-radius:.3rem;font-size:.9em}
p{color:#a1a1aa;line-height:1.6}
a{color:#a78bfa}
</style></head><body><main>
<h1>Flagcel UI not built</h1>
<p>This binary was built without a compiled web UI.
Run <code>pnpm --dir web install &amp;&amp; pnpm --dir web build</code> and rebuild,
or visit <a href="/docs">/docs</a> for the API.</p>
</main></body></html>`

// Handler serves the embedded SvelteKit build, falling back to index.html for
// unknown paths so client-side routing works. When no real build is embedded,
// every request gets a placeholder page.
func Handler() http.Handler {
	sub, err := fs.Sub(buildFS, "build")
	if err != nil {
		panic(err)
	}

	indexHTML, err := fs.ReadFile(sub, "index.html")
	hasBuild := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasBuild {
			writeIndex(w, []byte(placeholderHTML))
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			writeIndex(w, indexHTML)
			return
		}

		if _, err := fs.Stat(sub, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		writeIndex(w, indexHTML)
	})
}

func writeIndex(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(body)
}
