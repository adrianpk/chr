# Slice 4: Dumps Grid Width

Status: delivered
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `fix/slice-4-dump-grid`
PR: `#401`

## Purpose

Deliver the measured Dumps catalog sizing defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md).
T4.1 and T4.2 were integrated through [PR #401](https://forge.adrianpk.com/adrianpk/chr/pulls/401)
at `54fb4f6eeba8407c05893a77c4b54c74668018e6` on 2026-10-18.

## Delivered Behavior

Dumps packs fixed 300 × 226 px cards into the available width. The measured
1280 px desktop window fits three columns at 100% zoom and four at 90%.
768 px fits two columns; 390/320 px fit one. Remaining track space is centered
without stretching cards. Dumps-scoped minimum-width and padding handling
prevent horizontal page overflow at 320 px.

Short, long, unbroken and empty content keep the same card and preview size.
Titles occupy a bounded two-line region. A fixed inset preview shows five
opening lines with clipped excess text, on a very pale green surface inside
the pure-white card. Full content remains available in the existing editor.
Capture, Save, selection and modal actions retain their existing owners.

## Implementation Notes

- T4.1: `411925a719bd0bcd61cda4de9f8886f64f5fd0b5` — measured width reproduction, fixed track/card tokens, bounded title/preview and narrow container correction.
- T4.2: `1043709e4846510dba05f18c6f954b89b801b414` — owned zoom/narrow captures, content/count geometry assertions and shared editor/capture/Board browser regressions.

The original auto-fit maximum reserved 404 px tracks before choosing a column
count. In a 1280 × 900 window it displayed two columns at 100%, with a 1226 px
available grid; at 90% it fit three. Fixed 300 px tracks account for the existing
14 px gap and expose the usable third column at default zoom. Column count
remains a consequence of available width.

The preview text has its own bounded inner area so padding does not expose a
partial sixth line. A single remaining card keeps its size. The browser fixture
waits for the observed viewport after resizing before measuring layout.

## Contracts Added or Changed

Theme tokens define dump card width/height, preview height and the pale green
document-preview surface. The preview has no internal scrollbar. No source
material, capture binding, body writer or lifecycle contract changes.

The owned browser fixture accepts an isolated profile for real Chromium page
zoom. The desktop pair keeps its 1280 × 900 outer window and the same six-document
catalog. Observed CSS viewport and DPR verify zoom. Narrow views use declared
CSS viewport emulation at DPR 1. Shared screenshot capture waits for paint.

## Files of Interest

- `internal/webui/assets/static/chr.css` and `assets/templates/shell.html`: grid sizing, fixed card anatomy and bounded preview text.
- `internal/app/web_dump_grid_browser_test.go`: real zoom, responsive geometry, long/empty/unbroken content and card-count checks.
- `internal/app/web_workflow_browser_pipe_test.go` and `web_board_browser_helpers_test.go`: isolated browser profile and shared screenshot fixture.
- `files/slice-4/`: captured views with adjacent JSON measurements.
- `docs/how-to/user-guide/index.md` and `CHANGELOG.md`: user-facing catalog behavior.

## Validation

Successful task checks:

- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app -run '^TestWebDumpGridBrowser$' -count=1 -v` — T4.1 zoom and 768/390/320 px geometry checks passed.
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_DUMP_GRID_SCREENSHOTS="$PWD/ops/default/report/slices/web-ui-workflow-consolidation/files/slice-4" go test ./internal/app -run '^(TestWebDumpGridBrowser|TestWebDumpEditorBrowser|TestWebDumpCaptureBrowser|TestWebBoardBrowser)$' -count=1` — five captures and affected shared browser journeys passed.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/app` — passed with zero issues.
- `git diff --check` — passed.

The first geometry regression reproduced the missing third column. Narrow
checks caught the existing page minimum-width overflow. Further capture checks
identified fractional border rounding at 90% and a resize acknowledgement before
viewport reflow. Assertions retain equal geometry across all content, permit less
than one physical pixel of nominal preview-width rounding, and wait for the
actual resized viewport. Preview capture also verified clean five-line clipping.

All views use the same six documents, including short, long, unbroken and empty
bodies and a long unbroken title. Each card is 300 × 226 CSS px; its preview is
270 × 122 CSS px at DPR 1. Fractional 90% measurements remain in the JSON files.
The single-card variant preserves those dimensions. All page-overflow assertions
pass. Each fixture owns and stops its Chromium/server processes and uses an
isolated store with deterministic curation.

| View | CSS viewport | DPR | Columns | Capture | Measurements |
| --- | --- | --- | --- | --- | --- |
| Desktop 100% | 1280 × 813 | 1 | 3 | [100%](files/slice-4/desktop-100.png) | [JSON](files/slice-4/desktop-100.json) |
| Same window, 90% | 1422 × 903 | 0.9 | 4 | [90%](files/slice-4/desktop-90.png) | [JSON](files/slice-4/desktop-90.json) |
| Medium 100% | 768 × 844 | 1 | 2 | [768 px](files/slice-4/width-768.png) | [JSON](files/slice-4/width-768.json) |
| Narrow 100% | 390 × 844 | 1 | 1 | [390 px](files/slice-4/width-390.png) | [JSON](files/slice-4/width-390.json) |
| Minimum 100% | 320 × 844 | 1 | 1 | [320 px](files/slice-4/width-320.png) | [JSON](files/slice-4/width-320.json) |

Final clean-HEAD validation passed through Looper for
`54fb4f6eeba8407c05893a77c4b54c74668018e6` in canonical order:
`make verify`, the focused grid/editor/capture/Board browser command and
`git diff --check`. The canonical PR lists the exact successful final commands;
Looper retains the tested HEAD and ordered evidence. Development checks do not
establish scheduled exhaustive/nightly certification.

## Risks and Follow-ups

The controller owns integration into dev. No live project data was changed.
