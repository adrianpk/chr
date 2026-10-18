# Web UI and Workflow Delivery Order

Status: Approved

Date: 2026-10-18

Approved: maintainer, 2026-10-18; the explicit request to proceed with Looper accepts this set and its implementation design.

Delivery set: web-ui-workflow-consolidation

Plan: [Web UI and Workflow Delivery](../plan/web-ui-workflow-consolidation.md)

Spec: [Web UI and Workflow Consolidation](../spec/stable/web-ui-workflow-consolidation.md)

Model: [Web UI and Workflow Model](../spec/stable/web-ui-workflow-consolidation-model.md)

Implementation authorization: granted on 2026-10-18 for this exact set through
Looper, including controller-owned Rebase + Fast-forward integration into dev.

Active slice: Slice 2, Canonical project dump; T2.1.
The controller checkpoint and tickets own subsequent progress.

First planned slice: Slice 1, Board and shared ticket modal; T1.1.

## Ordering Contract

Required catalog presentation, clarified 2026-10-18: fixed width/height per card
type and a fixed preview of opening lines with clipped overflow. Content never
enlarges a card or preview; responsive grids vary column count. The owning
catalog slices and final review verify this under the linked specification.

Read slices in ascending numeric order and tasks in their listed order. Every
slice uses its one existing source ticket, one branch and one PR targeting dev.
Task IDs are commit-sized instructions within that ticket, not extra tickets.
Do not infer execution order from filename timestamps, file discovery, display
sorting or the tickets' shared in-progress classification.

This document defines scope, dependency and delivery identity. Ticket records
and slice reports carry processing state, implementation commits, blockers,
PR references and validation results. Do not add progress checkboxes, task-state
columns, current processing phases or mutable completion counters here.
Update this map only when its declared order or scope is deliberately amended.

The set is sequential. Integrate the preceding slice before starting its
successor. The existing product remains capable of concurrent runs; tests must
preserve those ownership and stale-write boundaries.

## Canonical Slice Map

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

## Ordered Ticket and Task Map

### Slice 1: Board and shared ticket modal

Ticket: [TKT-20261017201015](../ticket/solved/20261017201015-restore-board-page-loading.md).

Prerequisite: approved committed specification/model/plan/tracker and explicit execution authority.

| Task | Ordered work |
| --- | --- |
| T1.1 | Repair canonical/native/legacy reads, deduplicate projections, retain attention cases and page past 128 tickets without disguising failures as empty data. |
| T1.2 | Extend shared ticket creation, body/metadata, classification, ownership and delivery-reference writes; preserve unknown metadata and enforce native transition and revision guards. |
| T1.3 | Deliver the six lanes, readable cards, shared normal/maximized New/Edit modal, selector and drag moves, and independent comments without losing dirty input. |

Validation: Cover canonical-only, legacy, mixed and empty stores, more than 128 tickets, all retained fields, stale and rejected writes, comments with dirty form input, and desktop/narrow modal screenshots.

### Slice 2: Canonical project dump

Ticket: [TKT-20261017190508](../ticket/in_progress/20261017190508-unify-dump-capture-and-project-wide-ownership.md).

Prerequisite: Slice 1 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T2.1 | Add and validate the project-level canonical dump reference under the existing metadata and authoring owners; preserve unambiguous existing identity and raw/history components. |
| T2.2 | Route all general capture writers and retries through that destination; implement conflict-safe explicit reconciliation preview/apply with retained original bytes, provenance and alias resolution. |
| T2.3 | Replace the two Web entrypoints with one New capture form offering Project-wide and existing named cases; preserve case-scoped behavior. |

Validation: Cover first/repeated/concurrent capture, retry, namespace changes, interrupted publication/reconciliation, old references, preserved originals and restart/index rebuild; focused race coverage for shared destination writes.

### Slice 3: Single dump editor

Ticket: [TKT-20261017191106](../ticket/in_progress/20261017191106-clarify-dump-reading-editing-and-appending.md).

