# Slice 5: Runtime Artifact Cleanup Preview

Status: reviewing
Slicing: Maintenance and Recovery
Plan: `.prj/plans/slicing-maintenance-and-recovery-plan.md`
Tracker: `.prj/trackers/slicing-maintenance-and-recovery-tracker.md`
Branch: `feat/slice-5-runtime-artifact-cleanup-preview`
PR: #130

## Purpose

Preview cleanup of stale runtime artifact directories without deleting
anything.

This slice helps users understand which runtime run directories are no longer
referenced by execution records before any destructive cleanup command exists.

## Delivered Behavior

- Added runtime artifact cleanup preview diagnostics.
- Compared runtime artifact leaf directories against execution record
  `artifact_run_dir` values for the resolved project.
- Added `chc maintenance execution cleanup-preview`.
- Kept cleanup preview read-only.
- Added focused domain and CLI tests.
- Documented cleanup preview behavior and safety boundaries.

## Implementation Notes

The preview scans the resolved runtime artifact directory from workspace
configuration. Missing artifact directories produce an empty preview.

Only leaf directories are reported as cleanup candidates. Parent grouping
directories are not reported directly.

## Contracts Added or Changed

- Added `RuntimeArtifactCleanupPreviewService`.
- Added `chc maintenance execution cleanup-preview`.
- Updated `docs/reference/maintenance.md`.

## Files of Interest

- `internal/chc/runtime_artifact_cleanup_preview.go`
- `internal/chc/runtime_artifact_cleanup_preview_test.go`
- `internal/chcapp/maintenance.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/maintenance.md`

## Validation

```sh
go test ./internal/chc -run 'TestRuntimeArtifactCleanupPreview'
go test ./internal/chcapp -run 'TestMaintenanceExecutionCleanupPreview'
make verify
```

All commands pass.

## Risks and Follow-ups

- This slice intentionally does not delete candidates.
- Diagnostic bundles remain deferred to Slice 6.
