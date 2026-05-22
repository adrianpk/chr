# Slice 1: CLI Output Readability

Status: reviewing
Slicing: UX Operation Surface
Plan: `.prj/plans/slicing-ux-operation-surface-plan.md`
Tracker: `.prj/trackers/slicing-ux-operation-surface-tracker.md`
Branch: `feat/slice-1-cli-output-readability`
PR: pending

## Purpose

Make common CLI output easier to scan without removing the command-line surface
needed for automation, tests, and runbooks.

## Delivered Behavior

- Runtime validation now prints project-relative paths for project-local files.
- Runtime validation checks are grouped under `checks:` and no longer repeat
  `ok` at the start and end of each successful check.
- Authoring session list output now includes headers.
- Authoring session creation output is compact and groups generated files under
  a `files:` section.
- Static quiz creation output avoids printing a second full session-created
  block and reports the generated quiz path relative to the project.

## Implementation Notes

- Added small CLI output helpers in `internal/chcapp/output.go`.
- Kept structured key/value output compatible with existing command parsing
  where practical.
- Preserved absolute paths when a path is outside the managed project.
- Updated tests that intentionally assert the CLI output contract.

## Contracts Added or Changed

- Human runtime validation output prefers project-relative paths.
- Runtime validation path checks use a compact `checks:` section.
- `chc author session list` starts with `id`, `status`, `target`, and `topic`
  headers.
- `chc author quiz new` reports `quiz: created <path>` for non-dry-run output
  and keeps `dry_run: true` for previews.

## Files of Interest

- `internal/chcapp/output.go`
- `internal/chcapp/runtime_proving.go`
- `internal/chcapp/author.go`
- `internal/chcapp/cli_test.go`
- `docs/how-to/full-system-lab-runbook.md`
- `docs/reference/authoring-sessions.md`
- `docs/reference/project-workspace.md`
- `docs/reference/runtime-config.md`

## Validation

- `go test ./internal/chcapp` - pass
- `make verify` - pass

## Risks and Follow-ups

- Some older docs and historical planning records still preserve pre-slice
  output examples as historical context.
- A later TUI slice should reuse the same domain semantics instead of parsing
  this human CLI output.
