# Slice 4: Dash Agent Work Panel

Status: reviewing

## Purpose

Turn Dash into the operational surface for agent work by showing the ready
queue, current execution or loop state, recent outcomes, and direct work
actions.

## Delivered

- Added a Dash agent-work renderer with Queue, Current, and Recent panes.
- Added ready-ticket queue rendering for `ready_for_dev` tickets.
- Added Dash key actions for selected-ticket start, run-all-ready, and stop
  request.
- Added the `work-run-all-ready` command palette action.
- Wired the TUI work adapter to `service.StartAllReadyAgentWork`.
- Kept the legacy Agent and Runtime views available for detailed observability.
- Added TUI tests for Dash rendering and run-all action planning.
- Updated TUI and CLI/TUI action matrix documentation.

## Validation

- `go test ./internal/chctui ./internal/chcapp`
- `make verify`
- `git diff --check`
