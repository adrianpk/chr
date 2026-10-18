---
id: TKT-20261017201015
title: "Implement the Kanban Board and shared ticket modal"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: board, kanban, tickets, modal
source: chat
reported_at: 2026-10-17T20:10:15Z
commits: [02c24fcfdc174a41755d89259810300e369d901d, f6907f9fa8b6644225a07ce2245c9aeb28299f97]
branch: feat/slice-1-web-board
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Implement the Kanban Board and Shared Ticket Modal

## Outcome and Authorization

Part of [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
Deliver a usable project Board: tickets appear as readable cards in lifecycle
columns, opening a card edits that ticket, and New opens the same modal to
create a ticket from scratch. This ticket consolidates the Board loading defect
and the separately captured ticket-modal findings into one delivery scope.

The maintainer authorized defining this ticket. Application implementation,
service changes and product-ticket mutations are not authorized by capture.
Keep this ticket open until it is selected for implementation.

## Verified Baseline

The running development Board currently returns:

```text
GET /board
HTTP 500
load board
```

Relevant existing implementation:

1. `internal/app/web_views.go` owns Board and ticket-detail projections;
   `web_workflow_ticket.go` also reads native workflow tickets. The current
   template renders those native tickets separately from the column grid.
2. `internal/core/board_query.go` defines six normal columns, stable ticket
   ordering and attention cases. `ticket_status.go` owns status aliases.
3. `internal/webui/assets/templates/shell.html` contains the Board and ticket
   modal templates; `internal/app/web_actions.go` exposes ticket metadata and
   comment writes. The current metadata operation covers only limited fields.
4. `internal/app/service_artifact_ticket.go` provides shared ticket creation,
   metadata update and comment operations. The new form must use and extend
   these owners rather than store browser-only ticket state.

The Board has implementation, but its load path fails in the current workspace.
The underlying cause has not been diagnosed in this capture.

Browser inspection of a working Kanban example covered the Board, a populated
edit modal, its complete internal scroll area, maximized presentation and the
empty creation form. No ticket or comment was submitted. The target below is
self-contained and requires no access to that example.

## Board Layout and Cards

1. Use Board as the page heading, without a project-name prefix. Align a single
   compact New action at the top right. Do not add an Artifacts counter,
   filesystem diagnostics or a separate native-ticket list above the Kanban.
2. Render the project's visible tickets once each in the common Board view,
   using the correct canonical owner for the selected workflow mode. Fix the
   HTTP 500 without disguising it as an empty result or requiring valid tickets
   to be recreated or moved back to obsolete storage.
3. Preserve CHR's six normal columns in their current order: Backlog, Ready For
   Dev, In Progress, Ready For Review, Ready For Merge and Done. Do not replace
   those stages with a different product's four-stage lifecycle. Resolve legacy
   aliases through the existing shared status mapping.
4. Each column has a readable heading and a small count of its tickets, then
   vertically stacked cards. Keep the column surfaces subtle, their headings
   aligned, and the cards visually separate from the column background.
5. Each card has its human-readable ticket key and title at the top, a compact
   priority badge, a short excerpt from its description and a sparse footer
   for meaningful tags and assignment when present. The full description stays
   in the modal. Avoid raw artifact paths, internal IDs and long protocol text.
6. Preserve discoverability of blocked tickets and other attention cases, with
   a clear reason using the existing attention model. Do not silently omit a
   ticket because its status lies outside the six normal lanes.
7. Keep the declared stable Board ordering. Board layout or visual dragging
   must not redefine slice_order, task_order or the agent's execution sequence.
   Bounded reads must provide an explicit way to reach additional tickets,
   rather than silently stopping at the current 128-ticket native preview cap.
8. Render an understandable empty state for a genuinely empty project and
   empty columns. A data-load failure must remain distinguishable from no data.

### Visual Composition

Use the available page width with about 16-24 CSS px of outer spacing and
12-16 px between columns. Columns should share available width while retaining
roughly 260-360 px of readable card width. Permit horizontal scrolling when
the viewport cannot fit the lanes; do not shrink six columns into unreadable
strips or require reduced browser zoom. Keep long columns scrollable without
losing access to their heading or cards.

Use the existing theme's raised card surface, subtle border, approximately
10-12 px corner radius and 12-14 px internal spacing. Titles wrap cleanly and
remain more prominent than excerpts and secondary metadata. Use a few readable
excerpt lines rather than dumping the complete document. Small tag pills and
assignment text form a compact bottom row. Omit empty secondary values.

Timing rows such as total or review time require meaningful existing canonical
data. Do not invent a time-accounting subsystem, show fabricated zero timers or
label ticket age as work duration merely to reproduce a sample card.

## Ticket Creation and Editing

1. Clicking a card opens the selected ticket's edit modal with its current
   fields populated. Keep the Board visible behind a dimmed backdrop.
2. New opens that same form in creation mode, with empty title and description,
   appropriate native metadata defaults and no inherited ticket identity,
   content, ownership or discussion from a previously opened card.
3. Share the field organization and interactions between creation and editing.
   Use Create for the new document and Save changes for an existing ticket.
   Creation uses the active namespace and ticket owner, allocating a new key
   through the canonical allocator; editing keeps the existing key and
   ownership. Selecting a use case does not silently insert a new ticket into
   a slice set or start its execution.
4. Provide one full Description editor. Derive the card excerpt from the stored
   description. Where the shared model requires a Synopsis, map it explicitly
   without adding two competing editors for the same text or an LLM request.
5. Use project-wide or existing use-case names in the scope selector. Do not
   ask users to type slugs or raw IDs. Respect fixed associations for tickets
   already owned by a slice set.
6. Closing without submission leaves stored data unchanged. After a successful
   create or save, refresh the affected card, count and column while preserving
   useful Board position. Surface validation errors beside the form and retain
   unsaved input; do not claim a save succeeded when the shared operation failed.

### Modal Layout

Normal mode is a centered, rounded panel about 900-1000 CSS px wide on a large
desktop, bounded by the viewport. Its header has the mode title, maximize or
restore control, and close control. Start with a full-width Title and a large
Description area, followed by clearly grouped fields. Use two columns for
short metadata where space permits and full width for the named use-case
selector and longer text fields. The main form scrolls internally; every
field and its submit action must remain reachable.

Maximized mode uses most of the viewport. Give the description editor the
larger left area and place grouped metadata in a narrower right area with
its own usable overflow. Keep all retained fields reachable, including
Ownership and Delivery. Maximize/restore changes presentation without losing
input or changing the selected ticket. On narrow viewports, use a readable
single-column form rather than clipping fields or actions.

### Form Fields and Ownership

| Group | Controls and behavior |
| --- | --- |
| Main content | Title text input and one large Description editor. |
| Workflow | Status selector mapped to allowed CHR transitions; system-owned workflow state is read-only when it adds distinct information. |
| Scope and classification | Named Use Case selector including Project-wide; Priority, Kind, Severity, Scope, Source and Resolution controls. Kind maps to the native type field; use CHR's actual enum values and defaults. |
| Ownership | Assignee and Owner metadata. These are descriptive responsibility fields, never editable execution claims or runner ownership tokens. |
| Delivery | Branch, PR reference and multiline commit references, using the shared evidence model and its write guards. System-owned verified evidence remains read-only where applicable. |
| Submission | Create for new tickets; Save changes for existing tickets. |

Extend the shared service/model for required fields that lack a supported
write path. Do not implement decorative inputs that fail to persist, a
parallel Web-only schema, or a second lifecycle vocabulary. Preserve existing
unknown metadata, provenance, comments and execution relationships when saving
description or metadata changes. Do not expose protocol frontmatter as the
document editor.

## Column Changes and Execution Safety

1. Permit allowed column changes from the Status selector and by dragging a
   card to a destination column. Both interactions invoke the same guarded
   operation and produce the same resulting state. Keep the selector available
   when pointer dragging is unsuitable.
2. A visual move cannot bypass active claims, task prerequisites, review,
   integration receipts or closure requirements. Read-only states and rejected
   transitions must have a comprehensible reason. Do not mark a grouped task
   delivered merely because a user placed its card in Done.
3. Use observed revisions for edits and moves. A concurrent agent update must
   not be silently overwritten. On rejection, restore the correct column and
   show the current state; never leave a successful-looking move unpersisted.
4. Refresh live progress without replacing a focused modal, discarding its
   unsaved text or jumping the Board's scroll position. Refresh and browsing
   are read operations, not triggers to start agent execution.

## Existing-Ticket Comments

Below the main edit form, provide a distinct Comments section with a count,
existing entries or an empty state, one multiline Markdown comment input and
an Add comment action. Comments show their existing author and timestamp in
readable form and retain their native ordering.

Comment submission is separate from saving ticket metadata and must not discard
unsaved main-form content. Reuse the canonical comment owner and preserve the
ticket's existing discussion when saving its description. New mode does not
show another ticket's discussion. Do not introduce notification delivery or
messaging integrations based on mention syntax in a placeholder.

## Delivery Boundary

This is one Board feature ticket, including the load repair, Kanban view,
shared create/edit modal, allowed moves and comments. Keep unrelated Dash
failures, global navigation changes, multi-board administration, new messaging
systems and tracker renaming outside it. The internal tracker remains absent
from the Web UI under its
[separate requirement](20261017200232-associate-ticket-tracking-with-each-plan-revision.md).

One maintained ticket replaces the preliminary loading and modal captures.
Implementation must remain consistent with the
[ticket-native workflow contract](../../spec/stable/ticket-native-workflow.md).

## Acceptance

1. Reproduce and repair the actual Board load failure. `/board` and its refresh
   show a populated native workflow and the supported legacy path correctly,
   without a duplicated native-ticket list. Verify a genuinely empty project.
2. Cards and per-column counts agree with canonical data. Attention cases and
   tickets beyond the first bounded page remain reachable. Verify the six
   native columns and legacy alias handling without changing execution order.
3. Click multiple cards to verify correct document selection. New uses the same
   form but starts fresh. Create, reload and edit a ticket; verify the retained
   field values and identity. Closing without saving does not alter storage.
4. Verify one allowed move through the selector and through drag-and-drop,
   one disallowed transition, and a stale edit after a concurrent ticket change.
   No path may fabricate delivery or overwrite newer state.
5. Verify comments persist through the shared owner and description saves
   preserve existing comments, metadata and evidence. A comment action does
   not silently save or discard unsaved ticket-form edits.
6. Capture screenshots at declared desktop and narrow viewports, 100% browser
   zoom, for the populated Board, normal edit modal, lower scroll area,
   maximized modal and new-ticket form. Verify readable cards, reachable fields
   and actions, and retained input when maximizing/restoring.
7. Use focused validation for the repaired read path, shared ticket mutations
   and Web interactions, following the repository's normal development gates.

## Implementation Progress

T1.1 is implemented in `02c24fcfdc174a41755d89259810300e369d901d`.
Canonical-only and empty stores previously failed because the legacy ticket
directory was required; mixed stores omitted canonical tickets. The shared
Board read now combines selected native authority and supported loose documents,
deduplicates native projections, retains attention records and pages beyond
128 tickets with complete column totals. Native inspection preserves canonical
bytes and rejects pending transitions without recovery.

T1.2 is implemented in `f6907f9fa8b6644225a07ce2245c9aeb28299f97`.
Shared operations now read and create canonical ticket documents and persist
full descriptions, classification, descriptive ownership, delivery references
and separate comments. Source comparison occurs at the existing persistence
owner; native execution, membership and receipt guards remain authoritative.
Repository classification aliases are explicit, and unsupported status-directory
transitions have a read-only reason. Regression fixes preserve nested metadata,
earlier comments in a final Comments section, native status/case identity during
metadata edits, and key allocation across namespaces and native envelopes.

The [Slice 1 report](../../report/slices/web-ui-workflow-consolidation/slice-1-web-board.md)
records the exact successful focused race, changed-code lint, signature and
whitespace commands. Final controller validation has not run. T1.3 shared
modal, interactions, read-only presentation and browser evidence remain;
the ticket is still in progress and no slice PR is ready.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).
