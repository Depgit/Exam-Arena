.PHONY: build run dev test docker-up docker-down docker-logs docker-logs-frontend docker-db-only db-shell clean

# ── Local Development ────────────────────────────────────────────────
build:
	go build -o bin/exam-arena ./cmd/server

run: build
	./bin/exam-arena

dev:
	go run ./cmd/server

test:
	go test ./... -v -race -count=1

# ── Docker ───────────────────────────────────────────────────────────
# Starts Postgres, the Go API (http://localhost:8080) and the React UI
# (http://localhost:3000). See docker/docker-compose.yml for the
# VITE_* build args if the API is not reachable at localhost:8080.
docker-up:
	docker compose -f docker/docker-compose.yml up --build -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-logs:
	docker compose -f docker/docker-compose.yml logs -f app

docker-logs-frontend:
	docker compose -f docker/docker-compose.yml logs -f frontend

docker-db-only:
	docker compose -f docker/docker-compose.yml up postgres -d

# ── Database ─────────────────────────────────────────────────────────
db-shell:
	docker exec -it examarena-db psql -U examarena -d examarena

# ── Cleanup ──────────────────────────────────────────────────────────
clean:
	rm -rf bin/
	docker compose -f docker/docker-compose.yml down -v