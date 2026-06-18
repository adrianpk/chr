GO := go
DEV_ROOT ?= _workspace
CHC_DEV_HOME := $(DEV_ROOT)/chc
CHC_DEV_CONFIG := $(DEV_ROOT)/config/chc
GO_BUILD_CACHE := $(CURDIR)/$(DEV_ROOT)/go-build-cache
GOLANGCI_LINT_CACHE := $(CURDIR)/$(DEV_ROOT)/golangci-lint-cache
BIN_DIR := bin
BIN := $(BIN_DIR)/chc
COVERAGE_OUT ?= /tmp/chc-coverage.out
RUN_ARGS ?= --help
QA_CASE ?= blog
QA_REF ?= 001
GATE_BASE ?= origin/main
GATE_HEAD ?= origin/dev

.PHONY: fmt test test-coverage lint lint-strict check verify check-go-style check-package-readmes dev-dirs run build build-run install clean-dev gate-dev-main qa-reset qa-test qa-smoke qa-gate qa-reset-blog qa-test-blog qa-smoke-blog qa-reset-tictactoe qa-test-tictactoe qa-smoke-tictactoe

fmt:
	@$(GO) fmt ./...

test:
	@$(GO) test ./...

test-coverage:
	@$(GO) test -coverprofile=$(COVERAGE_OUT) ./...
	@$(GO) tool cover -func=$(COVERAGE_OUT)

lint:
	@scripts/check-go-func-signatures.sh
	@scripts/check-package-readmes.sh
	@golangci-lint run ./...

lint-strict:
	@scripts/check-go-func-signatures.sh
	@scripts/check-package-readmes.sh
	@golangci-lint run --timeout=5m ./...

check-go-style:
	@scripts/check-go-func-signatures.sh

check-package-readmes:
	@scripts/check-package-readmes.sh

check: fmt check-go-style test

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

gate-dev-main:
	@mkdir -p $(GO_BUILD_CACHE) $(GOLANGCI_LINT_CACHE)
	@GOCACHE=$(GO_BUILD_CACHE) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) $(GO) run ./cmd/chc gate dev-main --base $(GATE_BASE) --head $(GATE_HEAD)

qa-reset:
	@scripts/qa/reset.sh $(QA_CASE) $(QA_REF)

qa-test:
	@scripts/qa/test.sh $(QA_CASE)

qa-smoke:
	@scripts/qa/smoke.sh $(QA_CASE) $(QA_REF)

qa-gate:
	@scripts/qa/gate.sh

qa-reset-blog:
	@$(MAKE) qa-reset QA_CASE=blog QA_REF=001

qa-test-blog:
	@$(MAKE) qa-test QA_CASE=blog

qa-smoke-blog:
	@$(MAKE) qa-smoke QA_CASE=blog QA_REF=001

qa-reset-tictactoe:
	@$(MAKE) qa-reset QA_CASE=tictactoe QA_REF=003

qa-test-tictactoe:
	@$(MAKE) qa-test QA_CASE=tictactoe

qa-smoke-tictactoe:
	@$(MAKE) qa-smoke QA_CASE=tictactoe QA_REF=003
