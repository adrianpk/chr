# Slice 6: Agent Work Filters And Ordering

Status: reviewing

## Purpose

Expose agent-work ticket selection filters and ordering through the shared
facade, CLI, and Dash.

## Delivered

- Added `AgentWorkSelectionOptions` for shared status, filter, order, and loop
  limit selection.
- Wired `chc work run-all` to accept status, metadata filters, and order flags.
- Added Dash priority filtering and queue ordering controls.
- Passed Dash run-all selection state into the TUI work action executor.
- Added service, CLI, and TUI tests for filtered selection and ordered queues.
- Updated command, TUI, and CLI/TUI action matrix references.

## Validation

- `go test ./internal/chcapp ./internal/chctui`
- `make verify`
- `git diff --check`
