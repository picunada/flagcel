# Production image for the Flagcel server.
#
# The Go binary embeds the SvelteKit dashboard (web/embed.go -> //go:embed
# all:build), the SQL migrations, and the OpenAPI spec, so the runtime image only
# needs the static binary plus CA certificates (provided by distroless/static).
#
# This is the from-source build used by contributors, the production
# docker-compose example, and reproducible local builds. Releases use the slim
# Dockerfile.release, which copies the GoReleaser-built binary instead.

# --- Stage 1: build the embedded dashboard ---------------------------------
FROM node:22-alpine AS web
WORKDIR /web
RUN corepack enable && corepack prepare pnpm@10 --activate
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# --- Stage 2: build the Go binary with the dashboard embedded --------------
FROM golang:1.26-alpine AS build
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
# evalcore is a local-replace module; its go.mod must be present for the module
# graph before `go mod download` runs (mirrors Dockerfile.dev).
COPY evalcore/go.mod evalcore/go.sum ./evalcore/
RUN go mod download
COPY . .
# The build stage's dashboard output replaces any web/build copied from context.
COPY --from=web /web/build ./web/build
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /flagcel ./cmd/server

# --- Stage 3: minimal runtime ----------------------------------------------
FROM gcr.io/distroless/static:nonroot
COPY --from=build /flagcel /flagcel
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/flagcel"]
