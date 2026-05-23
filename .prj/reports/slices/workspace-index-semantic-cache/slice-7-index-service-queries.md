# Slice 7: Service Queries And Suggestions

Status: reviewing
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-7-index-service-queries`
PR: pending

## Purpose

Expose adapter-facing workspace index queries through shared service behavior so
CLI, TUI, and future adapters can ask for cases, related documents, path
resolution, and search without parsing workspace files directly.

## Delivered Behavior

- Add structural workspace index queries for listing cases.
- Add ranked case suggestions over slug, name, and description.
- Add workspace path resolution through index path lookups.
- Add related-document resolution for case-associated workspace nodes.
- Add workspace index search over structural metadata.
- Add optional semantic digest matching when cached digests are available.
- Preserve semantic digest status in search results when evaluated states are
  available.
- Add service methods that build or load the cached index and expose the query
  results to adapters.

## Implementation Notes

The service methods use `WorkspaceGraphBuilder.BuildCachedIndex`, so adapters can
depend on the service instead of rebuilding graph/index state or parsing files.

Semantic search is opportunistic. Missing or empty digest caches still return
structural search results.

## Contracts Added or Changed

- Added `WorkspaceCaseSuggestion`.
- Added `WorkspaceRelatedDocumentsResult`.
- Added `WorkspaceIndexSearchResult`.
- Added `WorkspaceIndex.ListCases`.
- Added `WorkspaceIndex.SuggestCases`.
- Added `WorkspaceIndex.ResolvePath`.
- Added `WorkspaceIndex.RelatedDocuments`.
- Added `WorkspaceIndex.Search`.
- Added `WorkspaceIndex.SearchWithSemanticDigestStates`.
- Added service query methods for cases, path resolution, related documents, and
  workspace index search.

## Files of Interest

- `internal/chc/workspace_index_queries.go`
- `internal/chc/workspace_index_test.go`
- `internal/chcapp/service_workspace_queries.go`
- `internal/chcapp/service_test.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-7-index-service-queries.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chc ./internal/chcapp` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- TUI consumption of these queries remains deferred to Slice 8.
- Search ranking is intentionally simple and deterministic; it can be tuned after
  real adapter usage exposes better weighting signals.
