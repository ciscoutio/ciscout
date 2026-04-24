.PHONY: help dev build test fmt vet tidy lint clean db-up db-down db-reset db-migrate db-migrate-new

help: ## Show available targets
	@grep -E '^[a-z]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Run with live-reload (air) or plain go run
	@command -v air > /dev/null && air || go run ./cmd/ciscout

build: ## Compile binary to bin/ciscout
	@mkdir -p bin
	go build -o bin/ciscout ./cmd/ciscout

test: ## Run tests with race detector
	go test ./... -race -count=1

fmt: ## Format code with gofmt
	gofmt -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod and go.sum
	go mod tidy

lint: ## Run staticcheck if installed
	@command -v staticcheck > /dev/null && staticcheck ./... || echo "staticcheck not installed; skipping"

clean: ## Remove build artifacts
	rm -rf bin/
	go clean -testcache

db-up: ## Start Postgres container and wait until healthy
	docker compose up -d --wait postgres

db-down: ## Stop Postgres container
	docker compose down

db-reset: ## Reset database (remove volume and recreate)
	docker compose down -v
	$(MAKE) db-up db-migrate

GOOSE := $(shell command -v goose 2>/dev/null || echo $(shell go env GOPATH)/bin/goose)

db-migrate: ## Apply database migrations
	@test -x "$(GOOSE)" || (echo "goose not found; install with: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	$(GOOSE) -dir db/migrations postgres "$$DATABASE_URL" up

db-migrate-new: ## Create new migration (usage: make db-migrate-new NAME=migration_name)
	@test -x "$(GOOSE)" || (echo "goose not found; install with: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	@test -n "$(NAME)" || (echo "NAME not provided; usage: make db-migrate-new NAME=migration_name" && exit 1)
	$(GOOSE) -dir db/migrations create $(NAME) sql
