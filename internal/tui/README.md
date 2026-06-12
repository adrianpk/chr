# `internal/tui`

## Purpose

`internal/tui` implements the local terminal UI for Chrysocoord project
operation.

## Boundary

The package owns terminal state, command palette routing, keyboard interaction,
view projections, and confirmation flows. It delegates application actions
through executor interfaces supplied by `internal/app` and should not own
durable persistence or reusable domain workflows.

## API

The main boundary is `Run`, which starts the Bubble Tea event loop from
`Options`. Exported view, action, and executor types are adapter contracts used
by `internal/app` tests and TUI assembly.

## Behavior

The model renders workspace context, tickets, dumps, quizzes, specs, plans,
agent work, and observability state. Actions are planned in the UI, confirmed by
the user, then delegated to executor interfaces.

## Validation

Package changes should be validated with:

```sh
go test ./internal/tui
```
