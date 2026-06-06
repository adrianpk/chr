# Slice 5: Acceptance Feature TUI Action

Status: delivered
Slicing: Separate Acceptance Feature Workflow
Plan: `ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md`
Tracker: `ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md`
Branch: `feat/slice-5-acceptance-feature-tui-action`
PR: `#298`

## Purpose

Expose acceptance feature generation from the TUI using the same service
boundary already available through the CLI.

## Delivered Behavior

The Specs view now supports `f` to generate an acceptance feature for the
selected use case. The command palette also exposes `Generate acceptance
feature` when the Specs view is active.

When generation finishes, the TUI status line includes the acceptance feature
status and either the generated `.feature` path or the feature-round note path
when more information is needed.

## Implementation Notes

The action remains in the Specs surface because it depends on an existing spec
and selected use-case scope. It does not introduce a new TUI view.

The adapter calls `GenerateAcceptanceFeature`, then reloads document and spec
state just like existing spec actions.

## Contracts Added or Changed

- Added `SpecActionGenerateFeature`.
- Added `f` key handling in the Specs view.
- Added command palette entry `generate-acceptance-feature`.
- Extended `SpecActionResult` with acceptance feature status, feature path,
  feature-round path, and reason fields.

## Files of Interest

- `internal/chctui/spec_actions.go`
- `internal/chctui/specs.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model_test.go`
- `internal/chcapp/tui_adapters_spec_plan_ticket.go`
- `internal/chcapp/tui_adapter_parity_test.go`

## Validation

- `go test ./internal/chctui -run 'SpecGenerationCommands|GeneratesSpecFromSpecsKey|GeneratesAcceptanceFeatureFromSpecsKey'` (pass)
- `go test ./internal/chcapp -run 'TUISpecAdapterGeneratesAcceptanceFeature'` (pass)
- `go test ./internal/chcapp -run 'TUISpecAdapterGeneratesAcceptanceFeature' && go test ./internal/chctui -run 'SpecGenerationCommands|GeneratesAcceptanceFeatureFromSpecsKey'` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Final user-guide and fixture validation are pending in Slice 6.
- `godog` execution remains deferred.
