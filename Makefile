MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build clean fmt help install lint run test uninstall

BINARY_NAME := pbasic

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

clean: ## Remove build artifacts
	@go clean
	@rm -f bin/$(BINARY_NAME)
	@mkdir -p bin
	@touch bin/.gitkeep


install: build ## Install binary (~/.local/bin or /usr/local/bin)
	@./.make/install.sh

uninstall: ## Remove installed binary
	@./.make/uninstall.sh
