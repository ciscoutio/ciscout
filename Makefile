.PHONY: help dev build test fmt vet tidy lint clean

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
