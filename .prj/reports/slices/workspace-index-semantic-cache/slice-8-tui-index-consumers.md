# Slice 8: TUI Index Consumers

Status: delivered
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-8-tui-index-consumers`
PR: #171

Verified merge: `dec1921c9b2ee93459a0e7a9f10e3eabbe2d9ab4`, an ancestor of dev
as checked on 2026-10-20.

## Purpose

Make the TUI consume workspace index query results for case-oriented surfaces
and expose index health without blocking startup on semantic indexing.

## Delivered Behavior

- Initialize TUI case metadata from the structural workspace index when
  available.
- Refresh TUI case metadata from the refreshed workspace index.
- Filter cases through workspace index suggestions when the index is available.
- Count case artifacts from index case document lookups instead of graph-only
  projections.
- Carry semantic digest states through TUI startup and workspace refresh.
- Show workspace index health and semantic pending/stale digest counts in the
  Dash runtime panel.
- Surface use-case suggestions in the create-ticket confirmation view.

## Implementation Notes

Semantic digest state loading reads the local discardable cache and evaluates it
without generating digests. Startup and refresh remain non-blocking with respect
to semantic generation.

The TUI still keeps `WorkspaceGraph` compatibility for unresolved reference
rendering and fallback projection paths.

## Contracts Added or Changed

- Added `SemanticDigestStates` to TUI options and refresh results.
- Added TUI state for semantic digest states.
- Added index-backed case projection helpers.
- Added index-backed case artifact counting.
- Added workspace index health rendering.

## Files of Interest

- `internal/chrtui/model.go`
- `internal/chrtui/workspace_refresh.go`
- `internal/chrtui/workspace_graph_projection.go`
- `internal/chrtui/shell.go`
- `internal/chrtui/observability.go`
- `internal/chrtui/artifact_actions.go`
- `internal/chrtui/model_test.go`
- `internal/chrapp/app.go`
- `internal/chrapp/service_workspace.go`
- `internal/chrapp/tui_adapters.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-8-tui-index-consumers.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chrtui ./internal/chrapp` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- The create-ticket flow only surfaces use-case suggestions in this slice; it
  does not yet pass a selected use case into ticket creation.
