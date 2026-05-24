# Slice 1: Dump Service Contract

Status: delivered
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-1-dump-service-contract`
PR: `#175`

## Purpose

Introduce the service-owned dump operation boundary before adding TUI dump
workflows.

This keeps dump path resolution, scope interpretation, and workspace-index
projection outside the TUI layer.

## Delivered Behavior

- The application service can list dumps from the workspace index.
- Dump scope is explicit: empty case slug means project-wide; non-empty case
  slug means use-case-scoped.
- The service exposes request/result contracts for future create, append, save,
  and open dump operations.
- The canonical dump directory is resolved through the service/core boundary as
  `.chc/dumps`.
- The generic TUI document loader recognizes `.chc/dumps` as an active dump
  source while preserving `.prj/dumps` compatibility.

## Implementation Notes

- Added dump contracts in `internal/chcapp` so TUI adapters can depend on an
  executor boundary instead of implementing dump logic.
- Added a core `chc.DumpDirectory` helper for the canonical project-local dump
  path.
- Dump listing is derived from `WorkspaceIndex` dump nodes rather than parsing
  Markdown in the TUI.
- Append, save, open, and create request contracts are present for later slices;
  the behavior implementation is intentionally left to the matching workflow
  slices.

## Contracts Added or Changed

- Added `DumpScope`, `DumpView`, `DumpListResult`, `DumpCreateRequest`,
  `DumpAppendRequest`, `DumpSaveRequest`, `DumpOpenRequest`, and
  `DumpOperationResult` in `internal/chcapp`.
- Added `service.ListDumps()`.
- Added `chc.DumpDirectory(projectDir)`.
- Changed active TUI document discovery to include `.chc/dumps`.

## Files of Interest

- `internal/chc/dump.go`
- `internal/chcapp/service_dump.go`
- `internal/chcapp/service_test.go`
- `internal/chctui/documents.go`
- `internal/chctui/model_test.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui ./internal/chc`
- `make verify`

## Risks and Follow-ups

- `.prj/dumps` remains as a compatibility read source until a later slice or
  cleanup decision removes it.
- Create, append, save, and open dump behavior is contract-only in this slice
  and will be implemented in later slices.
