# Web UI and Workflow Consolidation

Status: Approved

Date: 2026-10-18

Approved: maintainer, 2026-10-18; the explicit request to proceed with Looper accepts this set and its implementation design.

Source baseline: `a0383b84722c63bda9734cdfcf7633b044c967af`.

Model: [Web UI and Workflow Model](web-ui-workflow-consolidation-model.md).

Plan: [Web UI and Workflow Delivery](../../plan/web-ui-workflow-consolidation.md).

Tracker: [Web UI and Workflow Delivery Order](../../tracker/web-ui-workflow-consolidation.md).

Implementation authorization: granted on 2026-10-18 for this exact delivery
set through Looper, including its canonical slice PRs and controller-owned
Rebase + Fast-forward integration into dev. Validate domain/migration behavior
on isolated fixtures; applying reconciliation to live user data or changing
production services is outside this execution authority.

## 1. Purpose and Scope

Make the existing Web UI usable for collecting context, answering quizzes,
authoring specs, reading plans, managing tickets and observing agent work.
Use one coherent server-rendered interface and the existing shared application
operations. Include the domain and persistence changes necessary for those
journeys, rather than hiding incomplete behavior behind working controls.

This specification unifies the following twelve development tickets. They
retain their identities and individual acceptance requirements. All are selected
as in progress, with `severity: medium` and `priority: normal`; normal is the
repository's intermediate priority value. Selection does not claim delivered
application behavior.

| Source ticket | Responsibility | Sections |
| --- | --- | --- |
| [TKT-20261017105432](../../ticket/in_progress/20261017105432-review-draft-web-ui-usability.md) | Complete functional, visual and UX review | 2, 3, 12 |
| [TKT-20261017185837](../../ticket/in_progress/20261017185837-fix-dumps-grid-width-at-default-zoom.md) | Dumps width at normal zoom | 3, 4 |
| [TKT-20261017190508](../../ticket/in_progress/20261017190508-unify-dump-capture-and-project-wide-ownership.md) | One evolving project-wide dump and one capture entrypoint | 4, model |
| [TKT-20261017191106](../../ticket/in_progress/20261017191106-clarify-dump-reading-editing-and-appending.md) | Single dump editor and Save | 4 |
| [TKT-20261017192913](../../ticket/in_progress/20261017192913-show-quiz-cards-and-pending-scope-action.md) | Quiz catalog, pending scopes and continuation | 5 |
| [TKT-20261017193927](../../ticket/in_progress/20261017193927-provide-focused-quiz-answering-screen.md) | Quiz interaction, history and modes | 6, model |
| [TKT-20261017194552](../../ticket/in_progress/20261017194552-present-spec-cards-and-user-selected-creation-state.md) | Individual specs, creation state and editing | 7, model |
| [TKT-20261017195458](../../ticket/solved/20261017195458-present-plans-and-tracking-as-read-only-cards.md) | Plan cards and read-only detail | 8 |
| [TKT-20261017195741](../../ticket/in_progress/20261017195741-contextualize-plan-generation-and-check-source-readiness.md) | Contextual generation and sufficient discovery | 8, model |
| [TKT-20261017200232](../../ticket/solved/20261017200232-associate-ticket-tracking-with-each-plan-revision.md) | Required internal ordering document per plan revision | 9, model |
| [TKT-20261017201015](../../ticket/solved/20261017201015-restore-board-page-loading.md) | Board, shared ticket modal, moves and comments | 10, model |
| [TKT-20261017202340](../../ticket/in_progress/20261017202340-implement-live-agent-work-dash.md) | Live work observation | 11 |

Pi integration and the separate Telegram message review are outside this set.
Existing LAN access, branding, section titles and hidden project diagnostics
are the baseline, not new delivery items. This set does not rename trackers or
replace the execution engine.

## 2. Existing Architecture and Required Changes

`internal/core` owns artifact identity, persisted state, mutation exclusion,
ticket lifecycle, ordering and receipts. The central operation service in
`internal/app` owns reusable reads and writes. CLI, TUI, Telegram and Web remain
adapters to those operations. `internal/webui` owns HTTP behavior, view types,
Go templates and the existing CSS, JavaScript and HTMX assets.

