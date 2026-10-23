# `internal/tui`

## Purpose

`internal/tui` implements the local terminal UI for chr project
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
agent work, and observability state. Manual work actions retain their inspection
and confirmation flow. The guided screen sends current-step input, controls and
original-document queries to the shared application owner; that owner progresses
within the initiating development intent and native policy. The adapter owns
only drafts, frozen request bindings, scrolling and asynchronous response fences.
It does not persist or orchestrate an independent guided workflow.

## Validation

Package changes should be validated with:

```sh
go test ./internal/tui
```
