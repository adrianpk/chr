# Slice 3: Missing CLI Actions

Status: delivered
Slicing: CLI TUI Surface Parity
Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`
Tracker: `.prj/trackers/slicing-cli-tui-surface-parity-tracker.md`
Branch: `feat/slice-3-missing-cli-actions`
PR: #206

## Purpose

Expose CLI entrypoints for the TUI operational actions that already had service
boundaries but no command-line surface.

## Delivered Behavior

- Added project metadata save command:
  `chr metadata project set`.
- Added use-case metadata create/update command:
  `chr metadata case set <slug>`.
- Added dump commands:
  `chr dump list`, `new`, `show`, `save`, `append`, and `open`.
- Added runtime quiz commands:
  `chr quiz list`, `start`, `resume`, `answer`, `skip`, `complete`,
  `cancel`, and `mode`.
- Added spec commands:
  `chr spec list` and `generate`.
- Added plan commands:
  `chr plan list`, `generate`, `regenerate`, and
  `chr plan tracking create`.
- Updated the CLI/TUI parity matrix to mark these service-backed TUI actions as
  matched.

## Implementation Notes

The new commands are thin CLI adapters over existing `service_*` methods. They
do not reimplement generation, dump persistence, quiz runtime state, metadata
save behavior, or plan tracking materialization.

Use-case metadata creation validates CLI slugs before composing workspace paths.

No new ticket or work commands were needed in this slice. The Slice 2 audit
already marked ticket creation, ticket comments, ticket status changes,
single-ticket work start, ready-loop work start, and daemon stop as matched.

## Contracts Added or Changed

- `chr metadata project set` maps to `service.SaveProjectMetadata`.
- `chr metadata case set` maps to `service.SaveCaseMetadata`.
- `chr dump *` maps to the existing dump service methods.
- `chr quiz *` maps to the runtime quiz service methods.
- `chr spec *` maps to the spec list/generation service methods.
- `chr plan *` maps to the plan list/generation/tracking service methods.

## Files of Interest

- `internal/chrapp/metadata_cli.go`
- `internal/chrapp/dump_cli.go`
- `internal/chrapp/quiz_cli.go`
- `internal/chrapp/spec_plan_cli.go`
- `internal/chrapp/app.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/cli-tui-action-matrix.md`

## Validation

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass after lint annotations for model-backed CLI calls and
  whitespace fixes.

## Risks and Follow-ups

- Model-backed commands still depend on the same generation runtime and model
  configuration as the TUI.
- Workspace graph refresh remains a known parity gap because `chr index
  rebuild` and `chr index status` only cover the ticket index.
- Slice 4 should add deeper parity tests around the new command adapters and
  the TUI service dispatch points.
