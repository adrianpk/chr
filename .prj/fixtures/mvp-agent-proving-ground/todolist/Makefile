GO := go

.PHONY: compile fmt-check lint test verify

compile:
	@$(GO) test -run '^$$' ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

lint: fmt-check
	@$(GO) vet ./...

test:
	@$(GO) test ./...

verify: compile lint test
