.PHONY: build run dev test docker-up docker-down docker-logs docker-logs-frontend docker-db-only db-shell clean deploy deploy-logs deploy-url

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

# ── Deploy (Google Cloud Run, Mumbai) ────────────────────────────────
# `make deploy` builds the image from this folder and ships it.
# Secrets come from env.yaml (never committed, never uploaded — see
# .gitignore and .gcloudignore). The project is pinned so a deploy can't
# land in another project if `gcloud config` changes.
#
# max-instances 1: live matches, the queue and the WebSocket hub live in
#                  memory, so there must be exactly one server.
# concurrency 1000: every online player holds a WebSocket (one request);
#                  the default of 80 would cap the game at 80 players.
GCP_PROJECT ?= zoholedger
GCP_REGION  ?= asia-south1
SERVICE     ?= mindrace-api

deploy:
	@test -f env.yaml || (echo "env.yaml missing — it holds the production settings" && exit 1)
	go vet ./... && go test ./... -count=1
	gcloud run deploy $(SERVICE) \
		--project $(GCP_PROJECT) \
		--source . \
		--region $(GCP_REGION) \
		--port 8080 \
		--allow-unauthenticated \
		--env-vars-file env.yaml \
		--max-instances 1 \
		--concurrency 1000 \
		--timeout 3600 \
		--memory 512Mi \
		--cpu-boost

# Recent production logs.
deploy-logs:
	gcloud run services logs read $(SERVICE) --project $(GCP_PROJECT) --region $(GCP_REGION) --limit 100

# The live URL.
deploy-url:
	@gcloud run services describe $(SERVICE) --project $(GCP_PROJECT) --region $(GCP_REGION) --format='value(status.url)'

# ── Cleanup ──────────────────────────────────────────────────────────
clean:
	rm -rf bin/
	docker compose -f docker/docker-compose.yml down -v