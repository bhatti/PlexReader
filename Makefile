# PlexReader — Makefile
# Targets:
#   Dev:    make dev-backend  make dev-ui  make dev (both in split panes)
#   Test:   make test         make backend-test  make ui-analyze
#   Build:  make build        make backend-build  make ui-build
#   Docker: make docker-up    make docker-down   make docker-build  make docker-logs
#   Misc:   make proto        make tidy   make clean   make help

.PHONY: help \
        backend-build backend-test backend-test-v backend-run tidy \
        ui-deps ui-build ui-run ui-test ui-analyze \
        dev dev-backend dev-ui \
        build test \
        docker-build docker-up docker-up-d docker-down docker-logs docker-ps \
        docker-backend docker-ui docker-restart \
        publish \
        proto lint-proto \
        version release \
        clean

# ─── Tool detection ──────────────────────────────────────────────────────────
BACKEND_DIR := backend
UI_DIR      := ui
BACKEND_BIN := $(BACKEND_DIR)/bin/plexreader

# Version: read base from VERSION file, append short git hash.
VERSION_BASE := $(shell cat VERSION 2>/dev/null | tr -d '[:space:]' || echo "0.1.1")
GIT_HASH     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION      := $(VERSION_BASE)-$(GIT_HASH)

FLUTTER ?= flutter
ifeq ($(shell which $(FLUTTER) 2>/dev/null),)
  ifneq ($(wildcard $(HOME)/flutter/bin/flutter),)
    FLUTTER := $(HOME)/flutter/bin/flutter
  endif
endif

DOCKER_COMPOSE := docker compose
ifeq ($(shell docker compose version 2>/dev/null),)
  DOCKER_COMPOSE := docker-compose
endif

# Backend API URL used when running Flutter locally
DEV_API_URL ?= http://localhost:8080

# ─── Help ────────────────────────────────────────────────────────────────────
help: ## Show all available targets
	@printf "\033[1mPlexReader — available targets\033[0m\n\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@printf "\n"

# ─── Backend ─────────────────────────────────────────────────────────────────
backend-build: ## Compile backend binary (CGO + FTS5)
	mkdir -p $(BACKEND_DIR)/bin
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go build -tags fts5 \
	  -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/plexreader ./cmd/server
	@echo "→ $(BACKEND_BIN) ($(VERSION))"

backend-test: ## Run all Go tests with race detector
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go test -tags fts5 -race -timeout 60s ./...

backend-test-v: ## Run Go tests verbose
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go test -tags fts5 -race -v -timeout 60s ./...

backend-test-cover: ## Run Go tests with HTML coverage report
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go test -tags fts5 -race -timeout 60s \
	  -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "→ $(BACKEND_DIR)/coverage.html"

backend-run: ## Start backend server on :8080 (foreground, hot-reloadable)
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go run -tags fts5 \
	  -ldflags="-X main.Version=$(VERSION)" ./cmd/server

tidy: ## Tidy Go module dependencies
	cd $(BACKEND_DIR) && go mod tidy

# ─── Flutter UI ──────────────────────────────────────────────────────────────
ui-deps: ## Install Flutter dependencies
	cd $(UI_DIR) && $(FLUTTER) pub get

ui-build: ui-deps ## Build Flutter web app for production
	cd $(UI_DIR) && $(FLUTTER) build web --release \
	  --dart-define=API_BASE_URL=

ui-run: ui-deps ## Run Flutter web in Chrome with hot reload (dev mode)
	cd $(UI_DIR) && $(FLUTTER) run -d chrome \
	  --dart-define=API_BASE_URL=$(DEV_API_URL)

ui-run-port: ui-deps ## Run Flutter web in Chrome on a fixed port (port 3001)
	cd $(UI_DIR) && $(FLUTTER) run -d chrome --web-port 3001 \
	  --dart-define=API_BASE_URL=$(DEV_API_URL)

ui-test: ui-deps ## Run Flutter widget/unit tests
	cd $(UI_DIR) && $(FLUTTER) test

ui-analyze: ui-deps ## Run Flutter static analysis
	cd $(UI_DIR) && $(FLUTTER) analyze

# ─── Dev (local, no Docker) ──────────────────────────────────────────────────
dev-backend: ## Start backend in background, tail its logs
	@echo "Starting backend on :8080 ($(VERSION))..."
	@mkdir -p $(BACKEND_DIR)/data
	cd $(BACKEND_DIR) && CGO_ENABLED=1 go run -tags fts5 \
	  -ldflags="-X main.Version=$(VERSION)" ./cmd/server

dev-ui: ui-deps ## Start Flutter UI pointing at local backend
	cd $(UI_DIR) && $(FLUTTER) run -d chrome --web-port 3001 \
	  --dart-define=API_BASE_URL=$(DEV_API_URL)

