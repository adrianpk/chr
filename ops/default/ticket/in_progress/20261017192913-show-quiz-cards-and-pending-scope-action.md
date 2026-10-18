---
id: TKT-20261017192913
title: "Make the quiz catalog readable and scope-driven"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: quizzes, catalog, interaction
source: chat
reported_at: 2026-10-17T19:29:13Z
commits: []
branch: feat/slice-6-quiz-catalog
ready_at: 2026-10-18T08:34:50Z
started_at: 2026-10-18T08:34:50Z
---

# Make the quiz catalog readable and scope-driven

## Context and Evidence

Finding from [the Web UI review](20261017105432-review-draft-web-ui-usability.md).
The maintainer has now supplied two screenshots covering the upper catalog and
the lower action area of CHR's Quizzes page. The observations below describe
that actual interface. The expected behavior is stated independently of any
other application.

Capture only. This ticket remains open; no application change, quiz operation,
service update or data mutation is authorized. The quiz review is ongoing.

## Observed Problems

1. The page foregrounds an Artifacts count of 7 without explaining answered or
   pending quiz work.
2. The seven visible cards carry raw not_started and available labels. Their
   meaning and relationship are unclear in an ordinary quiz workflow. The
   screenshots establish the displayed labels, not whether the persisted
   answers agree with those labels.
3. Card contents are dominated by Scope and long operational file paths.
   Repeating the case name as Scope adds little context, while wrapped paths
   dictate irregular card heights and displace useful quiz information.
4. Cards provide no visible answer summary comparable to their substantial
   storage metadata. The page reads as an artifact inventory rather than a
   catalog for choosing a quiz to inspect or answer.
5. Start Quiz and Resume Quiz are separate full-width forms below the catalog.
   Their position separates them from the quiz cards and requires additional
   scrolling to discover the operations.
6. Each form asks the user to type a Case slug. The user must know an internal
   identifier and infer the intended project-wide versus case destination.
7. Start Quiz also exposes a Deep mode selector in that separate global form.
   Its role in the eventual question-and-answer flow has not yet been reviewed.

## Expected Catalog

1. Present a readable grid of quiz cards. The first card is the general Project
   Quiz, which is not attached to a use case. Keep that card visible even when
   it has no answers, with a clear human-readable empty state.
2. Follow it with one card for each use case whose quiz has been answered.
   Identify the case by its human-readable name.
3. Use a consistent card hierarchy: quiz identity, useful answer summary or
   empty state, and only progress information that helps the user choose work.
   Keep operational paths, repeated scope names and raw machine-state badges
   out of the ordinary catalog presentation.
4. Let the selected card provide the context for opening its quiz. The catalog
   should not require an internal file path or manually supplied case slug.
5. Use card dimensions and spacing that support reading and comparison across
   the available width. Avoid letting long storage paths determine card size.

## Visual Target

The outcome must reproduce the composition described here, not merely expose
correct commands in the existing artifact-card layout. These dimensions are
CSS pixels; browser zoom and device pixel ratio must not be confused with
measurements taken from a raster screenshot.

### Page Composition

1. Keep the existing shared application header and navigation. Below it, center
   the quiz content in a container with a maximum width of 860px and padding
   of 20px vertically and 16px horizontally. Leave calm outer margins on a
   large desktop rather than filling the viewport with narrow technical cards.
2. Place Quizzes at the left of a single title row and the conditional New Quiz
   action at its right edge. Align both to the same content container.
3. The title is approximately 1.3rem, weight 600, with no project-name prefix
   or repeated project subtitle. The project remains identified in the shared
   header.
4. Start the card grid about 12px below the heading region. Do not insert the
   current Artifacts counter or a technical status dashboard between them.

### Cards and Answer Excerpts

| Element | Target presentation |
| --- | --- |
| Desktop grid | Three equal fixed-width tracks within the bounded container, with a 10px gap. One or two cards keep that exact width and align from the left; they do not stretch to fill the row. |
| Card surface | Pure white, thin subdued border, 12px corner radius, approximately 10px inner padding and restrained hover/focus accent. |
| Card geometry | Explicit fixed width and height for every quiz card. Use a broad landscape composition with room reserved for title and preview. Content cannot change either dimension; minimum height alone is insufficient. |
| Card title | Top-left, approximately 0.92rem with semibold weight. Use Project Quiz or the case's recognizable name. |
| Content excerpt | A very pale green, almost-white inset panel subtly distinct from the pure-white card, below the title and anchored consistently toward the card bottom. Use a subtle border, 6px radius and approximately 7px vertical / 8px horizontal padding. |
| Excerpt height | One fixed area selected in the approximately 96–106px range, showing only opening quiz-content lines. Clip excess text without a preview scrollbar; short or long content never changes preview or card dimensions. |
| Empty project quiz | Show No answers yet in the same inset-panel position and geometry. |
| Interaction | The card is the selection target, with a clear pointer and keyboard focus treatment; keep a dense collection of secondary controls out of the card. |

