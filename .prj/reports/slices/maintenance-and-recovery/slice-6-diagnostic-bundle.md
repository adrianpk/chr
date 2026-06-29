# Slice 6: Diagnostic Bundle

Status: delivered
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-6-diagnostic-bundle`
PR: #131

## Purpose

Write a deterministic local diagnostic bundle under `.chr/maintenance/`.

This slice gives users one project-local Markdown artifact that collects current
maintenance diagnostics without copying secrets or requiring manual SQLite and
filesystem inspection.

## Delivered Behavior

- Added diagnostic bundle rendering and writing.
- Added `chr maintenance bundle`.
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
.chr/maintenance/diagnostic-bundle.md
```

The bundle is deterministic for a given local state. It does not include a
generation timestamp.

## Contracts Added or Changed

- Added `DiagnosticBundle`.
- Added `RenderDiagnosticBundle`.
- Added `WriteDiagnosticBundle`.
- Added `chr maintenance bundle`.
- Updated `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chr/diagnostic_bundle.go`
- `internal/chr/diagnostic_bundle_test.go`
- `internal/chrapp/maintenance.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chr -run 'Test.*DiagnosticBundle'
go test ./internal/chrapp -run 'TestMaintenanceBundle'
make verify
```

All commands pass.

## Risks and Follow-ups

- Slice 7 will add broader maintenance documentation and cross-links.
