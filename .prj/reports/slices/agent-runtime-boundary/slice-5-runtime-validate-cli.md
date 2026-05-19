# Slice 5: Runtime Validate CLI

Status: reviewing
Slicing: Agent Runtime Boundary
Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
Branch: `feat/slice-5-runtime-validate-cli`
PR: `#33`

## Purpose

Expose runtime readiness checks before ticket execution or agent invocation
exists.

## Delivered Behavior

`chc runtime validate` resolves the target project, loads runtime config,
computes the runtime plan, prints Docker boundary inputs, and checks required
local runtime paths. It supports `--dir`, active project resolution through
`--dev-root`, current-directory fallback, and `--dry-run`.

The command does not run Docker, execute agents, or run the configured
validation command.

## Implementation Notes

Runtime validation is split into a core checker and CLI rendering. The core
checker returns structured checks so later commands can reuse validation without
parsing CLI output.

## Contracts Added or Changed

New command:

```sh
chc runtime validate
chc runtime validate --dry-run
chc runtime validate --dir /path/to/project --dry-run
```

Core validation checks:

- workspace directory exists;
- configured Dockerfile exists and is a file;
- configured artifact directory exists and is a directory.

## Files of Interest

- `internal/chc/runtime_validate.go`
- `cmd/chc/main.go`
- `internal/chc/runtime_validate_test.go`
- `cmd/chc/main_test.go`
- `docs/reference/project-workspace.md`

## Validation

- `go test ./internal/chc ./cmd/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

Validation is intentionally local and dry-run only. Docker availability,
container build checks, and execution of `runtime.validation_command` remain
out of scope for this slicing.
