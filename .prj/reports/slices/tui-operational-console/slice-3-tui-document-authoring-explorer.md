# Slice 3: Document And Authoring Explorer

Status: reviewing
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-3-tui-document-authoring-explorer`
PR: pending

## Purpose

Expose project-local documents and authoring sessions as navigable read-only
TUI records.

This slice makes the TUI document-aware without introducing a competing project
state model or mutating authoring artifacts.

## Delivered Behavior

- Loads Markdown artifacts from project-local `.prj` and `.chc` document
  families.
- Renders the Documents view as a typed document index with status, title, path,
  updated time, and preview.
- Loads authoring sessions through the existing authoring session repository.
- Renders the Authoring view as a session list with selected session detail.
- Shows session identity, status, mode, target, topic, transcript path,
  artifacts path, inputs, and outputs.
- Adds record selection for document and authoring views.

## Implementation Notes

`internal/chcapp` keeps ownership of project resolution and TUI startup data
loading. It now loads documents through `internal/chctui.LoadDocumentViews` and
authoring sessions through `chc.NewAuthoringSessionRepository`.

Document discovery is intentionally read-only and file-backed. It summarizes
Markdown title/status/preview metadata and keeps source paths visible as
secondary detail.

Authoring session display uses the existing domain repository and layout helper
instead of parsing command output.

## Contracts Added or Changed

- `chctui.Options` now accepts document view records and authoring sessions.
- The TUI startup path reads project-local Markdown documents and authoring
  sessions before entering Bubble Tea.
- The Documents and Authoring views use up/down selection, matching ticket view
  navigation.

## Files of Interest

- `internal/chcapp/app.go`
- `internal/chctui/documents.go`
- `internal/chctui/authoring.go`
- `internal/chctui/model.go`
- `internal/chctui/model_test.go`

## Validation

```sh
go test ./internal/chctui ./internal/chcapp
make verify
```

All commands pass.

## Risks and Follow-ups

- Document discovery is based on known `.prj` and `.chc` artifact directories;
  later slices can add richer artifact classification if the domain exposes a
  shared document index.
- The first document preview is plain text. Rich Markdown viewport behavior is
  left for later TUI layout work.