| Verified source | Current limitation | Implementation direction |
| --- | --- | --- |
| `internal/app/web_artifacts.go` and `service_spec.go` | Generic artifact presentation and latest-per-scope spec queries do not provide the requested individual-document catalog | Add shared reads of concrete spec/plan documents; keep latest-in-scope lookup separate for generation |
| `internal/app/service_dump.go` and `service_dump_owner.go` | Create allocates another raw/curated pair; authoring exclusion already exists | Resolve the canonical project destination inside shared exclusion; distinguish capture from full-body replacement |
| `internal/app/service_quiz.go` | Shared answer, skip, complete, cancel and mode operations exist | Extend their concurrency bindings and wire the focused Web flow to them |
| `internal/app/service_spec.go` and `internal/core/spec_generation.go` | Generation writes a draft; the required manual create/edit/state flow needs a shared write path | Add manual create and save operations using the same spec identity and persistence owner |
| `internal/app/service_plan.go` | Regeneration archives the existing plan before generating its replacement | Check source readiness first; stage and validate a complete replacement before publishing it |
| `internal/app/service_plan_workflow.go` | Approved native materialization exists, without the newly required order-only companion | Create the companion during generation and bind canonical tickets at guarded materialization |
| `internal/core/workflow_tracker_image.go` | Tracker hashes and renders whole task revisions, processing fields and receipts | Hash and render the structural order map, independently of processing state |
| `internal/app/web_workflow_ticket.go` and `web_views.go` | Native cards are separate from the legacy Board; native preview is capped at 128; Dash begins with the legacy Board read | Resolve the selected authority, create one Board projection with paging, and a native-aware Dash read |
| `internal/webui/assets/templates/shell.html` | Catalogs, detached generation forms and edit surfaces mix technical details with user work | Introduce family-specific catalog/detail partials and the shared modal structures below |

The source tickets record HTTP 500 responses for Board and Dash. Their root
cause has not been established. Implementation must reproduce those failures
with canonical-ops and supported legacy fixtures before changing the read path.
Do not infer that all failures have the same cause or hide errors as empty data.

Preserve the ownership and adoption boundaries of the
[ticket-native workflow](ticket-native-workflow.md) and
[native integration queue](native-integration-queue.md). Section 9
proposes a deliberate change to the optional tracker contract; align that
contract and its tests during implementation, without converting legacy
`tracker_compat` projects or modifying an active controller's frozen inputs.

## 3. Shared Interaction and Visual Contract

1. Retain the shared project header and concise section titles: Dumps, Quizzes,
   Specs, Plans, Board and Dash. Use named scope selectors; identifiers remain
   hidden request bindings, not fields the user must discover and type.
2. Keep normal catalog content focused on documents. Do not foreground artifact
   counts, absolute paths, repeated scope labels, raw frontmatter or generic
   availability badges. Board column counts and Dash readiness counters have
   distinct useful meanings and remain.
3. Use real opening body text for document previews, with a bounded inset area.
   Place that area inside the card below its title, with a very pale green,
   almost-white background subtly distinct from the pure-white card surface,
   inset padding and a subtle boundary. It should resemble the beginning of a
   small document contained within the card. Use a theme token for this pale
   green surface and retain readable text contrast. Apply this same anatomy to
   Dumps, Quizzes, Specs and Plans.
   Parse frontmatter structurally; omit a duplicate document heading when the
   card already displays it. Do not replace the preview with an LLM summary.
   Escape text and use the existing safe Markdown rendering path for detail.
4. Catalog cards always have an explicit fixed width and height per card type.
   This applies to Dumps, Quizzes, Specs, Plans and other document catalogs in
   this set. Content length, title length, missing text and the number of cards
   cannot change those dimensions. Keep title, preview and actions in bounded
   regions with consistent border, padding and focus treatment. Reuse theme
   size tokens rather than introduce another component system.
5. The preview has its own fixed area inside the card and shows only the
   document's first lines. Clip excess text; do not expand the area, add an
   internal preview scrollbar or use an auto-height/min-height-only card.
   Opening the existing detail/editor provides the full content. Short and
   very long documents must produce the same card and preview dimensions.
   Pack fixed-size tracks into the available container width, accounting for
   padding and gaps. Responsive layout changes the number of columns rather
   than stretching cards to fill a row. At the reviewed desktop width, Dumps
   must fit the usable third column at 100% zoom. These are the maintainer's
   explicit card-size and first-lines clarification of 2026-10-18.
