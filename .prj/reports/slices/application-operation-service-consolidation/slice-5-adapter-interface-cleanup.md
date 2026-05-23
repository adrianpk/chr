# Slice 5: Adapter Interface Cleanup

Status: reviewing
Slicing: Application Operation Service Consolidation
Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`
Tracker: `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`
Branch: `feat/slice-5-adapter-interface-cleanup`
PR: pending

## Purpose

Clean up adapter-facing interfaces after moving operations into the shared
application service.

This keeps TUI-specific mappings out of the main app bootstrap file and adds a
guardrail that service files stay adapter-neutral.

## Delivered Behavior

- Moved all TUI executor wrapper structs and mapping methods out of
  `internal/chcapp/app.go`.
- Added `internal/chcapp/tui_adapters.go` as the dedicated TUI adapter mapping
  file.
- Kept `app.go` focused on command bootstrap, TUI option loading, and shared
  TUI state loaders.
- Preserved existing TUI executor interfaces and runtime behavior.
- Added a boundary test that prevents `service*.go` files from importing
  `internal/chctui`.

## Implementation Notes

The service implementation remains central and split by capability files. The
TUI wrappers still exist because `internal/chctui` owns TUI-facing interfaces
and result models, but those wrappers now live in a file named for the adapter
role.

The boundary test scans production `service*.go` files and excludes tests. This
keeps adapter-neutral service code from drifting back toward TUI-specific
contracts.

## Contracts Added or Changed

- Added `internal/chcapp/tui_adapters.go` for TUI adapter mappings.
- Added `TestServiceFilesDoNotImportTUIAdapter`.
- Removed TUI executor wrapper definitions from `internal/chcapp/app.go`.

## Files of Interest

- `internal/chcapp/tui_adapters.go`
- `internal/chcapp/app.go`
- `internal/chcapp/service_boundary_test.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- CLI convergence remains for the next slice; this slice only cleans up the
  adapter boundary after TUI migration.
