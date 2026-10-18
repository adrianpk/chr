# Slice 4: Dumps Grid Width

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `fix/slice-4-dump-grid`
PR: pending

## Purpose

Deliver the measured Dumps catalog sizing defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md).
Slice 3 is integrated. T4.1 establishes the width defect and fixed geometry;
T4.2 retains desktop zoom and narrow browser evidence.

## Delivered Behavior

The Dumps grid packs fixed 300 × 226 px cards into available width, with a
reserved title region and a 122 px clipped preview. Cards stay pure white;
previews use a very pale green inset theme token. Short, long, unbroken and
empty content retain identical geometry. Remaining track space is centered
without stretching cards. Narrow viewports reduce column count.

## Implementation Notes

T4.1 reproduced two 404 px tracks at 100% in a 1280 × 900 browser window.
At 90% the same catalog and window fit three. The corrected fixed tracks fit
three columns at 100% and four at 90%; 768 px fits two and 390/320 px fit one.
Dumps-scoped minimum-width/padding handling avoids horizontal page overflow
at 320 px while retaining card dimensions. Capture and Save owners are unchanged.

## Contracts Added or Changed

Theme tokens define dump card width/height, preview height and the shared pale
green document-preview surface. The owned browser fixture accepts an isolated
profile to set Chromium page zoom; CSS viewport and DPR verify the applied value.

## Files of Interest

- `internal/webui/assets/static/chr.css`: Dumps grid and card geometry.
- `internal/app/web_dump_grid_browser_test.go`: owned browser measurements and captures.

## Validation

- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app -run '^TestWebDumpGridBrowser$' -count=1 -v` — passed for 100%/90% desktop zoom and 768/390/320 px views, with identical card geometry and no horizontal overflow.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/app` — passed with zero issues after whitespace corrections in the fixture.
- `git diff --check` — passed.

Before the correction the same command reproduced the missing third column.
The fixed-card run also caught a 320 px page minimum-width overflow; the scoped
container correction passed without changing other pages. T4.2 screenshot and
final clean-HEAD controller evidence follow.

## Risks and Follow-ups

The controller owns integration into dev. No live project data is changed.
