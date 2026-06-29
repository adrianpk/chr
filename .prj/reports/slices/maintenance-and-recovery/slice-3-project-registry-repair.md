# Slice 3: Project Registry Repair

Status: delivered
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-3-project-registry-repair`
PR: #128

## Purpose

Add explicit repair operations for stale project registry entries.

This slice lets users disable stale registrations or update a registered project
path without editing SQLite directly.

## Delivered Behavior

- Added project disable support to the registry.
- Added project path update support to the registry.
- Added `chr maintenance registry disable`.
- Added `chr maintenance registry set-path`.
- Added `--dry-run` support for both repair commands.
- Required project names for repair operations.
- Required `set-path` targets to be initialized workspaces with `.chr/`.
- Documented registry repair behavior.

## Implementation Notes

Repair commands mutate only the project registry row. They do not mutate ticket
files, runtime artifacts, execution records, or active project selection.

`set-path` validates the target workspace before updating the registry.

## Contracts Added or Changed

- Extended `ProjectRegistry` with `DisableProject`.
- Extended `ProjectRegistry` with `UpdateProjectPath`.
- Added `chr maintenance registry disable`.
- Added `chr maintenance registry set-path`.
- Updated `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chr/project_registry.go`
- `internal/chr/project_registry_test.go`
- `internal/chrapp/maintenance.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chr -run 'TestProjectRegistry'
go test ./internal/chrapp -run 'TestMaintenanceRegistry'
make verify
```

All commands pass.

## Risks and Follow-ups

- Registry repair is intentionally explicit and does not auto-repair drift.
- Execution artifact drift detection remains deferred to Slice 4.
