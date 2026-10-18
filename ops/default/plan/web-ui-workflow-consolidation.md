# Web UI and Workflow Delivery

Status: Approved

Date: 2026-10-18

Approved: maintainer, 2026-10-18; the explicit request to proceed with Looper accepts this set and its implementation design.

Delivery set: web-ui-workflow-consolidation

Spec: [Web UI and Workflow Consolidation](../spec/stable/web-ui-workflow-consolidation.md)

Model: [Web UI and Workflow Model](../spec/stable/web-ui-workflow-consolidation-model.md)

Tracker: [Web UI and Workflow Delivery Order](../tracker/web-ui-workflow-consolidation.md)

Planning baseline: `4ea9beb236d3` (specification and selected source tickets).

Implementation authorization: granted on 2026-10-18 for all twelve recorded
slices through Looper, including canonical slice PRs and controller-owned
Rebase + Fast-forward merges into dev. The existing Telegram run is separate.
Use one development worker and one canonical open slice PR for this set.

Active slice: Slice 12, Consolidated usability review; T12.1.
The controller checkpoint and tickets own subsequent progress.

First planned slice: Slice 1, Board and shared ticket modal; T1.1.

Slice strategy: behavior-first

Reason: retain one existing ticket per complete, independently reviewable
behavior increment. Each increment includes its required domain, persistence,
adapter and UI work; avoid separate skeleton or deliberately failing slices.

Development gate: `make verify`, affected focused checks and
`git diff --check` for the exact candidate.

Exhaustive certification: the existing scheduled Forgejo nightly. Do not run
an interactive full suite as a substitute.

## 1. Delivery Units and State Ownership

Maintainer clarification, 2026-10-18: all document catalog cards have fixed
width and height per type, including Dumps, Quizzes, Specs and Plans. Their
preview areas also stay fixed and show only opening lines with clipped overflow.
The preview sits inside the card below its title, in a very pale green,
almost-white inset surface subtly distinct from the pure-white card. It
resembles the beginning of a small document.
Responsive grids change column count, not card geometry. Apply this contract in
the owning UI slices and verify identical geometry for short/long bodies in
the final review. This clarifies existing acceptance; slice identities and
execution ownership remain unchanged.

This plan contains twelve ordered slices, one branch and one PR targeting
`dev` per existing source ticket. The tracker supplies deterministic order and
commit-sized task instructions. Those instructions do not create more tickets.
The source tickets retain severity medium, priority normal and their currently
selected in-progress status; none is complete merely because it is planned.

Progress, implementation commits, blockers, PRs and acceptance evidence belong
in the existing source ticket and its slice report. This delivery-order tracker
has no duplicate task-state or completion checklist. Keep its structural map
unchanged when work progresses; amend it only for approved changes of scope,
ordering, dependency or delivery identity.

Slice 1 delivers Board's complete ticket scope first, including the shared
read/projection repairs needed by later views. Slice 5 delivers retained quiz
actions before Slice 6 depends on them for catalog continuation. Slice 8 adds
safe generation before Slice 9 attaches required order companions, and Slice
10 then removes their Web exposure. These dependency choices refine the
specification's high-level sequence without changing its scope or owners.

## 2. Entry Conditions and Boundaries

1. Approval and execution authority are recorded above. Commit and publish
   the approved specification, model, plan and tracker before creating Slice 1.
2. At activation verify current dev/upstream, source-ticket identities and
   unrelated work, existing controller/validation ownership, and the actual
   installed execution tool. Start only this approved set under its explicit
   Looper delegation.
3. Deliver this set sequentially, with one active development slice and at most
   one open canonical slice PR for this set. Product-level concurrency remains
   supported and is exercised in isolated fixtures, not disabled.
4. Preserve the separate Telegram delivery run and its workspace, reports,
   validation and private setup. Neither completion nor resumption of that run
   belongs to this plan. Pi and Telegram message-review tickets remain outside.
5. Use temporary project stores and deterministic model/transport boundaries.
   No live bot, paid model request, external repository write or production
   migration is a prerequisite for these implementation regressions. Private
   credentials and incidental machine identifiers stay outside artifacts.
6. Reconciliation is a deliverable shared preview/apply operation. Test it on
   fixtures; do not apply it to the user's project data merely to implement or
   demonstrate the UI. Existing documents and references remain preserved.
