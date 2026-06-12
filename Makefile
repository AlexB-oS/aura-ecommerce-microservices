# Convenience commands for the AURA platform.
# Requires GNU make (preinstalled on macOS/Linux; on Windows use Git Bash + make,
# WSL, or just run the underlying `docker compose` / `go` commands directly).

COMPOSE := docker compose

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## ----------------------------------------------------------------- stack ---
.PHONY: up
up: ## Build and start the whole stack (http://localhost:8080)
	$(COMPOSE) up --build -d
	@echo "AURA is up -> http://localhost:8080"

.PHONY: down
down: ## Stop the stack (keeps data volumes)
	$(COMPOSE) down

.PHONY: clean
clean: ## Stop the stack AND delete data volumes
	$(COMPOSE) down -v

.PHONY: logs
logs: ## Tail logs from all services
	$(COMPOSE) logs -f

.PHONY: ps
ps: ## Show running containers
	$(COMPOSE) ps

.PHONY: restart
restart: ## Rebuild and restart just the user service
	$(COMPOSE) up -d --build user-service

.PHONY: scale-user
scale-user: ## Run 3 replicas of the user service behind the gateway
	$(COMPOSE) up -d --scale user-service=3 user-service

## ------------------------------------------------------------------ go -----
.PHONY: test
test: ## Run Go unit tests
	go test ./... -count=1

.PHONY: tidy
tidy: ## Resolve dependencies and write go.sum
	go mod tidy

.PHONY: fmt
fmt: ## Format all Go code
	go fmt ./...

.PHONY: vet
vet: ## Static analysis
	go vet ./...

## ----------------------------------------------------------------- web -----
.PHONY: web-dev
web-dev: ## Run the Next.js storefront in dev mode (hot reload) on :3000
	cd web && npm install && npm run dev
