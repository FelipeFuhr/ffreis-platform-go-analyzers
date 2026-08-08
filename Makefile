SHELL    := /bin/bash

BIN_DIR  ?= ./bin
GO       ?= go
GITLEAKS ?= gitleaks

COVERAGE_MIN       ?= 75
MUTATION_PACKAGES  ?= ./analyzers/...
MUTATION_THRESHOLD ?= 60

.PHONY: help build test lint fmt fmt-check ci secrets-scan-staged clean install \
        coverage-gate integration-coverage-gate mutation quality-gates

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all analyzer binaries into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/nakedgo ./cmd/nakedgo

test: ## Run unit + analysistest with race + shuffle
	$(GO) test -race -shuffle=on ./...

fmt: ## Format Go source files
	$(GO) fmt ./...

fmt-check: ## Check Go formatting without modifying files
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "gofmt: the following files need formatting:"; \
		echo "$$out"; \
		exit 1; \
	fi

lint: ## Run golangci-lint (if installed)
	@command -v golangci-lint >/dev/null && golangci-lint run ./... || echo "golangci-lint not installed; skipping"

coverage-gate: ## Run tests with coverage and fail if below COVERAGE_MIN
	@COVERAGE_MIN="$(COVERAGE_MIN)" ./scripts/hooks/check_coverage_gate.sh

integration-coverage-gate: ## Run //go:build integration tests with coverage and fail if below COVERAGE_MIN (no-op if none exist)
	@COVERAGE_MIN="$(COVERAGE_MIN)" ./scripts/hooks/check_integration_coverage_gate.sh

mutation: ## Run mutation testing with gremlins (slow — CI only)
	@which gremlins >/dev/null 2>&1 || go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
	gremlins unleash --threshold-efficacy $(MUTATION_THRESHOLD) $(MUTATION_PACKAGES)

ci: fmt-check lint test coverage-gate ## Run all CI checks locally

quality-gates: test coverage-gate integration-coverage-gate ## Run strict pre-push quality gates (required by ffreis-platform-standards lefthook complex tier)

secrets-scan-staged: ## Scan staged diff for secrets (called by lefthook pre-commit)
	@command -v $(GITLEAKS) >/dev/null 2>&1 || (echo "Missing tool: $(GITLEAKS). Install: https://github.com/gitleaks/gitleaks#installing" && exit 1)
	$(GITLEAKS) protect --staged --redact

install: ## Install the nakedgo binary into $GOPATH/bin
	$(GO) install ./cmd/nakedgo

clean: ## Remove built binaries
	rm -rf $(BIN_DIR)
