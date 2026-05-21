# Slice 2: Project Registry Diagnostics

Status: reviewing
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-2-project-registry-diagnostics`
PR: pending

## Purpose

Expose registry drift across registered projects.

This slice lets users inspect registered project paths, workspace presence,
enabled state, and active project selection without mutating registry state.

## Delivered Behavior

- Added project registry diagnostic model.
- Added active project detection.
- Added path existence and `.chc/` workspace checks.
- Added `chc maintenance registry status`.
- Added stable line-oriented output for project diagnostics.
- Documented registry diagnostic behavior.

## Implementation Notes

The diagnostics command is read-only. It does not select, disable, remove, or
repair projects.

The command reports these statuses:

- `ok`;
- `disabled`;
- `missing_path`;
- `missing_workspace`.

## Contracts Added or Changed

- Added `ProjectRegistryDiagnostic`.
- Added `ProjectRegistryDiagnosticStatus`.
- Added `ProjectRegistryDiagnosticsService`.
- Added `chc maintenance registry status`.
- Updated `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chc/project_registry_diagnostics.go`
- `internal/chc/project_registry_diagnostics_test.go`
- `internal/chcapp/maintenance.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chc -run 'TestProjectRegistryDiagnostics'
go test ./internal/chcapp -run 'TestMaintenanceRegistryStatus|TestMaintenanceDBStatus'
make verify
```

All commands pass.

## Risks and Follow-ups

- This slice reports registry drift only. Repair operations remain deferred to
  Slice 3.
