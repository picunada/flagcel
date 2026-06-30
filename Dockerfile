# Multi-target image definitions for the Flagcel server.
#
# The Go binary embeds the SvelteKit dashboard (web/embed.go -> //go:embed
# all:build), the SQL migrations, and the OpenAPI spec, so the runtime image only
# needs the static binary plus CA certificates (provided by distroless/static).
#
# This is the from-source build used by contributors, the production
# docker-compose example, and reproducible local builds. Releases use the
# `release` target, which copies the GoReleaser-built binary instead.

# --- Target: dev hot-reload environment ------------------------------------
FROM golang:1.26-alpine AS dev
WORKDIR /app
RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
# evalcore is a local-replace module; its go.mod must be present for the
# module graph before `go mod download` runs.
COPY evalcore/go.mod evalcore/go.sum ./evalcore/
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# --- Target: release image from a prebuilt GoReleaser binary ----------------
FROM gcr.io/distroless/static:nonroot AS release
COPY flagcel /flagcel
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/flagcel"]

# --- Stage: build the embedded dashboard -----------------------------------
FROM node:22-alpine AS web
WORKDIR /web
RUN corepack enable && corepack prepare pnpm@10 --activate
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# --- Stage: build the Go binary with the dashboard embedded -----------------
FROM golang:1.26-alpine AS build
WORKDIR /app
RUN apk add --no-cache git
# sqlc generates the DB access layer (internal/store/postgres/sqlcgen is
# gitignored); install it to regenerate before building.
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
COPY go.mod go.sum ./
# evalcore is a local-replace module; its go.mod must be present for the module
# graph before `go mod download` runs.
COPY evalcore/go.mod evalcore/go.sum ./evalcore/
RUN go mod download
COPY . .
# The build stage's dashboard output replaces any web/build copied from context.
COPY --from=web /web/build ./web/build
# Regenerate sqlcgen from the committed queries.sql + migrations.
RUN sqlc generate
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /flagcel ./cmd/server

# --- Target: default from-source runtime image ------------------------------
FROM gcr.io/distroless/static:nonroot AS runtime
COPY --from=build /flagcel /flagcel
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/flagcel"]
