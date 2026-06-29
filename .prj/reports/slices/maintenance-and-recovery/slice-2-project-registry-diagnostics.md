# Slice 2: Project Registry Diagnostics

Status: delivered
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-2-project-registry-diagnostics`
PR: #127

## Purpose

Expose registry drift across registered projects.

This slice lets users inspect registered project paths, workspace presence,
enabled state, and active project selection without mutating registry state.

## Delivered Behavior

- Added project registry diagnostic model.
- Added active project detection.
- Added path existence and `.chr/` workspace checks.
- Added `chr maintenance registry status`.
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
- Added `chr maintenance registry status`.
- Updated `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chr/project_registry_diagnostics.go`
- `internal/chr/project_registry_diagnostics_test.go`
- `internal/chrapp/maintenance.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chr -run 'TestProjectRegistryDiagnostics'
go test ./internal/chrapp -run 'TestMaintenanceRegistryStatus|TestMaintenanceDBStatus'
make verify
```

All commands pass.

## Risks and Follow-ups

- This slice reports registry drift only. Repair operations remain deferred to
  Slice 3.