The visible card anatomy is title plus useful quiz-content excerpt. Do not
recreate the current stack of two technical badges, a repeated scope row and a
multiline storage path with different colors. Retain CHR's identity and use its
compatible theme tokens; this is a Quizzes surface change, not a global
rebranding task.

At narrower widths, reduce to two and then one column before readability is
lost. Preserve card hierarchy and avoid horizontal overflow or a requirement
to change browser zoom. Do not impose an arbitrary six-column desktop layout.

### New Quiz Menu

1. Use a compact, normally sized button in the title row rather than a
   full-width operation bar. The entire action disappears when no pending
   destination exists.
2. On activation, anchor a small menu directly below the button, aligned to
   its right. The menu has a white/raised surface, a subtle border, an 8px
   radius, approximately 6px padding and a restrained shadow. A width in the
   190–260px range is the desktop baseline.
3. Present each pending scope as one legible, full-row named choice. Selecting
   it carries that scope into the quiz flow, without an intermediate slug form.

### Visual Verification

Use reproducible quiz data containing the general card, at least one answered
case and at least one pending case. Capture the catalog with the menu closed
and open, then the all-answered state with the action absent. Record the CSS
viewport, browser zoom and device pixel ratio; use 100% browser zoom.

Check the composition against the dimensions and anatomy above: bounded and
centered content, heading/action alignment, consistent landscape cards,
readable titles, aligned inset excerpts and a compact pending-scope menu.
Verify the same presentation with longer titles/content and at narrower
widths. HTTP success or the presence of functional controls alone does not
establish this ticket's visual outcome.

## Pending Quiz Entry Action

1. Show one New Quiz action at the top right only while at least one eligible
   scope has an unanswered quiz. Eligible scopes are the general project and
   the project's existing use cases.
2. Let the user select from those pending destinations using recognizable
   names. Do not require typing a case slug or use an unexplained blank field
   to represent project-wide work.
3. Existing cases without answered quizzes belong in this pending selection,
   rather than as unanswered placeholder cards in the main grid.
4. When every eligible scope has been answered, hide New Quiz entirely.
5. Replace the detached global Start Quiz and Resume Quiz forms with an entry
   flow bound to the selected quiz or pending scope. Preserve unfinished work
   and make continuation available in its own context instead of asking the
   user to reconstruct that context in another slug form.

## State Examples

| Project quiz | Use cases | Cards | New Quiz |
| --- | --- | --- | --- |
| Unanswered | None | Project card first, with its empty state | Visible; project is pending |
| Answered | One answered, one pending | Project first, then the answered case | Visible; offers the pending case |
| Unanswered | All answered | Project first, then answered cases | Visible; project is pending |
| Answered | All answered | Project first, then answered cases | Hidden |
| Answered | A new unanswered case is added | Existing cards remain | Visible; offers the new case |

## Remaining Refinement

Map answered, pending and partially answered sessions to CHR's actual native
quiz states and retained answer data. Do not infer those facts solely from a
not_started badge, migration metadata or the existence of an artifact file.
Determine the display and continuation of partially answered work without
resetting it or making it inaccessible.

The focused question-and-answer screen, its four actions, answer history and
Deep/Fast behavior are captured separately in
[Provide a focused quiz answering screen with modes and answer history](20261017193927-provide-focused-quiz-answering-screen.md).
This catalog ticket owns entry and selection; it does not independently redefine
question generation or declare the whole quiz review complete.

## Acceptance

1. The page communicates quiz identity, answers and pending work without
   foregrounding an artifact counter, file paths or raw machine-state badges.
2. Verify membership, ordering and New Quiz visibility against the state
   examples using real retained quiz and case data.
3. Quiz cards and pending selections supply the scope; no detached global
   start/resume form asks the user to type a case slug.
4. Existing answers and incomplete sessions remain available, with start and
   continuation tied to the selected destination.
5. Match the Visual Target composition and retain the specified visual
   verification screenshots alongside focused catalog/state and navigation
   evidence when implementation is explicitly authorized. Capturing this
   ticket does not execute quiz actions.

## Consolidated Design

Selected for coordinated Web UI delivery on 2026-10-18. The earlier
capture-only statements describe the initial capture boundary. Implementation
is now authorized for the linked approved set through Looper, its recorded
slice PRs and controller-owned integration into dev. The ticket and slice
report carry progress; the delivery tracker defines order only.

Behavior, shared-domain changes and implementation design are consolidated in
[Web UI and Workflow Consolidation](../../spec/stable/web-ui-workflow-consolidation.md).

## Slice 6 Execution

T6.1 delivers shared retained-answer membership after verified Slice 5 integration through PR #402. Its focused repository checks passed. T6.2 is active for the bounded catalog, pending menu and focused selection flow. The [Slice 6 report](../../report/slices/web-ui-workflow-consolidation/slice-6-quiz-catalog.md) owns implementation and validation evidence.
