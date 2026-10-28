GO := go
GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT := $(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
DEV_ROOT ?= _workspace
CHR_DEV_HOME := $(DEV_ROOT)/chr
CHR_DEV_CONFIG := $(DEV_ROOT)/config/chr
GO_BUILD_CACHE := $(CURDIR)/$(DEV_ROOT)/go-build-cache
GOLANGCI_LINT_CACHE := $(CURDIR)/$(DEV_ROOT)/golangci-lint-cache
BIN_DIR := bin
BIN := $(BIN_DIR)/chr
COVERAGE_OUT ?= /tmp/chr-coverage.out
RUN_ARGS ?= --help
QA_CASE ?= blog
QA_REF ?= 001
GATE_BASE ?= origin/main
GATE_HEAD ?= origin/dev

.PHONY: fmt test test-ci test-full test-race nightly test-coverage lint lint-strict check verify check-go-style check-package-readmes dev-dirs run build build-run install clean-dev gate-dev-main qa-reset qa-test qa-smoke qa-gate qa-reset-blog qa-test-blog qa-smoke-blog qa-reset-tictactoe qa-test-tictactoe qa-smoke-tictactoe qa-native-integration qa-ticket-native-workflow

fmt:
	@$(GO) fmt ./...

test:
	@$(GO) test -short -timeout=3m ./...

test-ci:
	@python3 -m unittest discover -s scripts/qa -p 'test_nightly.py'

test-full:
	@$(GO) test -count=1 -timeout=90m ./...

test-race:
	@$(GO) test -race -count=1 -timeout=90m ./...

nightly:
	@python3 scripts/qa/nightly.py $(GO)

test-coverage:
	@$(GO) test -coverprofile=$(COVERAGE_OUT) ./...
	@$(GO) tool cover -func=$(COVERAGE_OUT)

lint:
	@scripts/check-go-func-signatures.sh
	@$(GO) run ./scripts/goquality
	@scripts/check-package-readmes.sh
	@$(GOLANGCI_LINT) run ./...

lint-strict:
	@scripts/check-go-func-signatures.sh
	@$(GO) run ./scripts/goquality
	@scripts/check-package-readmes.sh
	@$(GOLANGCI_LINT) run --timeout=5m ./...

check-go-style:
	@scripts/check-go-func-signatures.sh
	@$(GO) run ./scripts/goquality

check-package-readmes:
	@scripts/check-package-readmes.sh

check: fmt check-go-style test

verify: lint-strict test test-ci build

dev-dirs:
	@mkdir -p $(CHR_DEV_HOME)
	@mkdir -p $(CHR_DEV_CONFIG)

run: dev-dirs
	@$(GO) run ./cmd/chr $(RUN_ARGS)

build:
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN) ./cmd/chr

build-run: build
	@$(BIN) $(RUN_ARGS)

install:
	@$(GO) install ./cmd/chr

clean-dev:
	@rm -rf $(DEV_ROOT)

gate-dev-main:
	@mkdir -p $(GO_BUILD_CACHE) $(GOLANGCI_LINT_CACHE)
	@GOCACHE=$(GO_BUILD_CACHE) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) $(GO) run ./cmd/chr gate dev-main --base $(GATE_BASE) --head $(GATE_HEAD)

qa-reset:
	@scripts/qa/reset.sh $(QA_CASE) $(QA_REF)

qa-test:
	@scripts/qa/test.sh $(QA_CASE)

qa-smoke:
	@scripts/qa/smoke.sh $(QA_CASE) $(QA_REF)

qa-gate:
	@scripts/qa/gate.sh

qa-ticket-native-workflow:
	@python3 scripts/qa/ticket-native-workflow.py $(GO)

qa-native-integration:
	@CHR_NATIVE_INTEGRATION_QA=1 $(GO) test ./internal/core/workflow -run '^(TestNative(ForgejoMethods|GitHubMethods|HTTPSMethods|PolicyDenials|PRChanges|TransportLoss)|TestWorkflowGit(Atomic|Advertisement|HTTPS|SSH|SSHUnknown)|TestNativeCapabilityDenials)$$' -count=1 -v

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
