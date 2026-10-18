SHELL := /bin/bash

APP     := api
BIN_DIR := bin

.PHONY: help run build test test-integration lint fmt vet tidy sqlc migrate-up migrate-down migrate-status migrate-create docker-up docker-down clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: ## Run the API server (migrations run automatically by default)
	go run ./cmd/api

build: ## Build the API binary into bin/
	mkdir -p $(BIN_DIR)
	go build -trimpath -o $(BIN_DIR)/$(APP) ./cmd/api

test: ## Run unit tests
	go test ./...

test-integration: ## Run PostgreSQL integration tests (requires TEST_DATABASE_URL)
	TEST_DATABASE_URL="$${TEST_DATABASE_URL}" go test ./internal/adapter/postgres/...

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format Go code
	gofmt -w cmd internal db

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy module dependencies
	go mod tidy

sqlc: ## Regenerate type-safe query code
	sqlc generate

migrate-up: ## Apply all pending migrations (goose CLI)
	./scripts/migrate.sh up

migrate-down: ## Roll back one migration (goose CLI)
	./scripts/migrate.sh down

migrate-status: ## Show migration status (goose CLI)
	./scripts/migrate.sh status

migrate-create: ## Create a migration: make migrate-create NAME=add_tags
	./scripts/migrate.sh create $(NAME) sql

docker-up: ## Start local PostgreSQL (localhost:5433)
	docker compose up -d postgres

docker-down: ## Stop local PostgreSQL
	docker compose down

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) tmp
