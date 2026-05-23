# Slice 5: TUI Project And Cases Rewire

Status: reviewing
Slicing: Markdown-Only Workspace Model
Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`
Tracker: `.prj/trackers/slicing-markdown-only-workspace-model-tracker.md`
Branch: `feat/slice-5-tui-project-cases-rewire`
PR: pending

## Purpose

Point the TUI Project and Cases screens at Markdown-derived workspace graph
state instead of legacy list projections.

This completes the visible Project/Cases rewire that Slice 4 prepared by making
workspace graph refresh a runtime operation.

## Delivered Behavior

- TUI startup initializes Project metadata from the `WorkspaceGraph` when a
  project node is available.
- TUI startup initializes the Cases list from graph case nodes while preserving
  body fields already loaded by Markdown services.
- Project artifact counts render from project-wide graph nodes.
- Case artifact counts render from case-rooted graph nodes.
- Cases show unresolved case references from the graph without creating
  synthetic case records.
- Project and Cases `r` refresh now rebuilds the workspace graph from Markdown.
- Project and Case saves rebuild the workspace graph after the Markdown write.

## Implementation Notes

The TUI still uses the existing Markdown save services for inline editing. After
save, the model refreshes the graph and projects the refreshed graph back into
Project and Cases state.

Graph projection intentionally preserves body fields such as Project context and
Case context from the Markdown service load. The graph remains lightweight and
does not become the source for long-form document bodies.

Unresolved case references are rendered as a separate non-editable section in
the Cases list. They are diagnostic navigation metadata, not case definitions.

## Contracts Added or Changed

- `WorkspaceNode` carries `Description` for Project and Case graph projection.
- Project and Cases refresh keys now use `WorkspaceRefreshExecutor`.
- Project and Case save flows refresh the workspace graph after successful
  Markdown writes.
- `ProjectMetadataExecutor` and `CaseMetadataExecutor` now only expose save
  operations.

## Files of Interest

- `internal/chc/workspace_graph.go`
- `internal/chcapp/app.go`
- `internal/chctui/model.go`
- `internal/chctui/workspace_graph_projection.go`
- `internal/chctui/workspace_refresh.go`
- `internal/chctui/project_metadata_actions.go`
- `internal/chctui/case_metadata_actions.go`
- `internal/chctui/shell.go`
- `internal/chctui/model_test.go`

## Validation

- `go test ./internal/chctui ./internal/chcapp ./internal/chc ./cmd/chc` passed
  on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- This slice rewires Project and Cases visibility and refresh semantics, but it
  does not change LLM context assembly. Context assembly from graph remains
  planned for Slice 6.
