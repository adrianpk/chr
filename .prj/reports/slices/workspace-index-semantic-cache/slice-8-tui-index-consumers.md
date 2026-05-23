# Slice 8: TUI Index Consumers

Status: reviewing
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-8-tui-index-consumers`
PR: pending

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

- `internal/chctui/model.go`
- `internal/chctui/workspace_refresh.go`
- `internal/chctui/workspace_graph_projection.go`
- `internal/chctui/shell.go`
- `internal/chctui/observability.go`
- `internal/chctui/artifact_actions.go`
- `internal/chctui/model_test.go`
- `internal/chcapp/app.go`
- `internal/chcapp/service_workspace.go`
- `internal/chcapp/tui_adapters.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-8-tui-index-consumers.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chctui ./internal/chcapp` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- The create-ticket flow only surfaces use-case suggestions in this slice; it
  does not yet pass a selected use case into ticket creation.
