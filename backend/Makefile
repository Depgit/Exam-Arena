# ── Local Development & Builds ────────────────────────────────────────
build:
	go build -o bin/exam-arena ./cmd/server

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/exam-arena-linux-amd64 ./cmd/server

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o bin/exam-arena-linux-arm64 ./cmd/server

build-lambda:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-w -s" -o bin/bootstrap ./cmd/lambda

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