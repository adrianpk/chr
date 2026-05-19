# Slice 3: Runtime Init CLI

Status: delivered
Slicing: Agent Runtime Boundary
Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
Branch: `feat/slice-3-runtime-init-cli`
PR: `#31`

## Purpose

Expose runtime asset materialization for existing managed projects without
rerunning full workspace initialization.

## Delivered Behavior

`chc runtime init` resolves the target project and installs missing runtime
defaults. It supports explicit `--dir`, active project resolution through
`--dev-root`, current-directory fallback, and `--dry-run`.

Existing runtime files are skipped rather than overwritten.

## Implementation Notes

Runtime initialization uses a focused core service that materializes only
workspace profile files under `runtime/`. The CLI prints a runtime-specific
result header and the same `create` / `skip existing` path records used by
workspace initialization.

## Contracts Added or Changed

New command:

```sh
chc runtime init
chc runtime init --dry-run
chc runtime init --dir /path/to/project
```

The command requires a managed project context and writes only missing runtime
assets.

## Files of Interest

- `internal/chc/runtime_init.go`
- `cmd/chc/main.go`
- `internal/chc/runtime_init_test.go`
- `cmd/chc/main_test.go`
- `docs/reference/project-workspace.md`

## Validation

- `go test ./internal/chc ./cmd/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

Runtime init currently materializes the standard runtime asset set. The profile
templates are equivalent today; profile-aware runtime initialization can be
added later if profiles diverge.
