.PHONY: help web web-dev web-install build run docker-up db-reset clean \
        migrate-up migrate-down migrate-status migrate-version \
        bench bench-cpu bench-mem loadtest pprof-cpu pprof-cpu-loaded pprof-heap

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
	@echo ""
	@echo "Performance:"
	@echo "  bench           run engine benchmarks (-benchmem)"
	@echo "  bench-cpu       run engine benchmarks with a CPU profile -> cpu.prof"
	@echo "  bench-mem       run engine benchmarks with a heap profile -> mem.prof"
	@echo "  loadtest        run k6 load test against http://localhost:8080"
	@echo "  pprof-cpu       capture CPU profile (run loadtest in parallel for non-empty output)"
	@echo "  pprof-cpu-loaded  same, but starts k6 in the background for you"
	@echo "  pprof-heap      capture heap snapshot (no load required)"

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

# --- Performance ---------------------------------------------------------

BENCH ?= .
BENCHTIME ?= 1s
PPROF_ADDR ?= http://localhost:16000
LOAD_BASE_URL ?= http://localhost:8080
LOAD_FLAG_KEY ?= new-sidebar
LOAD_SCENARIO ?= steady
LOAD_SETUP ?= true
LOAD_API_TOKEN ?=
LOAD_ADMIN_EMAIL ?= admin@localhost
LOAD_ADMIN_PASSWORD ?= secret

bench:
	go test -run='^$$' -bench='$(BENCH)' -benchtime=$(BENCHTIME) -benchmem ./internal/engine/...

bench-cpu:
	go test -run='^$$' -bench='$(BENCH)' -benchtime=$(BENCHTIME) -benchmem \
		-cpuprofile=cpu.prof -o engine.test ./internal/engine
	@echo "profile: cpu.prof  (open with: go tool pprof -http=:0 engine.test cpu.prof)"

bench-mem:
	go test -run='^$$' -bench='$(BENCH)' -benchtime=$(BENCHTIME) -benchmem \
		-memprofile=mem.prof -o engine.test ./internal/engine
	@echo "profile: mem.prof  (open with: go tool pprof -http=:0 engine.test mem.prof)"

loadtest:
	k6 run \
		-e BASE_URL=$(LOAD_BASE_URL) \
		-e FLAG_KEY=$(LOAD_FLAG_KEY) \
		-e SCENARIO=$(LOAD_SCENARIO) \
		-e SETUP=$(LOAD_SETUP) \
		-e API_TOKEN=$(LOAD_API_TOKEN) \
		-e ADMIN_EMAIL=$(LOAD_ADMIN_EMAIL) \
		-e ADMIN_PASSWORD=$(LOAD_ADMIN_PASSWORD) \
		scripts/load/eval.js

PPROF_SECONDS ?= 30

pprof-cpu:
	@echo "Sampling CPU for $(PPROF_SECONDS)s from $(PPROF_ADDR)."
	@echo "Make sure load is hitting the API (run 'make loadtest' in another shell)."
	go tool pprof -http=:2317 "$(PPROF_ADDR)/debug/pprof/profile?seconds=$(PPROF_SECONDS)"

pprof-heap:
	go tool pprof -http=:2317 "$(PPROF_ADDR)/debug/pprof/heap"

# Convenience: run load in the background while pprof samples the CPU.
# Auto-stops k6 when sampling ends.
pprof-cpu-loaded:
	@echo "Starting k6 in background, then sampling CPU for $(PPROF_SECONDS)s."
	@k6 run -e BASE_URL=$(LOAD_BASE_URL) -e FLAG_KEY=$(LOAD_FLAG_KEY) \
		-e SETUP=$(LOAD_SETUP) -e API_TOKEN=$(LOAD_API_TOKEN) \
		-e ADMIN_EMAIL=$(LOAD_ADMIN_EMAIL) -e ADMIN_PASSWORD=$(LOAD_ADMIN_PASSWORD) \
		-e SCENARIO=steady -e DURATION=$$(($(PPROF_SECONDS)+5))s \
		scripts/load/eval.js >/tmp/k6-pprof.log 2>&1 & \
		K6_PID=$$!; \
		go tool pprof -http=:2317 "$(PPROF_ADDR)/debug/pprof/profile?seconds=$(PPROF_SECONDS)"; \
		kill $$K6_PID 2>/dev/null || true
