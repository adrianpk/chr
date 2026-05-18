GO := go
DEV_ROOT ?= _workspace
BIN_DIR := $(DEV_ROOT)/bin
BIN := $(BIN_DIR)/chc
RUN_ARGS ?= --help

.PHONY: fmt test test-coverage lint lint-strict check verify dev-dirs run build build-run install clean-dev

fmt:
	@$(GO) fmt ./...

test:
	@$(GO) test ./...

test-coverage:
	@mkdir -p $(DEV_ROOT)/coverage
	@$(GO) test -coverprofile=$(DEV_ROOT)/coverage/coverage.out ./...
	@$(GO) tool cover -func=$(DEV_ROOT)/coverage/coverage.out

lint:
	@golangci-lint run ./...

lint-strict:
	@golangci-lint run --timeout=5m ./...

check: fmt test

verify: lint-strict test

dev-dirs:
	@mkdir -p $(DEV_ROOT)/config/chc
	@mkdir -p $(DEV_ROOT)/data/chc
	@mkdir -p $(DEV_ROOT)/cache/chc
	@mkdir -p $(DEV_ROOT)/state/chc
	@mkdir -p $(BIN_DIR)

run: dev-dirs
	@$(GO) run ./cmd/chc $(RUN_ARGS)

build: dev-dirs
	@$(GO) build -o $(BIN) ./cmd/chc

build-run: build
	@$(BIN) $(RUN_ARGS)

install:
	@$(GO) install ./cmd/chc

clean-dev:
	@rm -rf $(DEV_ROOT)