6. Use accessible labels, keyboard selection and focus visibility. Modals
   contain focus, restore it to their trigger and expose reachable close and
   submit controls. Maximize/restore preserves the same form and input.
7. Reads and refreshes never start generation, allocate artifacts or claim
   work. A catalog/detail read must not require model access or a live runner.
8. Keep fragment refresh bounded, paused while hidden, and conditional on a
   changed semantic revision. Exclude observation timestamps from that revision
   so unchanged state does not force replacements. Preserve scroll, selection
   and dirty form values; do not hard-reload the page on a polling interval.
9. Show validation, conflict, loading, empty and unavailable states distinctly.
   A failed write retains user input. A stale write cannot overwrite newer
   agent or human changes. Returning to a form must not show another document's
   input or imply an unsuccessful save was committed.

## 4. Dumps: Destination, Capture and Editing

### 4.1 One Logical General Dump

One project has at most one canonical logical project-wide dump. The first
general capture establishes it; later captures add context to it. Enforce this
in the shared writer, including authoring continuations and generated writes,
not by hiding an extra Web button. Serialize destination resolution and write
under the existing project authoring owner.

The canonical reference is project-level, independent of the caller's active
ops namespace. Store the reference through the existing project metadata owner;
the referenced document and its raw/history components remain ops artifacts.
Preserve an existing unambiguous general dump identity. A fresh project creates
the shared destination in the default ops namespace. The model companion
defines reference validation and reconciliation of multiple existing captures.

Use-case captures select an existing case and keep that association. This set
does not introduce a singleton rule for each use case or manufacture cases
from unknown slugs. A raw record, curated body and retained capture history can
belong to one logical dump; they are not separate general catalog cards.

Existing duplicates require a non-destructive, repeatable reconciliation:
inventory content and references, retain original bytes and provenance, publish
one canonical reference and route old references to it. Reads do not migrate
data. Ambiguous or conflicting writes report the required reconciliation
instead of selecting the newest file or allocating another general dump.

### 4.2 New and Save

Use one New action and one capture modal offering Project-wide and existing
case names. Selecting Project-wide reuses the general destination. The title
and content inputs needed by the shared capture operation remain in that form;
New describes adding context, not an obligation to allocate another artifact.

Opening a dump shows one body editor prefilled with its stored body and Save.
Remove the duplicate yellow read-only body and the separate Append form.
Save replaces the full body on the same identity; it never appends that body
as another capture. Preserve frontmatter, raw provenance and other retained
components outside the editable body. The append primitive remains available
to capture adapters that need it.

Move expected-revision comparison into the shared serialized mutation, not
only the Web adapter's preflight. Concurrent saves conflict; concurrent new
general captures retain both contributions in one destination. A retry uses
its existing operation identity rather than appending the same contribution
twice. Rebuild derived indexes after a committed write; an index failure does
not imply that the content was never saved.

## 5. Quizzes: Catalog and Pending Work

Show Project Quiz first, including a No answers yet card when empty. Follow it
with one card per existing case with retained answered entries. Derive answered
membership from actual answers, not file existence or a migration badge.
Partially answered and cancelled sessions with answers remain visible and
continue their retained session from the card.

New Quiz appears at the top right only when at least one eligible scope has
no retained answered entries. Its compact menu lists only those named scopes.
An active unanswered session is pending but selection resumes it; it does not
start a replacement session. With all scopes answered, hide the action. Adding
a new case makes that case pending. Cases with no answers do not receive
placeholder grid cards; Project Quiz is the intentional exception.

Remove detached Start Quiz and Resume Quiz slug forms. A card or pending menu
selection supplies its scope and continuation identity. A completed empty quiz
remains pending for catalog membership; starting a new session after Done must
be an explicit operation that retains the earlier session history.

### Visual Dimensions

Use a centered container of at most 860 CSS px, 20px vertical and 16px
horizontal padding. Align Quizzes and the conditional action on one title row.
The title is approximately 1.3rem/600. Start the grid about 12px below it.

