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
  `chc metadata project set`.
- Added use-case metadata create/update command:
  `chc metadata case set <slug>`.
- Added dump commands:
  `chc dump list`, `new`, `show`, `save`, `append`, and `open`.
- Added runtime quiz commands:
  `chc quiz list`, `start`, `resume`, `answer`, `skip`, `complete`,
  `cancel`, and `mode`.
- Added spec commands:
  `chc spec list` and `generate`.
- Added plan commands:
  `chc plan list`, `generate`, `regenerate`, and
  `chc plan tracking create`.
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

- `chc metadata project set` maps to `service.SaveProjectMetadata`.
- `chc metadata case set` maps to `service.SaveCaseMetadata`.
- `chc dump *` maps to the existing dump service methods.
- `chc quiz *` maps to the runtime quiz service methods.
- `chc spec *` maps to the spec list/generation service methods.
- `chc plan *` maps to the plan list/generation/tracking service methods.

## Files of Interest

- `internal/chcapp/metadata_cli.go`
- `internal/chcapp/dump_cli.go`
- `internal/chcapp/quiz_cli.go`
- `internal/chcapp/spec_plan_cli.go`
- `internal/chcapp/app.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/cli-tui-action-matrix.md`

## Validation

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass after lint annotations for model-backed CLI calls and
  whitespace fixes.

## Risks and Follow-ups

- Model-backed commands still depend on the same generation runtime and model
  configuration as the TUI.
- Workspace graph refresh remains a known parity gap because `chc index
  rebuild` and `chc index status` only cover the ticket index.
- Slice 4 should add deeper parity tests around the new command adapters and
  the TUI service dispatch points.
