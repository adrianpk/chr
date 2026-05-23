# Slice 2: Structural Workspace Index Model

Status: delivered
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-2-structural-workspace-index`
PR: #165

## Purpose

Introduce the structural workspace index model as a graph/read-model derived
from the existing Markdown-backed `WorkspaceGraph`.

## Delivered Behavior

- Added `WorkspaceIndex` as a structural read model.
- Added `WorkspaceEdge` and edge kinds for containment and case references.
- Added lookup maps for nodes by id, path, kind, case slug, case slug lookup,
  and ticket key lookup.
- Added deterministic node IDs derived from kind and relative path when
  available.
- Added `NewWorkspaceIndexFromGraph` to build the index from the existing graph.
- Added `WorkspaceGraphBuilder.BuildIndex` while preserving existing
  `WorkspaceGraph` APIs.
- Added focused tests for lookups, edges, project-wide documents, unresolved
  case references, and builder integration.

## Implementation Notes

This slice keeps the index as a compatibility layer over `WorkspaceGraph`.
Consumers can continue using the graph while later slices move service queries
and TUI consumers toward the structural index.

The index does not load full document bodies. It uses the metadata already
present in graph nodes: paths, kind, slug, title, status, case association,
source hash, modified time, and size.

## Contracts Added or Changed

- Added `WorkspaceIndex`.
- Added `WorkspaceEdge`.
- Added `WorkspaceEdgeKind`.
- Added `WorkspaceEdgeContains`.
- Added `WorkspaceEdgeReferencesCase`.
- Added `WorkspaceNodeID`.
- Added `NewWorkspaceIndexFromGraph`.
- Added `WorkspaceGraphBuilder.BuildIndex`.

## Files of Interest

- `internal/chc/workspace_index.go`
- `internal/chc/workspace_index_test.go`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-2-structural-workspace-index.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chc` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- `WorkspaceGraph` remains the current consumer-facing model. Later slices will
  migrate service queries and TUI consumers to the index where it adds value.