At desktop width use three equal fixed-width tracks and 10px gaps. One or two
cards retain that exact width and never stretch to fill the row. Use a raised
surface, thin border, approximately 12px corner radius and 10px padding. Card
titles are approximately 0.92rem/600. Choose explicit fixed card dimensions
that reserve space for the title and fixed preview; 16:9 is a landscape design
reference, not an instruction to derive variable height from available width.
Long content is clipped within those regions and cannot increase card height.

The very pale green, almost-white inset preview is subtly distinct from the
pure-white card. It has a subtle border, 6px radius, about 7px by 8px padding
and a bounded 96-106px content area displaying up to six useful short lines.
The empty project preview uses the same position and geometry. Use two then
one column as space narrows. The right-aligned pending menu is approximately
190-260px wide, with 8px radius and 6px padding, bounded by the viewport.

## 6. Quiz Interaction, History and Modes

The selected quiz opens a focused screen: scope/title and Deep/Fast selector
at the top; one current question; one large answer textarea; a compact
Skip, Answer, Cancel, Done action row beneath it. Answered question/answer
cards appear below in newest-answer-first order. Skipped questions are not
fabricated answers. Use persisted entries to reconstruct this view after
reload or continuation.

| Action | Shared result |
| --- | --- |
| Answer | Record the typed answer for the bound current question exactly once, then advance or generate the applicable follow-up |
| Skip | Advance without turning the textarea into an answer |
| Cancel | Exit with native cancellation semantics; retain recorded history and do not mark Done |
| Done | Complete the quiz even with remaining questions; retain history and stop further generation |

Bind each write to scope, retained session, current question, expected revision
and operation identity. The shared quiz owner validates those bindings under
exclusion. Duplicate requests return the committed result; a stale request
cannot answer whichever question became current in the meantime.

Preserve a typed but unsent response through mode changes, refresh and request
errors. Skip, Cancel, Done and navigation with an unsent response require a
clear discard choice; none silently submits it. Browser form state need not
become a canonical answer record. After a successful Answer, clear only that
question's draft. No autosave implies that Answer occurred.

### Proposed Queue Transition Policy

Deep retains adaptive nested follow-ups within the existing depth and budget
bounds. Fast creates a flat question batch in one successful model generation
request and retains its queue. Answering or skipping that queue makes no new
generation call. Provider failures retain their normal bounded retry semantics;
they do not count as successfully generating several batches.

A mode switch keeps the active question, its draft and all recorded answers.
Changing Deep to Fast retains the current question and generates the remaining
flat batch once if no retained Fast batch exists. Do not duplicate the active
question or already answered questions in that batch. Changing Fast to Deep
keeps the current question, suspends the remaining Fast queue and enables
adaptive follow-ups after it. Returning to Fast resumes its retained unanswered
queue without generating it again. An exhausted Fast queue permits Done; it
does not silently restart or fall back to Deep. Persist mode and queue changes
atomically so restart cannot duplicate a successful batch.

## 7. Specs: Individual Documents and Shared Authoring

Catalog concrete spec documents, not one latest document or missing placeholder
per case. Keep latest-in-scope selection as a separate generation query. Each
card has its title and bounded opening body text. Both manually written and
generated specs appear in this catalog. Clicking a card opens that exact spec
with its body prefilled and Save.

Use one New action. Its modal contains Title, a named Use Case selector with
Project-wide, one large Body editor, a compact themed Draft switch and Create
at the lower right. Draft on creates draft; off creates approved directly.
There is no separate Draft document family or forced draft/promotion journey.
Default the switch on, while honoring an explicit off selection.

Add shared manual create/save operations alongside GenerateSpec. Manual
creation needs no dump, quiz or model request. Generation assembles project
context, the canonical general dump and relevant scoped dump/quiz when present;
exclude unrelated case material. It remains an explicit operation and defaults
to draft. The minimal manual modal does not trigger generation or approve
generated content automatically.

Use `draft` and `approved` as the write vocabulary. Normalize existing
`accepted`/`final` approval aliases on read at the shared owner, rather than
create a Web-only state. Save the selected native state and required lifecycle
metadata. Editing retains identity, association and provenance; an approved
spec can be edited without creating a precursor draft. Its changed source hash
makes dependent generation bindings stale, without rewriting earlier plan
history or granting execution authority.

## 8. Plans: Reading and Contextual Generation

