---
id: TKT-20261017195458
title: "Present plan cards with read-only detail"
status: solved
kind: task
severity: medium
priority: normal
scope: ui
tags: plans, catalog, reading
source: chat
reported_at: 2026-10-17T19:54:58Z
commits: [c65b1dac7c39e7a15030fb4f7b2e28a4366e5a15, f39a05003326d993687f0e9d0df37f1da6645a69]
branch: feat/slice-10-plan-catalog
pr: "#407"
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
reviewed_at: 2026-10-18T20:32:57Z
closed_at: 2026-10-18T20:32:57Z
resolution: fixed
---

# Present plan cards with read-only detail

## Context and Evidence

Part of [the Web UI review](../in_progress/20261017105432-review-draft-web-ui-usability.md).
The maintainer supplied the current Plans page HTML. It contains an Artifacts
counter, status/availability chips, repeated scope information, document paths
and document text inside catalog cards. Some previews include raw frontmatter
and lengthy operational content. This establishes the rendered content and
structure, rather than the exact browser appearance.

The requested experience is simple: cards like those specified for Specs,
with a click opening the document for reading. Plans must not be manually
editable from this view because their content governs agent work. Tracking
documents are internal and excluded from this catalog, as specified in the
separate
[internal tracking ticket](20261017200232-associate-ticket-tracking-with-each-plan-revision.md).

Capture only. No application change or service update is authorized here.

## Expected Behavior

1. Remove the Artifacts counter from Plans. Keep the catalog focused on its
   document cards instead of artifact statistics and technical metadata.
2. Use the same card anatomy as the
   [Specs catalog](20261017194552-present-spec-cards-and-user-selected-creation-state.md):
   a document title above a bounded inset preview of its actual opening body
   lines, with consistent dimensions and spacing. Do not turn the preview into
   a full document dump or fill it with frontmatter, paths and generic badges.
3. Clicking a plan card opens that exact document in a readable
   detail view with its complete content. The bounded catalog preview does
   not limit the content available in the detail view.
4. The document detail is read-only. Do not provide an editable body, Edit,
   Save or Append controls for plan text in this browsing flow.
5. This read-only boundary applies to manual Web document editing. Agents
   continue to update plans through their existing workflow. Do not change
   generation or execution semantics as part of this presentation change.

## Acceptance

1. The Plans catalog has no Artifacts counter. Its cards show the title and
   bounded opening body preview using the specified Specs card composition.
2. Short and long documents keep coherent card dimensions; raw frontmatter,
   full paths and entire document bodies do not dominate the preview.
3. Clicking a plan card opens the correct full document. Verify that the
   detail view has no manual document-editing controls.
4. Browsing cards and details leaves stored plan content unchanged.
   Capture catalog and read-only detail screenshots when
   implementation is authorized.

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

## Slice 10 Execution

T10.1 and T10.2 are complete on the recorded branch. Concrete plan reads retain
exact document selection, fixed opening-text cards and complete read-only detail.
The Web tracking form/action and direct tracker access are removed while
internal files and contextual generation remain available through their owners.

Focused core/HTTP tests, affected-package strict lint and owned Chromium
captures passed, including mutation-free browsing and rejected legacy forms.
The [Slice 10 report](../../report/slices/web-ui-workflow-consolidation/slice-10-plan-catalog.md)
owns implementation, acceptance captures and task validation evidence. [PR #407](https://forge.adrianpk.com/adrianpk/chr/pulls/407) and the controller
ledger own final clean-HEAD evidence after these PR-reference updates. PR #407 was verified merged into dev on 2026-10-18T20:32:57Z at
`ce1a19c0c10a50c27807b976fa353c81a0f6912e`; Slice 10 is delivered.