Prerequisite: Slice 2 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T3.1 | Bind full-body Save to the same canonical artifact with revision comparison inside the serialized shared mutation; preserve metadata and provenance. |
| T3.2 | Remove the duplicate read-only body and separate Append form, preserve pending edits on error/refresh, and verify edit, add text, save and reopen. |

Validation: Exercise successful and stale full-body replacement, unsaved input, raw/history preservation and the absence of duplicate body/Append surfaces; keep append available to capture adapters.

### Slice 4: Dumps grid width

Ticket: [TKT-20261017185837](../ticket/in_progress/20261017185837-fix-dumps-grid-width-at-default-zoom.md).

Prerequisite: Slice 3 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T4.1 | Reproduce the unused-width defect and correct measured container, track and breakpoint sizing without changing capture semantics. |
| T4.2 | Capture the same populated catalog and viewport at 100% and 90% zoom plus narrow views; verify readable titles, spacing and overflow. |

Validation: Record CSS viewport and device pixel ratio. At the reviewed desktop width show the usable third column at 100%; do not force three columns at narrow widths.

### Slice 5: Focused quiz answering

Ticket: [TKT-20261017193927](../ticket/in_progress/20261017193927-provide-focused-quiz-answering-screen.md).

Prerequisite: Slice 4 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T5.1 | Extend shared quiz action bindings and versioned retained state for question/session identity, revisions, idempotent actions and resumable mode queues. |
| T5.2 | Implement the specified queue policy: adaptive Deep, one successfully generated Fast batch, preserved active question/history and suspended/resumed queues. |
| T5.3 | Wire the focused Web answer screen and mode selector, including dirty-draft/discard handling, errors, reload and independent Cancel/Done semantics. |

Validation: Cover duplicate/stale Answer and Skip, Done with queued questions, cancellation, answer committed before continuation failure, mode switch/restart and one successful Fast batch with no per-question generation; focused race and browser checks.

### Slice 6: Scope-driven quiz catalog

Ticket: [TKT-20261017192913](../ticket/in_progress/20261017192913-show-quiz-cards-and-pending-scope-action.md).

Prerequisite: Slice 5 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T6.1 | Derive catalog membership and continuation from retained answers and actual sessions; implement the project-empty exception and all-answered/new-case rules. |
| T6.2 | Deliver the specified bounded three-track catalog and compact pending menu; remove detached slug forms and link selection to the focused retained quiz flow. |

Validation: Cover empty, partial, answered, cancelled and completed-empty sessions, adding a case and all-answered action removal; capture closed/open menu, long content and narrow views at declared viewport/zoom.

### Slice 7: Individual spec authoring

Ticket: [TKT-20261017194552](../ticket/in_progress/20261017194552-present-spec-cards-and-user-selected-creation-state.md).

Prerequisite: Slice 6 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T7.1 | Add shared manual create/save/state operations, native approval-alias interpretation and revision-safe metadata/provenance preservation without mandatory model access. |
| T7.2 | Add actual-document listing alongside latest-in-scope generation lookup and verify project/general/scoped context assembly without unrelated case material. |
| T7.3 | Deliver title/opening-text cards and one minimal New/Edit modal with named scope, one Body editor, compact Draft switch and Create/Save. |

Validation: Create multiple specs in one scope, one draft and one directly approved document, edit both manual/generated identities, verify stale writes and selected lifecycle after reload, and capture catalog/modal views.

### Slice 8: Contextual safe plan generation

Ticket: [TKT-20261017195741](../ticket/in_progress/20261017195741-contextualize-plan-generation-and-check-source-readiness.md).

Prerequisite: Slice 7 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T8.1 | Build immutable scoped source observations and shared absent/insufficient/sufficient readiness decisions through the configured structured generation boundary. |
| T8.2 | Stage and validate output, recheck observed inputs and target revision under the shared owner, and publish only after success without preemptive destructive archiving. |
| T8.3 | Expose contextual Generate or Regenerate in the selected case and remove global slug forms, keeping spec generation and execution approval distinct. |

Validation: Cover no discovery, irrelevant/placeholder data, insufficient context, complete-dump-only, narrow short dump and quiz-only cases; model failure, cancellation and stale source/target must preserve the prior plan.

