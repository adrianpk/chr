---
id: TKT-20261017185837
title: "Use available Dumps grid width at default browser zoom"
status: in_progress
kind: bug
severity: medium
priority: normal
scope: ui
tags: web-ui, dumps, responsive-layout
source: chat
reported_at: 2026-10-17T18:58:37Z
commits: []
branch: fix/slice-4-dump-grid
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
