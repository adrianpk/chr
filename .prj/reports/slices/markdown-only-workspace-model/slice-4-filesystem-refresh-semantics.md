# Slice 4: Filesystem Refresh Semantics

Status: delivered
Slicing: Markdown-Only Workspace Model
Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`
Tracker: `.prj/trackers/slicing-markdown-only-workspace-model-tracker.md`
Branch: `feat/slice-4-filesystem-refresh-semantics`
PR: #152

## Purpose

Make refresh a real runtime operation that reloads workspace files into the
ephemeral `WorkspaceGraph`.

This establishes the file-to-memory refresh path needed before Project and
Cases are fully rewired to graph-backed state.

## Delivered Behavior

- TUI startup builds an initial `WorkspaceGraph` from `.chc/`.
- The command palette `Refresh` command rebuilds the graph from filesystem
  state without restarting the TUI.
- Refresh updates the model's runtime graph.
- Refresh reports graph diagnostics and unresolved case references through the
  status bar.
- Refresh is blocked while Project or Case fields are actively edited.
- Refresh is blocked while Project or Case changes are dirty and unsaved.
- Runtime refresh language now describes file reload/graph rebuild, not
  database resync.

## Implementation Notes

The refresh flow is intentionally narrow:

- `chcapp` owns the concrete `WorkspaceGraphBuilder` call.
- `chctui` receives a `WorkspaceRefreshExecutor`.
- The TUI stores the current `WorkspaceGraph` but does not yet use it as the
  source for Project/Cases rendering.
- Project and Cases remain on their existing Markdown services until the
  planned TUI rewire slice.

## Contracts Added or Changed

- Added `WorkspaceRefreshExecutor`.
- Added `WorkspaceRefreshResult`.
- Added active-edit and dirty-state guardrails before refresh.
- Command palette refresh now performs a real graph rebuild.

## Files of Interest

- `internal/chcapp/app.go`
- `internal/chctui/workspace_refresh.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model.go`
- `internal/chctui/model_test.go`

## Validation

- `go test ./internal/chctui ./internal/chcapp ./internal/chc ./cmd/chc` passed
  on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Slice 5 must decide how much of the graph becomes visible in Project and Cases
  screens.
- The current refresh updates graph state only; it does not reload all TUI
  read models such as tickets or document previews.
