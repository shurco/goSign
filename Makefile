.DEFAULT_GOAL := help

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
BUN           ?= bun
BIN_DIR       ?= bin
BINARY        ?= $(BIN_DIR)/goSign
PKG           ?= ./...

COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE   ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------
# Backend (Go)
# ------------------------------------------------------------------

.PHONY: build
build: ## Build the main binary into ./bin
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/goSign

.PHONY: run
run: ## Run the server in dev mode
	$(GO) run ./cmd/goSign

.PHONY: test
test: ## Run Go unit tests (short mode, no external services)
	$(GO) test -short -race -count=1 $(PKG)

.PHONY: test-all
test-all: ## Run all Go tests, including integration tests
	$(GO) test -race -count=1 $(PKG)

.PHONY: cover
cover: ## Generate coverage report (coverage.out + coverage.html)
	$(GO) test -short -race -covermode=atomic -coverprofile=coverage.out $(PKG)
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKG)

.PHONY: lint
lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run ./...

.PHONY: fmt
fmt: ## Format Go code with gofmt + goimports
	$(GO) fmt $(PKG)
	@command -v goimports >/dev/null 2>&1 && goimports -w -local github.com/shurco/gosign . || echo "goimports not installed"

.PHONY: tidy
tidy: ## Sync go.mod / go.sum
	$(GO) mod tidy

# ------------------------------------------------------------------
# Frontend (Vue + Vite, via Bun)
# ------------------------------------------------------------------

.PHONY: web-install
web-install: ## Install frontend dependencies
	cd web && $(BUN) install

.PHONY: web-dev
web-dev: ## Start the Vite dev server
	cd web && $(BUN) run dev

.PHONY: web-build
web-build: ## Build the production frontend bundle
	cd web && $(BUN) run build

.PHONY: web-test
web-test: ## Run Vitest unit tests
	cd web && $(BUN) x vitest run

.PHONY: web-typecheck
web-typecheck: ## Run vue-tsc type checking
	cd web && $(BUN) run typecheck

.PHONY: web-lint
web-lint: ## Run ESLint on the frontend
	cd web && $(BUN) run lint

# ------------------------------------------------------------------
# Combined quality gates
# ------------------------------------------------------------------

.PHONY: check
check: vet test web-typecheck web-test ## Run go vet, Go tests, typecheck, and frontend tests

.PHONY: ci
ci: lint check ## Run the full CI quality gate (lint + check)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out coverage.html web/dist
