BINARY  := forge
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -s -w -X github.com/DanielTso/forge/cmd.version=$(VERSION) -X github.com/DanielTso/forge/cmd.commit=$(COMMIT)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary into ./bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

.PHONY: install
install: ## Install forge into $GOBIN / $GOPATH/bin
	go install -ldflags "$(LDFLAGS)" .

.PHONY: test
test: ## Run all tests (including generated-project integration tests)
	go test -race -cover ./...

.PHONY: test-short
test-short: ## Run fast tests only
	go test -short ./...

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: snapshot
snapshot: ## Build release archives locally with goreleaser (nothing is published)
	goreleaser release --snapshot --clean --skip=publish

.PHONY: check
check: fmt vet lint test ## Format, vet, lint and test

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin dist
