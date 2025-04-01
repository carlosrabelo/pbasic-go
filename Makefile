MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build clean fmt help install install-system lint quality run test uninstall uninstall-system

BINARY_NAME := pbasic

export BINARY_NAME

help: ## Show available targets
	@echo "pbasic - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@./.make/build.sh

test: ## Run tests
	@./.make/test.sh

run: build ## Build and run the REPL
	@./bin/$(BINARY_NAME)

lint: ## Run linter
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

quality: fmt lint ## Run all quality checks

clean: ## Remove build artifacts
	@go clean
	@rm -f bin/$(BINARY_NAME)

install: ## Install to ~/.local/bin (requires prior make build)
	@./.make/install.sh

install-system: ## Install to /usr/local/bin (requires prior make build; sudo only for copy)
	@SYSTEM=1 ./.make/install.sh

uninstall: ## Remove from ~/.local/bin (sudo only if needed)
	@./.make/uninstall.sh

uninstall-system: ## Remove from /usr/local/bin (sudo only if needed)
	@SYSTEM=1 ./.make/uninstall.sh
