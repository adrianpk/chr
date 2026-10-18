---
id: TKT-20261017185837
title: "Use available Dumps grid width at default browser zoom"
status: solved
kind: bug
severity: medium
priority: normal
scope: ui
tags: web-ui, dumps, responsive-layout
source: chat
reported_at: 2026-10-17T18:58:37Z
commits: [411925a719bd0bcd61cda4de9f8886f64f5fd0b5, 1043709e4846510dba05f18c6f954b89b801b414]
branch: fix/slice-4-dump-grid
pr: "#401"
resolution: fixed
reviewed_at: 2026-10-18T16:53:55Z
closed_at: 2026-10-18T16:53:55Z
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Use available Dumps grid width at default browser zoom

## Context

Finding from [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
The maintainer supplied a screenshot and compared browser zoom levels while
actively reviewing the Dumps catalog. This ticket captures the finding only;
implementation requires a separate explicit request.

## Observed Behavior

At 100% browser zoom, the desktop Dumps catalog displays two card columns and
leaves a large unused region on the right, visually equivalent to a third
column. Three columns appear only after reducing browser zoom to 90%.
The maintainer considers this an inefficient distribution of the available
screen width.

## Reproduction

1. Open a populated Dumps catalog at the desktop window size used for review.
2. Set browser zoom to 100% and observe the two columns and unused right side.
3. Keep the window size unchanged and reduce browser zoom to 90%.
4. Observe that the grid now displays three columns.

## Expected Outcome

Use the available catalog width coherently at the default 100% zoom. At the
reported desktop window size, allow a balanced three-column layout without
requiring the user to reduce browser zoom. Keep column selection responsive
to the actual available width rather than requiring three columns at every
window size.

## Acceptance

1. Compare the same populated catalog and window size at 100% and 90% zoom.
2. At 100%, the grid uses the available desktop width without leaving the
   reported unused third-column-sized region.
3. Preserve readable card contents and spacing as the window width changes.
4. Keep this change limited to catalog layout; content and card actions remain
   outside this finding.

## Investigation Boundary

The CSS or container cause has not been established. Inspect sizing and
breakpoints when implementation is explicitly authorized. The observation
alone does not establish whether the cause is card width, a grid breakpoint,
a container constraint or another layout rule.

## Implementation Progress

T4.1 is implemented in `411925a719bd0bcd61cda4de9f8886f64f5fd0b5`.
A 1280 × 900 Chromium window reproduced two 404 px columns at 100% and three
at 90%. Fixed 300 × 226 px cards now fit three at default zoom, with a fixed
122 px inset preview and bounded titles. Available width determines column
count; centered tracks preserve card dimensions. Dumps-scoped narrow container
rules prevent page overflow at 320 px.

T4.2 is implemented in `1043709e4846510dba05f18c6f954b89b801b414`.
Five retained captures record 100%/90% desktop zoom and 768/390/320 px CSS views.
JSON measurements retain viewport, DPR, geometry and overflow state. Short,
long, unbroken, empty and single-card variants keep fixed dimensions; preview
text clips to five complete opening lines. Existing editor/capture/Board browser
journeys pass with the new layout. User Guide and Unreleased are updated.

Successful task checks:

- `CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app -run '^TestWebDumpGridBrowser$' -count=1 -v`
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_DUMP_GRID_SCREENSHOTS="$PWD/ops/default/report/slices/web-ui-workflow-consolidation/files/slice-4" go test ./internal/app -run '^(TestWebDumpGridBrowser|TestWebDumpEditorBrowser|TestWebDumpCaptureBrowser|TestWebBoardBrowser)$' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/app`
- `git diff --check`

The [Slice 4 report](../../report/slices/web-ui-workflow-consolidation/slice-4-dump-grid.md)
links the captures and measurements and records the failures and corrections.
All tasks are implemented in canonical [PR #401](https://forge.adrianpk.com/adrianpk/chr/pulls/401).
Final clean-HEAD validation is recorded through Looper with the ordered successful
commands in the PR. The controller merged PR #401 into dev at
`54fb4f6eeba8407c05893a77c4b54c74668018e6` on 2026-10-18 after all
three ordered exact-HEAD checks passed. The ticket is solved.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).

## Fixed Card Geometry

Maintainer clarification, 2026-10-18: catalog cards always use fixed width and
height per type. Their preview area is also fixed and shows the document's
first lines with excess text clipped. Place this preview inside the card below
its title, with a very pale green, almost-white background subtly distinct
from the pure-white card, inset padding and a subtle boundary so it resembles
the beginning of a small document.
Content and long titles cannot enlarge
the card or preview, and previews have no internal scrollbar. Responsive grids
change the number of columns rather than stretch the cards. Full content stays
available through the existing detail/editor. Verify equal card and preview
dimensions with short, long and unbroken text in the owning slice and review.
