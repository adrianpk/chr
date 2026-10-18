---
id: TKT-20261017105432
title: "Review draft Web UI functionality, visual design and UX"
status: in_progress
kind: follow_up
severity: medium
priority: normal
scope: ui
tags: web-ui, ux-ui, usability
source: chat
reported_at: 2026-10-17T10:54:32Z
commits: [4c4815c51d5603d2c6e9576a7d7e9195100d7162]
branch: test/slice-12-web-ui-review
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Review draft Web UI functionality, visual design and UX

## Context

The maintainer considers the current Web UI an early draft of the intended
experience. It needs substantial work to become usable. Working
access and individual implemented controls do not establish that the interface
is complete or ready for comfortable everyday use.

## Deferred Review

Perform an exhaustive review of the actual interface and complete user journeys,
considering these dimensions together:

- Functionality: task coverage, navigation, data presentation and editing,
  action behavior, feedback, loading, empty states and error recovery.
- Visual design: layout, hierarchy, typography, spacing, color, density and
  consistency across screens and components.
- UX/UI: information architecture, discoverability, clarity, interaction flows,
  accessibility, responsive behavior and the effort required to complete tasks.

The review should identify concrete gaps and improvements needed for a usable,
coherent interface. Cosmetic changes alone are insufficient; functional and
interaction problems belong in the same assessment.

## Execution Boundary

The maintainer is reviewing the actual interface. Keep this review open and
capture future findings as development tickets. Record prospective work rather
than creating retrospective tickets for already implemented changes.

Capturing or refining findings does not authorize implementation. Application
code changes and service updates require an explicit implementation request.

## Review Notes

### Dumps

The initial observation accepted the existing presentation and retained the
slug column, whose practical value remains uncertain. The maintainer considered
the three-column table layout visually balanced.

A subsequent screenshot review identified a separate card-grid distribution
issue: at 100% browser zoom, two columns leave a large unused region; three
columns appear at 90%. Capture this in
[Use available Dumps grid width at default browser zoom](../solved/20261017185837-fix-dumps-grid-width-at-default-zoom.md).
This finding does not authorize implementation or changes to other aspects of
the view.

### Dump Ownership and Capture

The project-wide dump should be one evolving source of general context. New
general contributions extend it instead of creating separate general dumps.
Use-case captures retain their selected case scope. The Web UI should expose
one New action with Project-wide or an existing use case selected in the form.
Track the shared behavior, existing-data decisions and UI change in
[Unify dump capture around one evolving project-wide dump](../solved/20261017190508-unify-dump-capture-and-project-wide-ownership.md).

### Dump Reading, Editing and Appending

The current editor duplicates the stored content in a yellow read-only area and
an editable textarea, then exposes a separate append form below. The maintainer
cannot infer their purpose or relationship. Capture the interaction problem in
[Use one editor and Save for dump content](../solved/20261017191106-clarify-dump-reading-editing-and-appending.md).
The maintainer explicitly requires one editor and Save, removing both the
duplicate read-only panel and the separate Append textarea and button. Adding
text at the end is ordinary editing of the same Markdown document.

### Quiz Catalog and Actions

Two CHR screenshots show a catalog dominated by artifact counts, technical
badges, repeated scope names and long file paths, with separate global Start
Quiz and Resume Quiz forms below it requiring typed case slugs. The maintainer
wants a readable quiz-card catalog and a conditional New Quiz action offering
only pending project or existing-case scopes. Capture the observed problems
and the self-contained expected behavior in
[Make the quiz catalog readable and scope-driven](../solved/20261017192913-show-quiz-cards-and-pending-scope-action.md).
The catalog ticket includes a self-contained visual target: bounded centered
content, title and conditional action on one row, consistent landscape cards,
aligned inset excerpts and a compact pending-scope menu. Its acceptance requires
visual evidence as well as correct behavior.
The detailed answering screen and mode behavior are tracked separately below;
the quiz review remains open.

### Quiz Answering Screen and Modes

The requested screen has one current question, one multiline response area,
Skip/Answer/Cancel/Done together below it, and a direct Deep/Fast selector at
the top right. Answered question/response cards accumulate below, newest first.
Done closes the quiz even with questions remaining. Deep permits adaptive
nested follow-ups; Fast generates a flat set in one LLM generation request.
Capture the self-contained conceptual layout and shared-operation behavior in
[Provide a focused quiz answering screen with modes and answer history](../solved/20261017193927-provide-focused-quiz-answering-screen.md).
No palette match or application implementation is authorized by this capture.

### Spec Card Presentation

The maintainer is describing the desired individual spec-card anatomy before
continuing the comparison with CHR. Each card represents one spec and shows its
title above a bounded inset preview resembling the beginning of a small
document. The preview displays the spec's actual first lines, rather than a
generated summary or internal artifact metadata. Keep that preview visually
contained so longer documents do not expand the cards into full document views.
Clicking a card opens that spec for editing, with its stored content prefilled
and a Save action, whether the spec was written manually or generated by CHR.

The maintainer also requires a single New action in Specs and a creation
modal where the user selects draft or approved directly. Draft is a spec state,
not a mandatory separate document stage. Capture the individual-card target and
creation-state contract in
[Present individual spec cards and allow choosing the initial state](../solved/20261017194552-present-spec-cards-and-user-selected-creation-state.md).
The creation modal is one vertical form: Title, Use Case (Project-wide or an
existing case), one large Body editor and a compact Create action at the bottom
right. At most one small themed Draft switch or checkbox lets the user choose
the initial state, without a separate State selector or lifecycle panel.

