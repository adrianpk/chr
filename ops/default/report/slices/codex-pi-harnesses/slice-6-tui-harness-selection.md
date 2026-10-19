# Slice 6: TUI Harness Selection

Status: reviewing
Delivery set: codex-pi-harnesses
Plan: [Selectable Harness Delivery](../../../plan/codex-pi-harnesses.md)
Tracker: [Selectable Harness Delivery Order](../../../tracker/codex-pi-harnesses.md)
Branch: `feat/slice-6-tui-harness-selection`
PR: pending

## Purpose

Inspect and change project worker and generation defaults from the existing TUI
project context while retaining the bindings of admitted work.

## Delivered Behavior

- Pressing `h` in Project opens a compact picker through the existing model's
  event loop and a thin application executor.
- The picker observes one shared persisted selection revision, readable role
  defaults and role-specific profile readiness. Legacy generation dispatch
  remains operation-specific.
- Retained worker bindings are separate from current defaults. The bounded
  projection shows four invocations, with more available through existing
  workflow inspection.
- Tab selects a role, arrows choose a named profile, `c` selects legacy dispatch
  and Enter/Ctrl+S saves either or both roles atomically. Esc discards unsent
  choices; `r` explicitly discards them and reloads current defaults.
- Saves bind the observed project/revision, exact choices and request identity.
  The shared application operation resolves and rechecks the local actor.
  Rejections retain visible input; unchanged retries recover committed receipts
  after a lost acknowledgement.
- The same model observes external Web changes after refresh and new work binds
  current persisted defaults. Existing admitted work retains its revision,
  status and worker binding.
- Inspection and saves perform no model call, driver launch or setup activation.
  The picker exposes named selections and public readiness, with private setup
  remaining under its existing owner.

## Implementation Notes

T6.1 maps the shared application inspection operation into public TUI values.
Reads use a ten-second context and complete asynchronously. Interaction identity
rejects late results after cancellation/reopening, and the picker owns keyboard
and mouse focus. Existing panel layout handles the compact projection; its
content scrolls independently with page keys.

T6.2 translates the displayed envelope to the same atomic configuration operation
used by the Web adapter. Actor identity is resolved on the server side and
rechecked under the shared configuration lock. Only ephemeral interaction state
lives in the TUI; operation admission continues reading persisted configuration.

One save runs at a time with copied role choices and a retained request identity.
The UI waits for acknowledgement before closing or accepting another save.
Stale/failed results preserve the original observed revision and choices rather
than silently rebasing. Fresh inspection failure after a committed change stays
recoverable through the original receipt.

The User Guide, configuration reference, action matrix and existing Unreleased
capability now describe the usable TUI controls.

## Contracts Added or Changed

- Optional TUI harness selection executor in application assembly.
- Public default/profile/readiness and retained-binding projections.
- Project-context picker with ephemeral interaction state only.
- Exact local selection save envelope through shared authority/revision/receipt
  guards, with correlated asynchronous completion.

## Files of Interest

- `internal/tui/harness_selection.go`, `model.go`, `shell.go`
- `internal/tui/project_metadata_actions.go`
- `internal/app/tui_adapters_harness.go`, `app.go`
- `internal/tui/harness_selection*_test.go`
- `internal/app/tui_adapters_harness*_test.go`

## Validation

Passed focused task checks:

- `go test ./internal/tui ./internal/app -run '^TestTUIHarnessSelection' -count=1 -timeout=3m`
- `go test -race ./internal/tui ./internal/app -run '^TestTUIHarnessSelection' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/tui ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The application journey drives real model events and production executor
assembly with prepared profiles and a live claim. It changes both roles,
drops one committed acknowledgement, retries the exact request and clears each
role independently. A separate process observes the persisted image. The old
invocation's revision/status/binding stay unchanged, while a new invocation
retains the new worker profile. Preview and cancel perform no write or launch.

A second owned process constructs independent Web/application adapters and
publishes a change while the TUI remains open. A stale keyboard save retains its
choice and original binding; explicit reload observes the external revision
without recreating the model or service. Local authority, foreign-project and
unprepared-profile refusals preserve configuration.

Model fixtures cover in-flight input exclusion, correlated late reads,
cancellation, sanitized failures and modal mouse ownership. Rendered 120×40
and 60×24 shells retain both role choices and controls, wrap long model labels
and scroll to retained bindings. Focused race instrumentation covers these
asynchronous interaction boundaries. Owned claims/subprocesses finish cleanup.

The controller records the final ordered development gate, focused application
checks, focused race checks and whitespace validation against the clean PR head
after reference follow-up commits. That exact-head evidence is separate from
these task checks.

## Risks and Follow-ups

None recorded.