7. Keep core persistence/lifecycle owners, the central application service,
   existing server-rendered theme and HTMX. Do not introduce a second scheduler,
   ticket store, browser-owned quiz runtime or Web-only generation policy.

## 3. Canonical Slice Map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Board and shared ticket modal | `feat/slice-1-web-board` | `feat(slice-1): deliver the Board and shared ticket editor` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-1-web-board.md` |
| Slice 2 | Canonical project dump | `feat/slice-2-project-dump` | `feat(slice-2): retain one evolving project-wide dump` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-2-project-dump.md` |
| Slice 3 | Single dump editor | `fix/slice-3-dump-editor` | `fix(slice-3): use one dump editor and Save` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-3-dump-editor.md` |
| Slice 4 | Dumps grid width | `fix/slice-4-dump-grid` | `fix(slice-4): use the available Dumps catalog width` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-4-dump-grid.md` |
| Slice 5 | Focused quiz answering | `feat/slice-5-quiz-answering` | `feat(slice-5): deliver focused quiz answering and modes` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-5-quiz-answering.md` |
| Slice 6 | Scope-driven quiz catalog | `feat/slice-6-quiz-catalog` | `feat(slice-6): show quiz cards and pending named scopes` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-6-quiz-catalog.md` |
| Slice 7 | Individual spec authoring | `feat/slice-7-spec-authoring` | `feat(slice-7): create and edit individual specs` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-7-spec-authoring.md` |
| Slice 8 | Contextual safe plan generation | `feat/slice-8-plan-generation` | `feat(slice-8): generate plans from sufficient use-case context` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-8-plan-generation.md` |
| Slice 9 | Revision-bound order-only trackers | `feat/slice-9-plan-order-tracker` | `feat(slice-9): bind internal order maps to plan revisions` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-9-plan-order-tracker.md` |
| Slice 10 | Read-only Plans catalog | `feat/slice-10-plan-catalog` | `feat(slice-10): present plan cards and read-only detail` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-10-plan-catalog.md` |
| Slice 11 | Live work Dash | `feat/slice-11-work-dash` | `feat(slice-11): observe live ticket work in Dash` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-11-work-dash.md` |
| Slice 12 | Consolidated usability review | `test/slice-12-web-ui-review` | `test(slice-12): verify the consolidated Web UI journeys` | `ops/default/report/slices/web-ui-workflow-consolidation/slice-12-web-ui-review.md` |

## 4. Slice Contracts

The tracker defines the exact task IDs and ordering for each slice. The
following contracts define its delivered behavior and required evidence.

### Slice 1: Board and shared ticket modal

Source: [TKT-20261017201015](../ticket/solved/20261017201015-restore-board-page-loading.md).

Dependency: approved and committed planning inputs.

Spec sections: 2, 3, 10; model 1 and 7.

Contract: A populated, empty or mixed-authority project opens one usable Kanban Board. New and existing tickets share one guarded editor, and status moves and comments preserve canonical ownership.

Validation: Cover canonical-only, legacy, mixed and empty stores, more than 128 tickets, all retained fields, stale and rejected writes, comments with dirty form input, and desktop/narrow modal screenshots.

### Slice 2: Canonical project dump

Source: [TKT-20261017190508](../ticket/solved/20261017190508-unify-dump-capture-and-project-wide-ownership.md).

Dependency: Slice 1 integrated into dev; use all earlier shared owners.

Spec sections: 4.1 and 4.2; model 1 and 2.

Contract: Every supported general capture resolves one project-wide logical dump across namespaces. Existing content and references have an explicit non-destructive reconciliation path.

Validation: Cover first/repeated/concurrent capture, retry, namespace changes, interrupted publication/reconciliation, old references, preserved originals and restart/index rebuild; focused race coverage for shared destination writes.

### Slice 3: Single dump editor

Source: [TKT-20261017191106](../ticket/solved/20261017191106-clarify-dump-reading-editing-and-appending.md).

Dependency: Slice 2 integrated into dev; use all earlier shared owners.

Spec sections: 3 and 4.2; model 1 and 2.

Contract: Opening a dump presents one body editor and Save, preserving its identity and rejecting stale replacement without duplicating or losing text.

Validation: Exercise successful and stale full-body replacement, unsaved input, raw/history preservation and the absence of duplicate body/Append surfaces; keep append available to capture adapters.

