# Slice 2: HTML HTMX And Helper Foundation

Status: ready for PR

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-2-html-htmx-helper-foundation`

Expected PR Title: `feat(slice-2): html htmx helper foundation`

## Summary

- Added the internal `webui` package boundary for server-rendered Web UI
  presentation.
- Added ChrysoCoord-owned htmx, render, and UI helper primitives without adding
  an external module dependency.
- Added embedded templates, CSS, htmx asset, root shell route, bootstrap
  fragment route, static asset serving, and shell smoke coverage.

## Changes

- `internal/webui/`
  - Adds the Web UI handler boundary, embedded assets, shell template, CSS, and
    route tests.
- `internal/webui/htmx/`
  - Adds htmx attribute builders, request/response header helpers, and tests.
- `internal/webui/render/`
  - Adds template rendering helpers, function-map composition, and tests.
- `internal/webui/ui/`
  - Adds initial button, chip, and nav primitives with escaping tests.
- `internal/app/web.go`
  - Connects `chc web serve` to the `internal/webui` handler and shell data.
- `internal/app/cli_test.go`
  - Adds shell and static asset route coverage.
- `ops/plans/2026-06-29-optional-local-web-ui-plan.md`
  - Marks Slice 2 as the active slice.
- `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`
  - Records Slice 1 delivery, Slice 2 task completion, commits, and validation.

## Validation

- `go test ./internal/webui/... ./internal/app -run 'WebServe|Handler|Attrs|Button|Chip|Nav|Renderer'` (pass)
- `go test ./internal/webui/... ./internal/app ./internal/core` (pass)
- `go run ./cmd/chc web serve --dir . --port 7332` plus:
  - `curl -fsS http://127.0.0.1:7332/` (pass)
  - `curl -fsS http://127.0.0.1:7332/assets/chc.css` (pass)
  - `curl -fsS http://127.0.0.1:7332/assets/htmx.min.js` (pass)
  - `curl -fsS http://127.0.0.1:7332/healthz` (pass)

## Notes

- The Web UI still exposes only shell/bootstrap surfaces; project, board, and
  dash read models start in Slice 3.
- The UI helper packages are presentation-only and are not imported by business
  packages.