Plans lists actual plan documents as title/opening-text cards, using the Specs
card anatomy. Exclude tracker documents, missing placeholders and the Artifacts
counter. Clicking a card opens the complete read-only plan. This Web browsing
surface has no plan body editor, Save or Append; agent-owned generation and
guarded amendments remain possible through their existing owners.

Put Generate Plan or Regenerate Plan in the selected case's detail context,
where its identity and existing plan are known. Remove both global slug forms
from Plans. Generate is available when no plan exists; Regenerate targets an
existing plan and its observed revision. Keep existing project-scoped API
capabilities, but do not add a new project-wide Web planning workflow here.

### Shared Readiness and Publication

1. Assemble the selected case, project context, relevant general/scoped dump,
   retained quiz answers and current applicable spec. At least one useful dump
   contribution or answered quiz entry must inform this case. A manual spec
   alone does not waive this requested discovery prerequisite.
2. Distinguish absent discovery from insufficient discovery. Empty placeholders,
   metadata-only files, skipped questions and generic text unrelated to the
   case do not establish readiness. Neither a fixed word count nor requiring
   both dump and quiz is valid.
3. Assess whether the assembled context identifies the intended outcome,
   bounded scope, necessary behaviors/constraints and verifiable results well
   enough to plan without inventing a material decision. Reuse the existing
   structured generation boundary for the semantic assessment, with explicit
   sufficient/insufficient output and missing-information reasons. A narrow
   case can be ready from a short dump; a complete dump can avoid a quiz.
4. Enforce this check in the shared generation operation before any existing
   plan is archived or replaced. Report missing information and link the Web
   user to the selected case's dump or quiz. Other adapters receive the same
   decision. Model/error/invalid assessment results do not imply readiness.
5. Bind generation to the observed source identities and hashes. Stage and
   validate the new plan and companion; recheck sources and target revision
   before publishing under the existing mutation owner. Preserve the current
   plan on readiness failure, cancellation, invalid output or stale sources.
6. Extra discovery can improve a spec and then a plan, but these are distinct
   operations. Regenerate Plan does not silently regenerate or approve a spec,
   adopt a workflow or authorize new ticket execution. If an applicable
   required spec/approval is missing or stale, explain that prerequisite.

## 9. Internal Tracker as an Ordering Document

Every successful plan generation/regeneration produces an associated Markdown
tracker for that exact plan revision. Keep the tracker name, artifact family
and compatible table structure. It is internal: no card, detail route, editor,
Create Tracking form or alternative export in the Web UI, including direct
tracker URLs through generic artifact endpoints.

The document only describes the ticket sequence. Native set, slice and task
records remain authoritative for ordering; tickets and receipts remain
authoritative for progress. Order slices by `slice_order` and their tasks by
`task_order`, with stable identities, titles and resolvable references. Include
the current map in implementation context and read live ticket state to find
the remaining work. Never parse tracker checkboxes to schedule or finish tasks.

Generation does not itself authorize ticket allocation or execution. Before
approved materialization the companion records planned stable identities and
order, explicitly unbound to tickets. It cannot be used for execution. Approved
materialization idempotently binds those entries to allocated canonical tickets
and publishes the executable map; it does not ask for a separate Web tracking
operation. Preserve per-plan-revision maps and their provenance.

Export, freshness and repair must depend only on plan binding, membership,
declared order, titles and references. Whole ticket hashes/revisions, phases,
claims, statuses and delivery receipts are not map dependencies. Advancing a
ticket must leave the tracker bytes and its validity unchanged. Retained legacy
progress columns are blank/non-authoritative; do not maintain them or attach
new completion checks to them.

Use versioned projection metadata for the changed hash semantics. Preserve
observed-output conflict checks: a manually edited map is not imported as
authority or overwritten without the supported observed-hash replacement.
Regeneration retains prior bindings and delivery history. Changes to active or
delivered work use existing guarded amendment rules, with no duplicate tasks,
reset completions or silent replacement of an active set. A missing required
map blocks preparing that plan's execution context until repaired; a stale
derived map does not roll back an already committed ticket transition.

## 10. Board and Shared Ticket Modal

### Board Read and Layout

Select canonical ticket authority before constructing the Board. Native tasks
and supported loose tickets appear once each, with native state mapped through
shared lifecycle rules. A native record wins over its own derived legacy
projection. Do not force files back into an obsolete directory or create a
second browser ticket store. Include explicit pagination/continuation for
bounded reads and keep total column counts distinct from loaded-page counts.

