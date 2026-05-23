# Slice 3: Workspace Graph Builder

Status: delivered
Slicing: Markdown-Only Workspace Model
Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`
Tracker: `.prj/trackers/slicing-markdown-only-workspace-model-tracker.md`
Branch: `feat/slice-3-workspace-graph-builder`
PR: #151

## Purpose

Introduce the first runtime `WorkspaceGraph` builder for Markdown-only
workspace navigation.

The graph is an ephemeral in-memory view reconstructed from `.chc/` files. It
does not persist a second source of truth and does not read full document bodies
as durable application state.

## Delivered Behavior

- Builds a graph from Markdown files under `.chc/`.
- Loads `.chc/project.md` as the project node when present.
- Loads `.chc/cases/*.md` as canonical case nodes.
- Classifies `.chc/dumps`, `.chc/quizzes`, `.chc/specs`, `.chc/plans`,
  `.chc/tickets`, and `.chc/reports` Markdown files.
- Separates project-wide documents from case-scoped documents.
- Groups case-scoped documents by canonical case slug.
- Records unresolved case references without implicitly creating cases.
- Records parse/frontmatter diagnostics without aborting the whole graph build.
- Keeps source path, relative path, metadata, source hash, modified time, and
  file size as lightweight graph metadata.

## Implementation Notes

The builder lives in `internal/chc` as a pure service:

- `WorkspaceGraphBuilder`
- `WorkspaceGraph`
- `WorkspaceNode`
- `WorkspaceGraphDiagnostic`
- `WorkspaceCaseReference`

Project and Case parsing reuse the existing Markdown metadata parsers. Ticket
nodes reuse ticket frontmatter parsing. Generic document nodes use lightweight
frontmatter fields and Markdown headings.

This slice does not wire the graph into TUI screens yet. That belongs to the
planned TUI Project/Cases rewire slice.

## Contracts Added or Changed

- Added runtime-only `WorkspaceGraph` model.
- Added case-rooted view lookup through `CaseView`.
- Added unresolved case reference reporting.
- Added graph diagnostics for invalid Markdown/frontmatter.
- Kept Project and Case Markdown services as the durable persistence boundary.

## Files of Interest

- `internal/chc/workspace_graph.go`
- `internal/chc/workspace_graph_test.go`

## Validation

- `go test ./internal/chc ./internal/chcapp ./internal/chctui ./cmd/chc` passed
  on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Slice 4 must define refresh semantics around rebuilding this graph while the
  app is running.
- Slice 5 must wire Project and Cases TUI screens to graph-backed state.
