---
id: TKT-20261017195458
title: "Present plan cards with read-only detail"
status: open
kind: task
severity: unclassified
priority: unclassified
scope: ui
tags: plans, catalog, reading
source: chat
reported_at: 2026-10-17T19:54:58Z
commits: []
---

# Present plan cards with read-only detail

## Context and Evidence

Part of [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
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