dev: ## Show how to run dev stack (backend + UI need separate terminals)
	@printf "\033[1mDev mode — open two terminal tabs:\033[0m\n\n"
	@printf "  \033[36mTerminal 1 — Backend:\033[0m\n"
	@printf "    make dev-backend\n\n"
	@printf "  \033[36mTerminal 2 — Flutter UI:\033[0m\n"
	@printf "    make dev-ui\n\n"
	@printf "  Backend health:  http://localhost:8080/healthz\n"
	@printf "  UI (Chrome):     http://localhost:3001\n"
	@printf "  Swagger:         http://localhost:8080/swagger/\n\n"
	@printf "  Override backend URL:  DEV_API_URL=http://myhost:8080 make dev-ui\n\n"

# ─── Combined build / test ───────────────────────────────────────────────────
build: backend-build ui-build ## Build backend binary + Flutter web release

test: backend-test ui-analyze ## Run all backend tests + Flutter analysis (CI gate)
	@echo ""
	@echo "All checks passed ✓"

# ─── Docker (production-like) ────────────────────────────────────────────────
docker-build: ## Build both Docker images
	VERSION=$(VERSION) $(DOCKER_COMPOSE) build

docker-build-nc: ## Build Docker images without cache
	VERSION=$(VERSION) $(DOCKER_COMPOSE) build --no-cache

docker-up: docker-build ## Build and start all services (foreground)
	VERSION=$(VERSION) $(DOCKER_COMPOSE) up

docker-up-d: docker-build ## Build and start all services (detached)
	VERSION=$(VERSION) $(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "  Running detached:"
	@echo "    Backend:  http://localhost:8080/healthz"
	@echo "    UI:       http://localhost:3000"
	@echo ""
	@echo "  make docker-logs    — tail logs"
	@echo "  make docker-down    — stop everything"

docker-down: ## Stop and remove containers
	$(DOCKER_COMPOSE) down

docker-down-v: ## Stop containers AND remove named volumes (resets DB)
	$(DOCKER_COMPOSE) down -v

docker-logs: ## Tail logs from all services
	$(DOCKER_COMPOSE) logs -f

docker-logs-backend: ## Tail backend logs only
	$(DOCKER_COMPOSE) logs -f backend

docker-logs-ui: ## Tail frontend logs only
	$(DOCKER_COMPOSE) logs -f frontend

docker-ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

docker-restart: ## Restart all services without rebuilding
	$(DOCKER_COMPOSE) restart

docker-backend: ## Rebuild and restart backend only
	$(DOCKER_COMPOSE) up -d --build backend

docker-ui: ## Rebuild and restart frontend only
	$(DOCKER_COMPOSE) up -d --build frontend

docker-shell-backend: ## Open a shell in the running backend container
	$(DOCKER_COMPOSE) exec backend sh

# ─── Publish (Docker Hub) ────────────────────────────────────────────────────
DOCKER_REPO ?= plexobject

publish: ## Build and push :latest images to Docker Hub (must be logged in)
	docker build -f $(BACKEND_DIR)/Dockerfile -t $(DOCKER_REPO)/plexreader-backend:latest \
	  --build-arg VERSION=$(VERSION) .
	docker build -f $(UI_DIR)/Dockerfile -t $(DOCKER_REPO)/plexreader-ui:latest .
	docker push $(DOCKER_REPO)/plexreader-backend:latest
	docker push $(DOCKER_REPO)/plexreader-ui:latest
	@echo ""
	@echo "  Published:"
	@echo "    $(DOCKER_REPO)/plexreader-backend:latest"
	@echo "    $(DOCKER_REPO)/plexreader-ui:latest"

# ─── Version / Release ───────────────────────────────────────────────────────
version: ## Show current version string
	@echo $(VERSION)

release: ## Bump patch version, tag it, and show the new version
	@set -e; \
	BASE=$$(cat VERSION | tr -d '[:space:]'); \
	MAJOR=$$(echo $$BASE | cut -d. -f1); \
	MINOR=$$(echo $$BASE | cut -d. -f2); \
	PATCH=$$(echo $$BASE | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_BASE="$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	NEW_VERSION="$$NEW_BASE-$(GIT_HASH)"; \
	echo "$$NEW_BASE" > VERSION; \
	echo "Bumped: $$BASE → $$NEW_BASE ($$NEW_VERSION)"

# ─── Proto ───────────────────────────────────────────────────────────────────
proto: ## Regenerate Go code from proto files (requires buf)
	buf generate
	@echo "Proto generation complete"

lint-proto: ## Lint proto files
	buf lint

# ─── Clean ───────────────────────────────────────────────────────────────────
clean: ## Remove build artifacts (bin/, Flutter build/, coverage)
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(UI_DIR)/build
	rm -f $(BACKEND_DIR)/coverage.out $(BACKEND_DIR)/coverage.html
	@echo "Clean complete"

clean-docker: ## Remove Docker images for this project
	docker rmi plexreader-backend:latest plexreader-ui:latest 2>/dev/null || true
	@echo "Docker images removed"
