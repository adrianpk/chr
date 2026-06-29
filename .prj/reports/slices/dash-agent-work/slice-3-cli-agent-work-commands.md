# Slice 3: CLI Agent Work Commands

Status: reviewing

## Purpose

Expose the agent-work service facade through CLI commands so users can run one
ticket or all ready tickets without using the lower-level ticket execution and
loop commands directly.

## Delivered

- Added the top-level `work` command.
- Added `chr work ticket <ticket-key>` for selected-ticket agent work.
- Added `chr work run-all` for all ready tickets using the shared agent-work
  selection contract.
- Added `chr work run-all --dry-run` to preview the selected queue without
  creating execution or loop records.
- Added CLI tests for selected-ticket execution, run-all execution, dry-run
  preview, and no eligible tickets.
- Updated command reference documentation.

## Validation

- `go test ./internal/chrapp`
- `make verify`
- `git diff --check`
