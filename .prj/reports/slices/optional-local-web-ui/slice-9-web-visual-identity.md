# Slice 9: Visual Identity And Layout Parity

Status: reviewing

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-9-web-visual-identity`

Expected PR Title: `feat(slice-9): web visual identity`

PR: `#351`

## Summary

- Added ChrysoCoord Web UI visual identity using a chrysocolla/Go-adjacent
  palette with amber and danger accents.
- Added a repo-native rounded bot/agent avatar asset.
- Refined layout density and responsive behavior across Project, Board, Dash,
  artifact catalogs, detail panes, and command forms.
- Added focus, wrapping, and sizing constraints for controls, cards, columns,
  panels, navigation, and code/path content.

## Changes

- `internal/webui/assets/static/chc-avatar.svg`
  - Adds the ChrysoCoord Web UI avatar asset.
- `internal/webui/assets/templates/shell.html`
  - Replaces the text-only brand mark with the embedded avatar image.
- `internal/webui/assets/static/chc.css`
  - Defines the visual token set for color, type, spacing, borders, radius,
    shadows, and status accents.
  - Removes the heavy outer workspace card so repeated items and operational
    panels carry the framing.
  - Adds responsive navigation, board, dash, command, and mobile layout rules.
  - Adds focus-visible states and overflow wrapping for buttons, links, cards,
    chips, paths, and code.
- `internal/webui/webui_test.go`
  - Covers the avatar asset route.

## Validation

- `go test ./internal/webui` (pass)
- Chromium screenshots generated and inspected from
  `go run ./cmd/chc web serve --dir . --port 7339 --require-token false`:
  - `/tmp/chc-slice9-screenshots/desktop-board.png` at `1440x1000` (pass)
  - `/tmp/chc-slice9-screenshots/desktop-dash.png` at `1440x1000` (pass)
  - `/tmp/chc-slice9-screenshots/mobile-project.png` at `390x900` (pass)
  - `/tmp/chc-slice9-screenshots/mobile-dumps.png` at `390x900` (pass)
- `make verify` (pass)

## Notes

- This slice is the visual pass after the core Web UI behavior is available.
- Preserve the existing operational model and route contracts while improving
  identity, layout density, responsive behavior, and accessibility.
- Screenshot files are not committed; the report records the capture paths and
  the validation result to avoid storing local-path UI captures in the repo.
