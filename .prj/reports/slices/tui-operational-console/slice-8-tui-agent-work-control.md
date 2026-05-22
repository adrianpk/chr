# Slice 8: Agent Work Control

Status: delivered
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-8-tui-agent-work-control`
PR: #146

## Purpose

Wire explicit agent work controls into the TUI after read-only observability and
safe local actions are available.

This slice lets the TUI start bounded work while preserving ticket-centered
visibility, confirmation, and evidence reporting.

## Delivered Behavior

- Registers work control commands in the command palette.
- Shows a work action confirmation panel before running long-running work.
- Starts the selected ticket through the existing runtime validation and harness
  execution path.
- Starts a bounded ready-ticket loop with explicit `status=ready` and
  `max_iterations=1`.
- Requests daemon stop through the daemon state repository.
- Reloads ticket and observability state after work actions.
- Shows completion status and failure messages as visible TUI status output.
- Includes execution result paths in selected-ticket completion messages.

## Implementation Notes

The TUI owns planning and confirmation state. The app layer owns execution
because it can resolve selected tickets, validate runtime configuration, run the
harness, query eligible loop tickets, request daemon stop, and reload runtime
observability.

The ready-ticket loop is intentionally bounded to one iteration. This keeps the
TUI action explicit and reviewable while still using the existing loop execution
path.

Daemon stop is a stop request, not a destructive process kill. If daemon state is
missing, the action reports that the daemon is already stopped.

## Contracts Added or Changed

- Work commands are available from the command palette.
- Pending work actions intercept keys until confirmed or canceled.
- `enter` confirms a work action.
- `esc` cancels a work action.
- Successful work actions reload tickets and observability before rendering the
  result message.
- Failed work actions clear pending confirmation and render
  `Work action failed: ...`.

## Files of Interest

- `internal/chctui/work_actions.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model.go`
- `internal/chctui/model_test.go`
- `internal/chcapp/app.go`

## Validation

```sh
go test ./internal/chctui ./internal/chcapp
make verify
```

All commands pass.

## Risks and Follow-ups

- Ticket execution and loop execution are synchronous from the Bubble Tea update
  path. A later slice can move these actions behind asynchronous commands if the
  runtime interaction needs non-blocking progress updates.
- The loop action is intentionally narrow. Broader filter selection needs a
  dedicated TUI input surface before exposing more loop dimensions.
