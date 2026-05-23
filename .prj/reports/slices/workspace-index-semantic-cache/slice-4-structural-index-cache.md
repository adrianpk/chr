# Slice 4: Discardable Structural Cache

Status: delivered
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-4-structural-index-cache`
PR: #167

## Purpose

Persist the structural workspace index as a local discardable cache so TUI and
service refresh flows can reuse a validated index across process restarts while
keeping the filesystem as the source of truth.

## Delivered Behavior

- Write `structural-index.json` under `.chc/tmp/cache/workspace-index/`.
- Store a schema version, write timestamp, and the full structural index.
- Load the cached index when schema and source facts remain valid.
- Treat missing cache files as cache misses.
- Treat source hash, modified time, size, added node, and removed node
  differences as stale cache inputs.
- Rewrite the cache after misses and stale reads.
- Expose cache status and path through workspace refresh results.
- Use the cached-index builder during TUI startup and service refresh.

## Implementation Notes

The cache remains discardable. It is derived from the filesystem-backed
workspace graph and is rewritten whenever it is absent or stale.

This slice intentionally does not introduce semantic digests or background LLM
indexing. Those responsibilities belong to later semantic cache slices.

## Contracts Added or Changed

- Added `WorkspaceIndexCacheStatus`.
- Added `WorkspaceIndexCacheState`.
- Added `CachedWorkspaceIndexResult`.
- Added `WorkspaceGraphBuilder.BuildCachedIndex`.
- Added `WorkspaceIndexCachePath`.
- Extended workspace refresh results with index cache state.

## Files of Interest

- `internal/chc/workspace_index_cache.go`
- `internal/chc/workspace_index_cache_test.go`
- `internal/chcapp/app.go`
- `internal/chcapp/service_workspace.go`
- `internal/chcapp/tui_adapters.go`
- `internal/chctui/workspace_refresh.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-4-structural-index-cache.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chc ./internal/chcapp ./internal/chctui` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- The structural cache validates against the freshly built graph in this slice,
  so it prioritizes correctness and restart persistence over avoiding the graph
  scan. More aggressive source-manifest shortcuts can be evaluated separately if
  startup cost becomes measurable.
