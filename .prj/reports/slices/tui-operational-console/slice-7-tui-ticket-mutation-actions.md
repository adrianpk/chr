# Slice 7: Ticket Mutation Actions

Status: delivered
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-7-tui-ticket-mutation-actions`
PR: #145

## Purpose

Expose explicit selected-ticket operations from the TUI without allowing
free-form opaque Markdown edits.

This slice adds ticket-specific confirmation flows and routes durable writes
through the existing ticket domain operations.

## Delivered Behavior

- Registers selected-ticket mutation commands in the command palette.
- Shows a ticket action confirmation panel before mutation.
- Supports canceling a pending ticket action with escape.
- Adds a standard comment to the selected ticket through the ticket commenter.
- Moves the selected ticket to `ready`, `in_progress`, or `reviewing` through
  the ticket status updater.
- Reloads ticket state after successful mutations.
- Reports ticket action validation and execution errors as visible status
  messages.
- Registers checklist toggle and checklist item creation actions, but reports a
  validation error because no safe checklist item writer exists yet.

## Implementation Notes

The TUI owns action planning and confirmation state. The app layer owns
execution because it can refresh the ticket index, resolve the selected ticket
by key, execute repository-backed mutations, and reload project tickets.

Comment creation uses author `tui` and body `TUI action recorded.`. Status
commands are explicit command palette entries rather than an inferred next-state
transition.

Checklist operations remain non-mutating. They are discoverable and fail with a
clear validation message until a domain operation can safely address checklist
items by identity.

## Contracts Added or Changed

- Ticket mutation commands operate on the currently selected ticket.
- Pending ticket actions intercept keys until confirmed or canceled.
- `enter` confirms a ticket action.
- `esc` cancels a ticket action.
- Successful ticket mutations reload the board/detail state before rendering the
  success message.
- Failed or unsupported ticket actions clear the pending confirmation and render
  `Ticket action failed: ...`.

## Files of Interest

- `internal/chrtui/ticket_actions.go`
- `internal/chrtui/command_palette.go`
- `internal/chrtui/model.go`
- `internal/chrtui/model_test.go`
- `internal/chrapp/app.go`

## Validation

```sh
go test ./internal/chrtui ./internal/chrapp
make verify
```

All commands pass.

## Risks and Follow-ups

- The default comment body is intentionally fixed because the current TUI does
  not have a free-form text entry surface for mutation payloads.
- Checklist mutation should be added only after a safe domain writer can target
  checklist items deterministically.
