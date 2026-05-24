.PHONY: help web web-dev web-install build run docker-up db-reset clean \
        migrate-up migrate-down migrate-status migrate-version

help:
	@echo "Targets:"
	@echo "  web-install     install web dependencies (pnpm)"
	@echo "  web-dev         run Vite dev server (port 5173, proxies /api -> :8080)"
	@echo "  web             build the SvelteKit app to web/build/"
	@echo "  build           web + go build (produces a single binary with embedded UI)"
	@echo "  run             go run the server against the local Postgres in docker-compose"
	@echo "  docker-up       docker compose up (hot-reload Go, with Postgres)"
	@echo "  db-reset        wipe the local Postgres volume; migrations re-apply on next boot"
	@echo "  migrate-up      apply all pending migrations"
	@echo "  migrate-down    roll back the most recent migration"
	@echo "  migrate-status  print migration status"
	@echo "  migrate-version print the current schema version"
	@echo "  clean           remove build artifacts"

web-install:
	pnpm --dir web install

web:
	pnpm --dir web build

web-dev:
	pnpm --dir web dev

build: web
	go build -o bin/flagcel ./cmd/server

run:
	go run ./cmd/server

docker-up:
	docker compose up

db-reset:
	docker compose rm -sfv db
	docker volume rm flagcel_db-data 2>/dev/null || true
	docker compose up -d db

migrate-up:
	go run ./cmd/server migrate up

migrate-down:
	go run ./cmd/server migrate down

migrate-status:
	go run ./cmd/server migrate status

migrate-version:
	go run ./cmd/server migrate version

clean:
	rm -rf bin web/build web/.svelte-kit
