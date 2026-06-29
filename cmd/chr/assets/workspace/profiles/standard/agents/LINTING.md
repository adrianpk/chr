# Linting

## Baseline Checks

Prefer project-owned commands when available.

Common Go checks:

```sh
go fmt ./...
go test ./...
go vet ./...
```

If configured:

```sh
golangci-lint run ./...
```

## Rules

- Run fast local checks before handoff.
- Use stricter checks before merge when available.
- Report exact commands and results.
- Do not skip tests for behavior changes.
