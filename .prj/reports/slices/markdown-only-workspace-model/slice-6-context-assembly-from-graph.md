# Slice 6: Context Assembly From Graph

Status: reviewing
Slicing: Markdown-Only Workspace Model
Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`
Tracker: `.prj/trackers/slicing-markdown-only-workspace-model-tracker.md`
Branch: `feat/slice-6-context-assembly-from-graph`
PR: pending

## Purpose

Add a reusable context assembly boundary that selects Markdown workspace
material from the ephemeral `WorkspaceGraph`.

This lets LLM-facing workflows start from case-rooted or project-wide graph
views without relying on duplicated Project/Case persistence.

## Delivered Behavior

- Added case-rooted workspace context requests.
- Added project-wide workspace context requests.
- Case-rooted assembly includes Project metadata, the selected Case document,
  and case-scoped dumps, quizzes, specs, and plans.
- Project-wide assembly includes Project metadata and project-wide dumps,
  quizzes, specs, and plans.
- Workspace document bodies are read during assembly, not stored in the graph.
- Missing files are reported through source inventory entries.
- Missing selected cases are reported as missing required sources.
- Workspace graph diagnostics and unresolved case references are rendered into
  assembled context.
- Context-aware quiz generation now includes a `workspace_graph_context`
  section assembled from the project-wide graph.

## Implementation Notes

The assembly API lives in `internal/chc` and uses existing context primitives:
`ContextDocument` and `ContextSourceInventoryEntry`.

The graph remains lightweight. `WorkspaceNode` metadata decides what to include,
and file bodies are loaded only when `AssembleWorkspaceContext` is called.

Tickets and reports are intentionally excluded from the graph-selected document
set for this slice. Ticket execution already has a ticket-specific context path,
and reports are not part of the planned case/project context collection.

## Contracts Added or Changed

- Added `WorkspaceContextRequest`.
- Added `WorkspaceContextAssembly`.
- Added `NewCaseRootedWorkspaceContextRequest`.
- Added `NewProjectWideWorkspaceContextRequest`.
- Added `AssembleWorkspaceContext`.
- Added `RenderWorkspaceContextAssembly`.
- Added `workspace_graph_context` to authoring quiz generation requests.

## Files of Interest

- `internal/chc/workspace_context.go`
- `internal/chc/workspace_context_test.go`
- `internal/chc/authoring_contextual_quiz.go`
- `internal/chc/authoring_contextual_quiz_test.go`

## Validation

- `go test ./internal/chc ./internal/chcapp ./internal/chctui ./cmd/chc` passed
  on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Case-rooted assembly is available as an internal API but is not yet connected
  to a TUI-selected case action.
- Context budget controls remain outside this slice.
