# gab — build & quality tasks for the gab-helper Go binary.
# Run `make` (or `make help`) to list targets.

BIN     := bin/gab-helper
PKG     := ./cmd/gab-helper
GOFILES := internal cmd

.DEFAULT_GOAL := help

## build: compile gab-helper into bin/ (where Claude Code puts it on the PATH)
.PHONY: build
build:
	go build -o $(BIN) $(PKG)

## test: run the full test suite
.PHONY: test
test:
	go test ./...

## vet: run go vet across all packages
.PHONY: vet
vet:
	go vet ./...

## fmt: gofmt-format the source in place
.PHONY: fmt
fmt:
	gofmt -w $(GOFILES)

## fmt-check: fail if any source is not gofmt-clean (used by CI)
.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l $(GOFILES)); \
	if [ -n "$$unformatted" ]; then \
		echo "these files are not gofmt-clean (run 'make fmt'):"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

## check: the full quality gate — fmt-check, vet, test
.PHONY: check
check: fmt-check vet test

## tidy: sync go.mod/go.sum with the source
.PHONY: tidy
tidy:
	go mod tidy

## clean: remove build artifacts
.PHONY: clean
clean:
	rm -rf bin

## help: list available targets
.PHONY: help
help:
	@echo "gab make targets:"
	@grep -E '^## [a-z-]+:' $(MAKEFILE_LIST) | sed 's/## /  /'
