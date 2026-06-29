# Slice 4: Execution Artifact Drift

Status: delivered
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-4-execution-artifact-drift`
PR: #129

## Purpose

Detect drift between execution records and runtime artifacts.

This slice gives users a read-only way to identify incomplete or damaged
execution evidence without inspecting SQLite records or runtime directories
manually.

## Delivered Behavior

- Added project-scoped execution record listing.
- Added execution artifact drift diagnostics.
- Added `chr maintenance execution artifacts`.
- Reported missing run directories, result artifacts, logs, validation logs, and
  execution reports.
- Kept artifact diagnostics read-only.
- Added focused repository, service, and CLI tests.
- Documented execution artifact drift behavior.

## Implementation Notes

Execution report paths are derived from each execution record's run directory
with the existing `ExecutionReportPath` helper.

The command resolves the project through `--dir` and inspects execution records
for that project id. It does not mutate execution records or artifact files.

## Contracts Added or Changed

- Extended `ExecutionRepository` with `ListProjectExecutions`.
- Added `ExecutionArtifactDriftService`.
- Added `chr maintenance execution artifacts`.
- Updated `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chr/execution_repository.go`
- `internal/chr/execution_artifact_drift.go`
- `internal/chr/execution_artifact_drift_test.go`
- `internal/chrapp/maintenance.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chr -run 'TestExecution(Repository|ArtifactDrift)'
go test ./internal/chrapp -run 'TestMaintenanceExecutionArtifacts'
make verify
```

All commands pass.

## Risks and Follow-ups

- Cleanup preview remains deferred to Slice 5.
- Diagnostic bundles remain deferred to Slice 6.
