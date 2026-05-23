SHELL := /bin/bash

BIN_DIR ?= ./bin
GO      ?= go

.PHONY: help build test lint clean install

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all analyzer binaries into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/nakedgo ./cmd/nakedgo

test: ## Run unit + analysistest with race + shuffle
	$(GO) test -race -shuffle=on ./...

lint: ## Run golangci-lint (if installed)
	@command -v golangci-lint >/dev/null && golangci-lint run ./... || echo "golangci-lint not installed; skipping"

install: ## Install the nakedgo binary into $GOPATH/bin
	$(GO) install ./cmd/nakedgo

clean: ## Remove built binaries
	rm -rf $(BIN_DIR)