Retain the six lanes: Backlog, Ready For Dev, In Progress, Ready For Review,
Ready For Merge, Done. Preserve shared ordering and legacy aliases. Blocked,
invalid or other attention cases remain reachable with their reasons outside
the normal lanes. Neither visual ordering nor dragging changes declared slice
or task execution order.

Align Board and one New action. Use 16-24px outer spacing, 12-16px lane gaps and
roughly 260-360px readable lanes. Scroll horizontally when needed rather than
squeeze six lanes into strips. Keep lane headings/counts and long columns
reachable. Cards have a raised surface, subtle border, 10-12px radius and
12-14px padding: key/title, compact priority, description excerpt, sparse
tags/assignment. Use existing measured timing data only when meaningful; do
not label age as work duration or invent zero timers.

### One Create/Edit Form

Card selection opens a populated edit modal; New opens the same form fresh.
Use Create or Save changes according to mode. Show Title and one full
Description editor, then grouped metadata. Derive the required Synopsis
deterministically from that description where needed, without another model
call or a competing editor. The full body remains editable and preserved.

| Group | Fields and guards |
| --- | --- |
| Workflow | Allowed native Status transitions; distinct system-owned phase read-only |
| Classification | Named Use Case, Priority, Kind/type, Severity, Scope/project area, Source and applicable Resolution |
| Ownership | Descriptive Assignee and Owner; never execution claim tokens |
| Delivery | Branch, PR and multiline commit references through shared evidence guards; verified receipt data read-only |
| Discussion | Existing comments, author/time and Add comment as a separate operation |

Extend shared create/update operations for missing supported fields; map
repository-development vocabulary and runtime vocabulary explicitly rather
than silently writing the wrong enum. Preserve unknown metadata, frontmatter,
comments, immutable associations and evidence. New allocates through the
canonical owner and does not enroll itself into a slice set by selecting a
case. Native task membership and claim ownership cannot be edited by this form.

On large screens, normal modal width is about 900-1000px within the viewport,
with internal scrolling and two-column short metadata when space permits.
Maximized mode places the large body editor left and grouped metadata right.
Narrow screens stack fields. Close, submit and every retained field remain
reachable; maximize/restore never resets input or selected identity.

Both the Status selector and drag/drop call the same guarded transition with
the observed ticket revision. A rejected or stale move restores the actual
column and gives a readable reason. Moving a grouped task to Done cannot
fabricate development, review, merge or closure evidence.

Comments use their existing canonical owner and ordering. Submitting one
neither saves the other form nor discards its unsaved edits. New mode never
inherits another ticket's discussion. Refresh changes the affected cards and
counts while preserving Board position and any open dirty modal.

## 11. Dash as a Live Observation

Create an adapter-neutral observation from canonical selection, invocation,
attempt, event and result reads. Ticket authority mode uses native workflow
state; the supported legacy mode uses its own observation path. Dash must not
depend on a TUI process, a development controller, tracker progress cells or
the success of an unrelated legacy Board read.

Use queue left, current work upper right, and activity/results below, stacking
at narrow widths. Keep readiness compact near current work. The primary
information is the ticket and phase, not artifact counts or raw identifiers.

1. Queue shows eligible work with key/title/scope, shared prerequisites and
   native order. Distinguish no work, excluded work and an unavailable source.
   A display filter never changes readiness or native scheduling.
2. Current shows active invocation, ticket, phase and meaningful tool or
   validation activity when events provide it. A quiet validation remains
   active with its observation time; silence is not failure or completion.
3. Concurrent invocations remain separate by project, invocation/attempt and
   ticket binding. Show multiple active contexts or an explicit selector; do
   not combine the latest event from one run with another run's status.
4. Retain bounded recent events without duplicate entries and the latest known
   completed/interrupted result even with no latest-phase event or remaining
   queue. Distinguish harness success from integrated delivery. Show actual
   processed/blocked/failed counts only where their definitions are supported.
5. Read failures retain the last observation with stale/unavailable labels.
   Reopening or recovery reconstructs persisted state without a new run.

