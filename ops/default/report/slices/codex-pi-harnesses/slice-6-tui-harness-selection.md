# Slice 6: TUI Harness Selection

Status: drafting
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
- Keyboard previews and cancellation do not mutate configuration or work.
  Inspection performs no model call, driver launch or setup activation.

## Implementation Notes

T6.1 maps the shared application inspection operation into public TUI values.
Reads use a ten-second context and complete asynchronously. Interaction identity
rejects late results after cancellation/reopening, and the picker owns keyboard
and mouse focus. Existing panel layout handles the compact projection; its
content scrolls independently with page keys.

T6.2 revision-bound saves, local actor authority and refresh/conflict acceptance
remain in progress.

## Contracts Added or Changed

- Optional TUI harness selection executor in application assembly.
- Public default/profile/readiness and retained-binding projections.
- Project-context picker with ephemeral interaction state only.

## Files of Interest

- `internal/tui/harness_selection.go`, `model.go`, `shell.go`
- `internal/tui/project_metadata_actions.go`
- `internal/app/tui_adapters_harness.go`, `app.go`
- `internal/tui/harness_selection_test.go`
- `internal/app/tui_adapters_harness_test.go`

## Validation

Passed focused T6.1 checks:

- `go test ./internal/tui ./internal/app -run '^TestTUIHarnessSelection' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --allow-serial-runners --fix --timeout=5m ./internal/tui ./internal/app`
- `git diff --check`

The real application fixture reads prepared profiles and a claimed invocation
through model events. Preview/cancel preserves the invocation revision and
binding and does not launch the owned driver. Model fixtures cover correlated
late reads, cancellation, sanitized setup failure and modal mouse focus.

## Risks and Follow-ups

Complete T6.2 before opening the canonical PR. Saves and cross-process refresh
acceptance remain pending.
