# Kodu — Go build
BINARY      := kodu
PKG         := ./cmd/kodu
DIST        := dist
MODULE      := github.com/uxname/kodu

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(MODULE)/internal/buildinfo.Version=$(VERSION) \
	-X $(MODULE)/internal/buildinfo.Commit=$(COMMIT) \
	-X $(MODULE)/internal/buildinfo.Date=$(DATE)

# tree-sitter requires CGO.
export CGO_ENABLED := 1

.PHONY: build
build: ## Build the binary into dist/
	@mkdir -p $(DIST)
	go build -ldflags="$(LDFLAGS)" -o $(DIST)/$(BINARY) $(PKG)

.PHONY: install
install: ## Install the binary into $GOPATH/bin
	go install -ldflags="$(LDFLAGS)" $(PKG)

.PHONY: test
test: ## Run the tests
	go test ./...

.PHONY: lint
lint: ## Static analysis
	gofmt -l . | tee /dev/stderr | (! read)
	go vet ./...
	golangci-lint run

.PHONY: tidy
tidy: ## Tidy up go.mod/go.sum
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(DIST)

.PHONY: help
help: ## Show the list of targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
