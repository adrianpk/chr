GO := go
DEV_ROOT ?= _workspace
CHC_DEV_HOME := $(DEV_ROOT)/chc
CHC_DEV_CONFIG := $(DEV_ROOT)/config/chc
BIN_DIR := bin
BIN := $(BIN_DIR)/chc
COVERAGE_OUT ?= /tmp/chc-coverage.out
RUN_ARGS ?= --help

.PHONY: fmt test test-coverage lint lint-strict check verify dev-dirs run build build-run install clean-dev

fmt:
	@$(GO) fmt ./...

test:
	@$(GO) test ./...

test-coverage:
	@$(GO) test -coverprofile=$(COVERAGE_OUT) ./...
	@$(GO) tool cover -func=$(COVERAGE_OUT)

lint:
	@golangci-lint run ./...

lint-strict:
	@golangci-lint run --timeout=5m ./...

check: fmt test

verify: lint-strict test

dev-dirs:
	@mkdir -p $(CHC_DEV_HOME)
	@mkdir -p $(CHC_DEV_CONFIG)

run: dev-dirs
	@$(GO) run ./cmd/chc $(RUN_ARGS)

build:
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN) ./cmd/chc

build-run: build
	@$(BIN) $(RUN_ARGS)

install:
	@$(GO) install ./cmd/chc

clean-dev:
	@rm -rf $(DEV_ROOT)