Users can author specs manually, including for an existing use case without a
dump or quiz. CHR also generates specs from project context plus the relevant
dump and quiz when present. Both authoring paths produce ordinary specs that
the user can edit and save to correct errors, add omissions or expand the text.
Keep one editable body; do not duplicate it or require an append form. These
requirements do not add document categories or a new generation trigger.
CHR's actual Specs view is still pending visual comparison.

### Plans Catalog

The maintainer supplied the current Plans page HTML and rejected the Artifacts
counter and unnecessary catalog detail. Present documents as simple cards with
the same title and bounded opening-text preview as Specs. Clicking a card opens
the full plan for reading. The detail must be read-only: manual text changes
could affect the agent's work. Agent-owned plan updates remain part of their
existing workflow. Capture this presentation and
interaction contract in
[Present plan cards with read-only detail](../solved/20261017195458-present-plans-and-tracking-as-read-only-cards.md).
Tracking remains an internal artifact, excluded from the Web UI as specified
in the dedicated tracking ticket below.

### Contextual Plan Generation and Regeneration

The supplied HTML shows generic Generate Plan and Regenerate Plan forms below
the listing, each asking for a Case slug. Use named use-case context instead of
manual internal identifiers. Generation requires at least useful dump content
or quiz answers, plus enough information to plan the selected use case. A
complete dump, or a short dump for a sufficiently narrow use case, may avoid
the need for a quiz; mere file presence is not readiness.

Additional dump content or quiz answers can improve the spec and inform plan
regeneration. Keep the document operations distinct and make their scope clear.
Capture placement, prerequisites and regeneration context in
[Generate plans from use-case context with sufficient source material](../solved/20261017195741-contextualize-plan-generation-and-check-source-readiness.md).
Tracking creation and behavior are excluded from that ticket. Manual spec
creation still does not require a dump or quiz.

### Ticket-Based Tracking per Plan Revision

The maintainer requires an internal tracker associated with every plan
generation or regeneration. It gives the agent the deterministic ticket
sequence, ordered by the native slice and ticket ordering fields. Processing
and completion are recorded in tickets, without a second processed-task
checklist in Markdown. Existing workflow records continue to own ordering and
state; the tracker supplies the agent's readable map.

Its only responsibility is to define that sequence. Advancing or completing
tickets does not require synchronizing tracker processing fields, refreshing
the document or repairing its unchanged order. Keep the current tracker name
and compatible structure during this transition. Any legacy progress columns
are unused; a future rename is separate work.

This Markdown artifact remains in the project for manual inspection. It has
no Web catalog card, detail view, editor or Create Tracking form. Automatic
creation belongs to the plan workflow. The current optional tracker contract
is the implementation baseline, rather than the new availability requirement.
Capture the requirement and its approval/materialization integration in
[Require an internal tracker for deterministic ticket order](../solved/20261017200232-associate-ticket-tracking-with-each-plan-revision.md).

### Kanban Board and Ticket Management

The maintainer reports that Board is unusable and shows a short text response.
A request to the current development service confirms HTTP 500 with the body
load board. Repository Board handlers and templates exist, so capture a loading
failure rather than claiming the feature has no implementation.

The maintainer then requested one complete Board ticket based on the supplied
Kanban screenshot and direct browser inspection. The combined target includes
readable cards in CHR's six native lifecycle columns, per-column counts, one
New action, populated card-editing and empty creation modes of the same modal,
complete field groups, internal scrolling, maximize/restore, allowed column
moves and existing-ticket comments. Preserve native state, membership,
execution-order and delivery guards rather than copying another lifecycle.

The loading and modal captures have been consolidated into one maintained
ticket with a self-contained visual and functional target and acceptance:
[Implement the Kanban Board and shared ticket modal](../solved/20261017201015-restore-board-page-loading.md).
The redundant preliminary modal file has been removed. This remains planning
and ticket capture, without implementation or service changes.

### Live Agent Work Dash

The maintainer wants Dash to show what the agent harness is doing while it
processes work. The current Web route returns HTTP 500 with load dash. TUI
source and the existing gallery establish queue, readiness, current execution,
phase/activity and recent/final results as the operational view.

The Web projection is partial: it does not render the latest loop/execution
results it loads or expose the loaded native workflow page. Capture a working
live view using canonical tickets, workflow state and events in
[Implement Dash as a live view of agent work](../solved/20261017202340-implement-live-agent-work-dash.md).
The tracker remains an internal order document, with no processing-state
synchronization. Monitoring must not block behind an active-action overlay,
require a TUI process, start work as a side effect or mix concurrent runners'
ticket and event context.

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

## Slice 12 Execution

T12.1 is complete: the combined persisted-workspace journey passes focused race
and strict lint checks. It covers capture, quiz, manual/generated spec edits,
contextual planning, approved order binding, Board metadata, Dash observation
and mutation-free reopening. T12.2 addresses consolidated presentation/interaction
acceptance, then T12.3 records the final matrix and candidate evidence.

T12.2 is complete: the combined browser verifies desktop/narrow geometry,
focus/labels, dirty forms and independent comments. The plan-bound title guard
was reproduced and corrected through the existing ticket/order owners. Focused
core/app/webui race and strict lint checks pass. T12.3 remains for the complete
source acceptance matrix, user-facing documentation and final candidate evidence.