### Slice 4: Dumps grid width

Source: [TKT-20261017185837](../ticket/solved/20261017185837-fix-dumps-grid-width-at-default-zoom.md).

Dependency: Slice 3 integrated into dev; use all earlier shared owners.

Spec sections: 3 and 4.

Contract: Dumps distributes cards across available desktop width at 100% zoom and remains readable at smaller widths.

Validation: Record CSS viewport and device pixel ratio. At the reviewed desktop width show the usable third column at 100%; do not force three columns at narrow widths.

### Slice 5: Focused quiz answering

Source: [TKT-20261017193927](../ticket/solved/20261017193927-provide-focused-quiz-answering-screen.md).

Dependency: Slice 4 integrated into dev; use all earlier shared owners.

Spec sections: 6; model 1 and 3.

Contract: One retained quiz exposes the current question, one answer input, all four actions, newest-first answer history and real Deep/Fast transitions.

Validation: Cover duplicate/stale Answer and Skip, Done with queued questions, cancellation, answer committed before continuation failure, mode switch/restart and one successful Fast batch with no per-question generation; focused race and browser checks.

### Slice 6: Scope-driven quiz catalog

Source: [TKT-20261017192913](../ticket/solved/20261017192913-show-quiz-cards-and-pending-scope-action.md).

Dependency: Slice 5 integrated into dev; use all earlier shared owners.

Spec sections: 5 and 6; model 3.

Contract: Project Quiz appears first, answered cases have cards, and New Quiz only offers pending named destinations while preserving unfinished sessions.

Validation: Cover empty, partial, answered, cancelled and completed-empty sessions, adding a case and all-answered action removal; capture closed/open menu, long content and narrow views at declared viewport/zoom.

### Slice 7: Individual spec authoring

Source: [TKT-20261017194552](../ticket/solved/20261017194552-present-spec-cards-and-user-selected-creation-state.md).

Dependency: Slice 6 integrated into dev; use all earlier shared owners.

Spec sections: 3 and 7; model 1 and 4.

Contract: Users create draft or approved specs directly and edit manual/generated documents through the same identity-preserving shared owner and catalog.

Validation: Create multiple specs in one scope, one draft and one directly approved document, edit both manual/generated identities, verify stale writes and selected lifecycle after reload, and capture catalog/modal views.

### Slice 8: Contextual safe plan generation

Source: [TKT-20261017195741](../ticket/solved/20261017195741-contextualize-plan-generation-and-check-source-readiness.md).

Dependency: Slice 7 integrated into dev; use all earlier shared owners.

Spec sections: 8; model 5 and 6.

Contract: Generate/Regenerate is bound to a named case and sufficient discovery; failed readiness, generation or stale sources leave the current plan usable.

Validation: Cover no discovery, irrelevant/placeholder data, insufficient context, complete-dump-only, narrow short dump and quiz-only cases; model failure, cancellation and stale source/target must preserve the prior plan.

### Slice 9: Revision-bound order-only trackers

Source: [TKT-20261017200232](../ticket/solved/20261017200232-associate-ticket-tracking-with-each-plan-revision.md).

Dependency: Slice 8 integrated into dev; use all earlier shared owners.

Spec sections: 9; model 6 and 8.

Contract: Each successful plan revision has its own internal order map; approved materialization binds canonical tickets idempotently, and processing never invalidates the map.

Validation: Verify plan/map interruption and retry, no allocation before approval, stable slice/task ordering, edited output, active-work amendment guards, historical bindings and byte-identical current tracker through ticket progress; focused race checks.

### Slice 10: Read-only Plans catalog

Source: [TKT-20261017195458](../ticket/solved/20261017195458-present-plans-and-tracking-as-read-only-cards.md).

Dependency: Slice 9 integrated into dev; use all earlier shared owners.

Spec sections: 3, 8 and 9.

Contract: Plans lists actual plan cards with opening-text previews and full read-only detail; internal tracker documents and generic creation controls stay absent.

Validation: Verify short/long plan previews, exact full document selection, no body write controls or reads that mutate data, and unavailable direct tracker views; capture catalog and read-only detail screenshots.

### Slice 11: Live work Dash

Source: [TKT-20261017202340](../ticket/solved/20261017202340-implement-live-agent-work-dash.md).

Dependency: Slice 10 integrated into dev; use all earlier shared owners.

Spec sections: 3 and 11; model 7.

