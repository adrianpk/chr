# Slice 3: Index Load And Refresh Semantics

Status: reviewing
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-3-index-refresh-semantics`
PR: pending

## Purpose

Make the structural workspace index part of startup and refresh flows while
keeping the filesystem as the source of truth.

## Delivered Behavior

- Load the structural workspace index at TUI startup from the current
  filesystem-backed workspace graph.
- Refresh the structural index whenever the workspace refresh operation rebuilds
  the graph.
- Expose index health from the service refresh result.
- Add index change detection for added, changed, and removed nodes based on
  source hash, modified time, and size.
- Keep existing `WorkspaceGraph` consumers working during the transition.
- Preserve direct-file-edit behavior through rebuild-on-refresh semantics.

## Implementation Notes

This slice does not introduce persistent cache files. Cache read/write behavior
belongs to Slice 4.

The TUI stores `workspaceIndex` beside `workspaceGraph`. If refresh results do
not include an index, the TUI derives one from the returned graph to preserve
compatibility with existing tests and executors.

## Contracts Added or Changed

- Added `WorkspaceIndexHealth`.
- Added `WorkspaceIndexChanges`.
- Added `WorkspaceIndex.Health`.
- Added `WorkspaceIndex.ChangesSince`.
- Extended workspace refresh results with index, health, and change fields.
- Added TUI options/state for `WorkspaceIndex`.

## Files of Interest

- `internal/chc/workspace_index.go`
- `internal/chc/workspace_index_test.go`
- `internal/chcapp/app.go`
- `internal/chcapp/service_workspace.go`
- `internal/chctui/model.go`
- `internal/chctui/workspace_refresh.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-3-index-refresh-semantics.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Index changes are computed in memory during refresh. Persisted cache staleness
  checks are intentionally deferred to Slice 4.
