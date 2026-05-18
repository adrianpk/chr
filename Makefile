GO := go

.PHONY: fmt test lint lint-strict check verify

fmt:
	@$(GO) fmt ./...

test:
	@$(GO) test ./...

lint:
	@golangci-lint run ./...

lint-strict:
	@golangci-lint run --timeout=5m ./...

check: fmt test

verify: lint-strict test