### Slice 9: Revision-bound order-only trackers

Ticket: [TKT-20261017200232](../ticket/in_progress/20261017200232-associate-ticket-tracking-with-each-plan-revision.md).

Prerequisite: Slice 8 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T9.1 | Add revision-bound retained companions and recoverable plan/map publication, preserving historical revisions even when generated body bytes match. |
| T9.2 | Introduce order-only structural hashing and compatible versioned rendering/freshness with no status, phase, claim or receipt dependencies; preserve observed-output conflicts. |
| T9.3 | Bind approved materialization to canonical ticket references and guarded amendments, supply the map to agent context, and align the accepted tracker contract and owning tests. |

Validation: Verify plan/map interruption and retry, no allocation before approval, stable slice/task ordering, edited output, active-work amendment guards, historical bindings and byte-identical current tracker through ticket progress; focused race checks.

### Slice 10: Read-only Plans catalog

Ticket: [TKT-20261017195458](../ticket/in_progress/20261017195458-present-plans-and-tracking-as-read-only-cards.md).

Prerequisite: Slice 9 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T10.1 | Use concrete plan-document reads and the spec card anatomy; remove artifact counters, placeholder cards, frontmatter/path noise and manual plan editing. |
| T10.2 | Remove tracker cards/detail/editor/Create Tracking exposure, including direct generic-artifact URLs, while preserving internal files and contextual plan actions. |

Validation: Verify short/long plan previews, exact full document selection, no body write controls or reads that mutate data, and unavailable direct tracker views; capture catalog and read-only detail screenshots.

### Slice 11: Live work Dash

Ticket: [TKT-20261017202340](../ticket/in_progress/20261017202340-implement-live-agent-work-dash.md).

Prerequisite: Slice 10 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T11.1 | Build shared authority-aware queue/readiness and invocation/attempt/ticket/event/result observations, keeping concurrent contexts correctly associated. |
| T11.2 | Render queue/current/recent composition, quiet validation and retained completed results without requiring a latest-phase event or hiding monitoring behind an action overlay. |
| T11.3 | Add revision-aware bounded refresh, event deduplication and stale/recovery behavior while preserving scroll/selection and the targets/guards of retained controls. |

Validation: Cover queued, quiet, blocked, failed, stopped and completed work, two active contexts, no phase event, temporary read outage/reopen/restart and mutation-free polling; focused concurrency and desktop/narrow browser evidence.

### Slice 12: Consolidated usability review

Ticket: [TKT-20261017105432](../ticket/in_progress/20261017105432-review-draft-web-ui-usability.md).

Prerequisite: Slice 11 integrated; execution authority for this set.

| Task | Ordered work |
| --- | --- |
| T12.1 | Run bounded deterministic cross-feature journeys covering context capture, quiz, manual/generated spec edits, contextual planning/order binding, Board edits and Dash observation. |
| T12.2 | Review visual hierarchy, density, names, error/empty/loading states, focus/keyboard behavior, narrow widths and dirty forms across the changed surfaces; correct only in-scope consolidated gaps. |
| T12.3 | Complete the source-ticket acceptance matrix, user-facing documentation and final exact-candidate development evidence; retain the separate scheduled-nightly evidence boundary. |

Validation: Map UI-01 through UI-09 to child evidence and screenshots, verify all twelve source tickets, execute make verify for the final integrated candidate, and report scheduled exhaustive evidence separately without starting live work.

## Candidate and Closure Rules

Use the plan's development gate and source-acceptance mapping. Before a slice
PR, collect successful `make verify`, relevant focused checks and
`git diff --check` for its exact candidate. Persist the actual commands, commit
references and report evidence in the owning ticket; do not rewrite this map
because a task starts, passes a check or integrates.

The final review verifies all twelve source tickets and UI-01 through UI-09.
Required scheduled exhaustive evidence remains separate. Looper runs this approved set under its own new configuration. Preserve the
existing Telegram run, controller budgets and any independent validation.
