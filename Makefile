.PHONY: build run dev test docker-up docker-down clean

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
docker-up:
	docker compose -f docker/docker-compose.yml up --build -d

docker-down:
	docker compose -f docker/docker-compose.yml down

docker-logs:
	docker compose -f docker/docker-compose.yml logs -f app

docker-db-only:
	docker compose -f docker/docker-compose.yml up postgres -d

# ── Database ─────────────────────────────────────────────────────────
db-shell:
	docker exec -it examarena-db psql -U examarena -d examarena

# ── Cleanup ──────────────────────────────────────────────────────────
clean:
	rm -rf bin/
	docker compose -f docker/docker-compose.yml down -v