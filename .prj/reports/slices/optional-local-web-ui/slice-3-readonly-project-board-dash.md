# Slice 3: Read-Only Project Board And Dash

Status: delivered

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-3-readonly-project-board-dash`

Expected PR Title: `feat(slice-3): read-only project board and dash`

PR: `#345`

## Summary

- Added read-only Web UI view models for Project, Board, and Dash.
- Added an application-layer provider that reuses existing board, readiness,
  and observability reads.
- Added full-page and htmx fragment routes for Project, Board, and Dash.
- Added templates, styles, tests, and smoke evidence for the read-only
  surfaces.

## Changes

- `internal/webui/views.go`
  - Adds Web UI read contracts for Project, Board, Dash, readiness, queue,
    events, loop, and execution summaries.
- `internal/app/web_views.go`
  - Maps existing `core.Board`, TUI readiness, and TUI observability
    projections into Web UI view models.
- `internal/webui/webui.go`
  - Adds `/project`, `/board`, `/dash`, and htmx fragment routes.
- `internal/webui/assets/templates/shell.html`
  - Renders Project, Board, and Dash content inside the shared shell.
- `internal/webui/assets/static/chc.css`
  - Adds styles for board columns, ticket cards, attention state, and dash
    panels.
- `internal/webui/webui_test.go`
  - Covers full-page and fragment rendering with a fake provider.
- `internal/app/cli_test.go`
  - Covers `/board` using a real project ticket fixture and provider.

## Validation

- `go test ./internal/webui ./internal/app -run 'WebServe|Handler|Project|Board|Dash'` (pass)
- `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- `go run ./cmd/chc web serve --dir . --port 7333` plus:
  - `curl -fsS http://127.0.0.1:7333/` (pass)
  - `curl -fsS http://127.0.0.1:7333/project` (pass)
  - `curl -fsS http://127.0.0.1:7333/board` (pass)
  - `curl -fsS http://127.0.0.1:7333/dash` (pass)
  - `curl -fsS -H 'HX-Request: true' http://127.0.0.1:7333/project/fragment` (pass)
  - `curl -fsS -H 'HX-Request: true' http://127.0.0.1:7333/board/fragment` (pass)
  - `curl -fsS -H 'HX-Request: true' http://127.0.0.1:7333/dash/fragment` (pass)
  - `curl -fsS http://127.0.0.1:7333/assets/chc.css` (pass)
- `make verify` (pass)

## Notes

- The Web UI remains read-only in this slice.
- Board data comes from `service.Board()` and `core.Board`.
- Dash readiness and observability reuse the existing TUI projections.