Contract: Dash observes native or supported legacy queue, current activity and retained results without depending on a TUI, tracker progress or an unrelated legacy Board read.

Validation: Cover queued, quiet, blocked, failed, stopped and completed work, two active contexts, no phase event, temporary read outage/reopen/restart and mutation-free polling; focused concurrency and desktop/narrow browser evidence.

### Slice 12: Consolidated usability review

Source: [TKT-20261017105432](../ticket/in_progress/20261017105432-review-draft-web-ui-usability.md).

Dependency: Slice 11 integrated into dev; use all earlier shared owners.

Spec sections: 1, 3 and 12; model 8.

Contract: The selected Web journeys have coherent functional, visual and accessibility evidence, and every child ticket's acceptance is accounted for before the review parent resolves.

Validation: Map UI-01 through UI-09 to child evidence and screenshots, verify all twelve source tickets, execute make verify for the final integrated candidate, and report scheduled exhaustive evidence separately without starting live work.

## 5. Validation and Review Evidence

1. For each task run the smallest meaningful affected regressions. For each
   slice run `make verify`, its specifically selected integration checks when
   changed IO/workflow behavior requires them, then `git diff --check`.
   Use focused race checks where shared mutation, queue, observation or map
   publication concurrency changes. Do not run a repository-wide race suite.
2. Record exact successful commands and candidate hashes in the source ticket
   and slice report. New test selectors must actually match the new and affected
   regressions; an empty test selection is not evidence.
3. Each UI slice supplies desktop and narrow viewport browser evidence at
   100% zoom. Record CSS viewport/device pixel ratio and relevant 90% Dumps
   comparison. Reuse canonical deterministic fixtures rather than operating
   live project work to obtain screenshots.
4. Reuse prior unchanged evidence where it proves the same contract; rerun
   affected checks when source or bindings change. Never relabel a focused
   pass as exhaustive or reuse validation of a different candidate HEAD.
5. After the last slice integrates, run `make verify` for that exact dev
   candidate and correlate UI-01 through UI-09 with source-ticket/report
   evidence. The exhaustive campaign stays in its scheduled nightly slot;
   distinguish development delivery from any pending nightly certification.
6. A report path in the map is an expected output, not an existing report or
   claim of completion. Create the report with actual implementation evidence.
   Keep only real execution/PR results in the ticket and report.

## 6. Source Acceptance Coverage

| Acceptance group | Primary slices | Evidence responsibility |
| --- | --- | --- |
| UI-01 | 4, 6, 7, 10 | Catalog geometry, previews, menu states and readable responsive layout |
| UI-02 | 2, 3 | Singleton identity, reconciliation, concurrency/retry and full-body editing |
| UI-03 | 5, 6 | Retained actions/history, membership and actual Deep/Fast behavior |
| UI-04 | 7 | Manual/generated spec identity, lifecycle, scope and editing |
| UI-05 | 8 | Discovery sufficiency, contextual actions and non-destructive regeneration |
| UI-06 | 9 | Revision-bound ordering, materialization and unchanged map through ticket progress |
| UI-07 | 1 | Canonical Board, paging, modal, moves, comments and field preservation |
| UI-08 | 11 | Concurrent observation, quiet/final results, stale recovery and read-only refresh |
| UI-09 | 12, with evidence from 1-11 | Complete journey, accessibility and visual/functional review |

The parent review ticket is owned by Slice 12. Earlier slices resolve only
their own accepted ticket after verified integration. Slice 12 must not invent
completion for an unresolved child or close the parent from screenshots alone.

## 7. Controller and Closure Boundaries

Looper owns validation, canonical PR verification and Rebase + Fast-forward
merges into dev. Workers prepare candidates and reports, and never merge
themselves. Use `make verify` as the required first validation command and
final development gate; add only affected focused checks, ending with
`git diff --check`. Keep exhaustive and live certification in the scheduled
nightly. Do not start or alter that schedule to complete an interactive gate.

Move each implemented source ticket and repair its links before that slice
candidate is validated. Preserve final old/new ticket paths in the local
configuration closure allowlist for administrative closure. Final set closure
may change only the explicitly configured artifact/report/ticket paths and
sets plan/tracker metadata to Completed while retaining Approved metadata and
the canonical order map. Ticket processing never requires a tracker progress
row or an export of runtime state into this document.
