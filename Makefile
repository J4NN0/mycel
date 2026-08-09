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

# === RUN =======================================================
run:
	@if [ ! -f .env ]; then \
		echo "---> No .env found; creating one from .env.sample"; \
		cp .env.sample .env; \
		echo "Fill in .env with your values, then re-run 'make run'."; \
		exit 1; \
	fi

	@echo "---> Starting dependencies"
	docker compose up -d --wait redis redis-commander
	@echo "---> Redis UI available at http://localhost:8081"

	@echo "---> Running $(PROJECT_NAME)"
	set -a && . ./.env && set +a && go run cmd/$(PROJECT_NAME)/main.go
.PHONY: run

install:
	@echo "---> Installing $(PROJECT_NAME)"
	go install ./cmd/$(PROJECT_NAME)

	@mkdir -p $(HOME)/.config/$(PROJECT_NAME)

	@if [ -f $(HOME)/.config/$(PROJECT_NAME)/.env ]; then \
		echo "---> Config already exists at $(HOME)/.config/$(PROJECT_NAME)/.env; leaving it untouched"; \
	elif [ -f .env ]; then \
		echo "---> Copying .env to $(HOME)/.config/$(PROJECT_NAME)/.env"; \
		cp .env $(HOME)/.config/$(PROJECT_NAME)/.env; \
	else \
		echo "---> No .env found; copying .env.sample to $(HOME)/.config/$(PROJECT_NAME)/.env"; \
		echo "Fill in $(HOME)/.config/$(PROJECT_NAME)/.env with your values before running 'mycel'."; \
		cp .env.sample $(HOME)/.config/$(PROJECT_NAME)/.env; \
	fi

	@echo "---> Run 'mycel' to start the agent"
.PHONY: install
