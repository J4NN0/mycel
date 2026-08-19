# === CONFIG =======================================================
PROJECT_NAME="mycel"
COVER_PROFILE="build/cover.out"

# === BUILD =======================================================
build:
	@echo "---> Building $(PROJECT_NAME)"
	go build -o build/$(PROJECT_NAME) cmd/$(PROJECT_NAME)/main.go
.PHONY: build

# === INSTALL =======================================================
# Check what is missing, build and install the agent, write its config and pull the model.
install:
	@./install/install.sh
.PHONY: install

# Report which dependencies are missing, without installing anything
doctor:
	@./install/doctor.sh
.PHONY: doctor

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

# === TEST =======================================================
test:
	@echo "---> Running tests"
	go test ./...
.PHONY: test

# Run the tests and report coverage per function
test-cover:
	@echo "---> Running tests with coverage"
	go test -coverprofile=$(COVER_PROFILE) ./...
	go tool cover -func=$(COVER_PROFILE)
.PHONY: test-cover

# === DEVELOPMENT =======================================================
pre-commit: mod-tidy fmt build test lint vet

# === DOCS =======================================================
DOCS_VENV=build/docs-venv

# Silence the Material for MkDocs banner about the upcoming MkDocs 2.0
export NO_MKDOCS_2_WARNING=1

# Install the docs toolchain in its own virtualenv, so a system-wide mkdocs
# (e.g. Homebrew's, which ships without the material theme) can't get in the way
$(DOCS_VENV)/bin/mkdocs: docs/requirements.txt
	@echo "---> Installing docs dependencies"
	python3 -m venv $(DOCS_VENV)
	$(DOCS_VENV)/bin/pip install --quiet --upgrade pip
	$(DOCS_VENV)/bin/pip install --quiet -r docs/requirements.txt

# Serve the documentation locally at http://localhost:8000
docs: $(DOCS_VENV)/bin/mkdocs
	@echo "---> Serving docs"
	$(DOCS_VENV)/bin/mkdocs serve
.PHONY: docs

# Build the documentation into site/
docs-build: $(DOCS_VENV)/bin/mkdocs
	@echo "---> Building docs"
	$(DOCS_VENV)/bin/mkdocs build --strict
.PHONY: docs-build
