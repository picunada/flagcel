.PHONY: help web web-dev web-install build run docker-up clean

help:
	@echo "Targets:"
	@echo "  web-install  install web dependencies (pnpm)"
	@echo "  web-dev      run Vite dev server (port 5173, proxies /api -> :8080)"
	@echo "  web          build the SvelteKit app to web/build/"
	@echo "  build        web + go build (produces a single binary with embedded UI)"
	@echo "  run          go run the server against the local Postgres in docker-compose"
	@echo "  docker-up    docker compose up (hot-reload Go, with Postgres)"
	@echo "  clean        remove build artifacts"

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

clean:
	rm -rf bin web/build web/.svelte-kit
