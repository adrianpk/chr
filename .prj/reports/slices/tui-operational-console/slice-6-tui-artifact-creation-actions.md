# Slice 6: Safe Artifact Creation Actions

Status: reviewing
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-6-tui-artifact-creation-actions`
PR: pending

## Purpose

Add explicit confirmation flows for low-risk artifact creation from the TUI.

This slice moves the command palette from read-only helpers into bounded write
actions while keeping all durable writes behind confirmation and existing domain
repositories.

## Delivered Behavior

- Registers artifact creation commands in the command palette.
- Shows a confirmation panel before durable artifact actions.
- Supports canceling a pending action with escape.
- Creates default task tickets through the ticket repository.
- Creates draft TUI authoring sessions through the authoring session repository.
- Reloads tickets, documents, and authoring sessions after successful writes.
- Reports action validation or execution errors as visible status messages.
- Keeps context-aware quiz and draft-file creation routed through explicit
  validation messages when provider-backed generation or promotion context is
  not available from the TUI action.

## Implementation Notes

The TUI owns action planning, confirmation state, and status rendering. The app
layer owns artifact execution because it already has the resolved project
context, user environment, and repository loading helpers.

Ticket creation uses the configured ticket key prefix and creates a normal task
ticket titled `TUI created ticket` with project area `tui`.

Authoring session creation creates a draft session with mode `tui`, target
`draft`, and topic `TUI artifact creation`.

Context-aware quiz creation is registered and confirmable, but the executor
returns a validation error instead of invoking provider-backed generation from a
synchronous TUI update path. Draft-file creation similarly reports the existing
promotion requirement rather than creating untracked draft files.

## Contracts Added or Changed

- Artifact commands are available from the command palette.
- Pending artifact actions intercept keys until confirmed or canceled.
- `enter` confirms an artifact action.
- `esc` cancels an artifact action.
- Successful write actions reload visible TUI state before returning a success
  message.
- Failed or unsupported write actions clear the pending confirmation and render
  `Action failed: ...`.

## Files of Interest

- `internal/chctui/artifact_actions.go`
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

- Context-aware quiz generation remains better suited to an asynchronous or
  CLI-backed flow because the current provider path can require model work.
- Draft-file creation should stay tied to quiz promotion or another provenance
  source instead of writing standalone draft artifacts from the TUI.
