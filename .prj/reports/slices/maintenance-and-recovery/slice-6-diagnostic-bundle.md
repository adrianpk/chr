# Slice 6: Diagnostic Bundle

Status: reviewing
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-6-diagnostic-bundle`
PR: pending

## Purpose

Write a deterministic local diagnostic bundle under `.chc/maintenance/`.

This slice gives users one project-local Markdown artifact that collects current
maintenance diagnostics without copying secrets or requiring manual SQLite and
filesystem inspection.

## Delivered Behavior

- Added diagnostic bundle rendering and writing.
- Added `chc maintenance bundle`.
- Included database migration status.
- Included project registry diagnostics.
- Included execution artifact drift diagnostics.
- Included runtime artifact cleanup preview.
- Included project report metadata.
- Omitted workspace config contents, environment values, credential files, and
  token values.
- Added focused renderer, writer, and CLI tests.

## Implementation Notes

The bundle path is fixed:

```text
.chc/maintenance/diagnostic-bundle.md
```

The bundle is deterministic for a given local state. It does not include a
generation timestamp.

## Contracts Added or Changed

- Added `DiagnosticBundle`.
- Added `RenderDiagnosticBundle`.
- Added `WriteDiagnosticBundle`.
- Added `chc maintenance bundle`.
- Updated `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chc/diagnostic_bundle.go`
- `internal/chc/diagnostic_bundle_test.go`
- `internal/chcapp/maintenance.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chc -run 'Test.*DiagnosticBundle'
go test ./internal/chcapp -run 'TestMaintenanceBundle'
make verify
```

All commands pass.

## Risks and Follow-ups

- Slice 7 will add broader maintenance documentation and cross-links.
