# === CONFIG =======================================================
PROJECT_NAME="mycel"
COVER_PROFILE="build/cover.out"

# Build
build:
	@echo "---> Building $(PROJECT_NAME)"
	go build -o build/$(PROJECT_NAME) cmd/$(PROJECT_NAME)/main.go
.PHONY: build

# === TOOLS =======================================================
# Fix go.mod and go.sum
mod-tidy:
	@echo "---> Checking module requirements"
	go mod tidy
.PHONY: mod-tidy

# Format go code
fmt:
	@echo "---> Formatting code"
	go fmt ./...
.PHONY: fmt

# Examine Go source code and reports suspicious constructs
vet:
	@echo "---> Checking Go source code"
	go vet ./...
.PHONY: vet

# Run application using linters: it runs linters in parallel, uses caching, supports yaml config, etc.
lint:
	@echo "---> Running linter"
	golangci-lint run ./... --timeout=3m
.PHONY: lint

# === DEVELOPMENT =======================================================
pre-commit: mod-tidy fmt build lint vet

# Start dependencies (Redis) via Docker and run the application locally
run:
	@if [ ! -f .env ]; then \
		echo "---> No .env found; creating one from .env.sample"; \
		cp .env.sample .env; \
		echo "Fill in .env with your values, then re-run 'make run'."; \
		exit 1; \
	fi
	@echo "---> Starting dependencies"
	docker compose up -d --wait redis
	@echo "---> Running $(PROJECT_NAME)"
	set -a && . ./.env && set +a && go run cmd/$(PROJECT_NAME)/main.go
.PHONY: run
