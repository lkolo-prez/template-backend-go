# template-backend-go — developer task runner.
# Run `make help` for an overview. Targets are POSIX-sh friendly.

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

# --- Build metadata stamped into the binary ---------------------------------
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)

BINARY     := server
BIN_DIR    := bin
PKG        := ./...
IMAGE      ?= ghcr.io/aircloud/template-backend-go
COVER_MIN  ?= 60

GOBIN      := $(shell go env GOPATH)/bin

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: setup
setup: ## Install developer tooling (lint, vuln, gosec, air, syft, cosign hints).
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/air-verse/air@latest
	@echo "Install syft + cosign from https://github.com/anchore/syft and https://github.com/sigstore/cosign"
	@command -v pre-commit >/dev/null 2>&1 && pre-commit install || echo "pre-commit not installed (pip install pre-commit)"

.PHONY: tidy
tidy: ## Tidy and verify go modules.
	go mod tidy
	go mod verify

.PHONY: lint
lint: ## Run golangci-lint (strict).
	$(GOBIN)/golangci-lint run ./...

.PHONY: format
format: ## Format code with gofmt + goimports (via golangci-lint).
	go fmt $(PKG)
	$(GOBIN)/golangci-lint run --fix ./... || true

.PHONY: format-check
format-check: ## Fail if code is not gofmt-clean.
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || (echo "Run 'make format'"; exit 1)

.PHONY: vet
vet: ## Run go vet.
	go vet $(PKG)

.PHONY: test
test: ## Run tests with the race detector.
	go test -race -count=1 $(PKG)

.PHONY: cover
cover: ## Run tests with coverage and enforce the $(COVER_MIN)% gate.
	go test -race -covermode=atomic -coverpkg=$(PKG) -coverprofile=coverage.out $(PKG)
	@go tool cover -func=coverage.out | tail -1
	@total=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $$total% (min $(COVER_MIN)%)"; \
	awk -v t="$$total" -v m="$(COVER_MIN)" 'BEGIN { exit (t+0 < m+0) }' \
		|| (echo "Coverage gate failed"; exit 1)
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: build
build: ## Build the static server binary into ./bin.
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/server

.PHONY: run
run: ## Run the server locally.
	ENVIRONMENT=development LOG_LEVEL=debug go run ./cmd/server

.PHONY: run-watch
run-watch: ## Run with live reload (air).
	$(GOBIN)/air

.PHONY: docker
docker: ## Build the container image.
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

.PHONY: security
security: ## Run govulncheck + gosec.
	$(GOBIN)/govulncheck $(PKG)
	$(GOBIN)/gosec -quiet -exclude-generated $(PKG)

.PHONY: sbom
sbom: ## Generate a CycloneDX SBOM with syft.
	syft scan dir:. -o cyclonedx-json=sbom.cdx.json
	@echo "Wrote sbom.cdx.json"

.PHONY: ci
ci: tidy format-check vet lint test cover security ## Run the full local CI pipeline.

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf $(BIN_DIR) tmp dist coverage.out coverage.html sbom.cdx.json build-errors.log
	go clean -cache -testcache