Browsing and refresh never claim, start, stop or mutate work. Retained Preview
Work and Stop Daemon controls keep their existing targets, confirmations and
guards; daemon stop is not redefined as stopping one ticket. The active-action
overlay must not hide the monitoring panels while work runs.

## 12. Implementation Order and Verification

This is the dependency order for later planning, not a new execution tracker:

1. Repair canonical/shared read projections for Board and Dash, retaining
   error distinctions and supported legacy behavior. Add concrete artifact
   listing, body previews, named scopes and reusable conflict-safe form binding.
2. Implement project-wide dump identity/reconciliation and shared save guards;
   then simplify capture/editing and correct the Dumps grid.
3. Implement quiz read membership, guarded actions, mode queues and retained
   history; then the catalog, pending menu and focused answer screen.
4. Implement shared manual spec creation/edit/state persistence and context
   assembly; then the spec catalog and minimal modal.
5. Implement shared plan readiness and safe staged regeneration, followed by
   revision-bound order-only trackers and approved ticket materialization.
   Then simplify Plans and remove all Web tracker exposure.
6. Complete Board form fields, create/edit, guarded moves, comments and paging;
   finish Dash observation/refresh. Shared lifecycle logic remains in its owner.
7. Review complete journeys and visual consistency across the changed surfaces.
   Record each source ticket's acceptance evidence before resolving it. The
   review parent closes only after the consolidated review and child evidence.

| Acceptance group | Required evidence |
| --- | --- |
| UI-01: catalog composition | Dumps at the same desktop width at 100%/90% zoom; responsive fixed-size tracks; Quizzes menu open/closed/all-answered; title above a very pale green, almost-white inset opening-line preview subtly distinct from every pure-white document card; short, long and unbroken content has identical card/preview dimensions, no metadata clutter and no preview scrollbar |
| UI-02: dump integrity | First capture, later capture, two concurrent captures, retry, Save/reopen, stale Save, namespace switch, restart/index rebuild and reconciliation preserving source bytes/references |
| UI-03: quiz state | Empty/partial/answered/cancelled scope membership; correct question binding; all four actions; retained history; draft-preserving switches; one successful Fast batch generation and no per-question calls |
| UI-04: spec authoring | Multiple documents per scope; manual draft/approved creation without model access; generated/manual editing of the same identity; correct context isolation and retained state |
| UI-05: plan workflow | Missing/insufficient discovery rejected; complete dump or narrow sufficient dump or quiz-only discovery accepted; stale/model failure leaves prior plan intact; no implicit spec regeneration |
| UI-06: ordering contract | Companion per plan revision; no allocation before approval; idempotent binding; slice/task order independent of discovery order; ticket progress leaves tracker bytes and freshness unchanged |
| UI-07: Board | Native/legacy/empty/mixed fixtures; one card per identity; attention cases and more than 128 tickets reachable; normal/maximized/new modals; persistence, comments, allowed/rejected/stale moves |
| UI-08: Dash | Queued/running/quiet/blocked/failed/stopped/completed states; two concurrent contexts; temporary outage/restart; no phase event; live refresh with no state mutations |
| UI-09: usability review | End-to-end navigation, feedback, loading, empty/errors, keyboard/focus, readable spacing/typography and narrow viewport behavior; no dirty input lost to refresh or comments |

Use deterministic service/core regressions for persistence and lifecycle
decisions, focused HTTP/template checks for binding/errors and a bounded
browser journey for the visible flows. Capture desktop and narrow viewport
screenshots at 100% zoom, recording CSS viewport and device pixel ratio. A
successful HTTP response alone does not establish the visual result.

During implementation, run the repository's `make verify` development gate
and the affected focused integration/race selections where behavior requires
them. Complete exhaustive, race and live-provider campaigns remain with the
scheduled Forgejo nightly; do not add them to each edit or ticket gate.
This specification-only change requires metadata, cross-reference, scope and
whitespace validation, without runtime tests or live model requests.

## 13. Design Status

The user's interaction and ownership requirements above are settled inputs.
The canonical project reference, reconciliation policy, quiz queue transitions,
manual spec write model, semantic readiness assessment and tracker schema
evolution are accepted implementation decisions, detailed in the companion.
The linked plan and tracker define their authorized delivery order. No
unresolved product question is delegated to an implementing agent by an
unexplained reference to another application.
