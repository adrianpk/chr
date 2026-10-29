# Web and TUI parity tracker

Status: In progress
Date: 2026-10-29
Delivery set: web-tui-parity
Approved: maintainer request, 2026-10-29; convert the supplied Web/TUI parity material to specification, slices, tickets and Looper execution within the recorded scope.
Implementation authority: execute these 14 slices in order through one Looper run.
Merge authority: Looper may merge only these 14 recorded slice PRs into dev using Rebase + Fast-forward after required exact-candidate checks pass.
Baseline: `4e2a556cbc34129f7ab91f76252bcb7ae80315ca` on dev.
Active slice: Slice 4 — Shared queue observation and stable refresh.
Completed slices: 3/14
Completed tasks: 19/70

Plan: [Delivery contract](../plan/web-tui-parity.md)
Spec: [Requirements](../spec/web-tui-parity.md)
Coverage: [All source IDs](../plan/web-tui-parity-coverage.csv)

## Canonical slice map

| Slice | Short name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Safe queue selection and canonical moves | fix/slice-1-tui-safe-actions | fix(slice-1): bind TUI actions to visible targets | ops/default/report/slices/web-tui-parity/slice-1-safe-actions.md |
| Slice 2 | Case-oriented shell and contextual help | feat/slice-2-tui-navigation | feat(slice-2): align TUI navigation with case context | ops/default/report/slices/web-tui-parity/slice-2-navigation.md |
| Slice 3 | Shared operation feedback and draft retention | feat/slice-3-tui-feedback | feat(slice-3): show actionable TUI operation results | ops/default/report/slices/web-tui-parity/slice-3-feedback.md |
| Slice 4 | Shared queue observation and stable refresh | feat/slice-4-tui-dash | feat(slice-4): share TUI queue readiness and activity | ops/default/report/slices/web-tui-parity/slice-4-dash.md |
| Slice 5 | Case and Dump forms | feat/slice-5-tui-case-dump | feat(slice-5): add complete TUI case and dump forms | ops/default/report/slices/web-tui-parity/slice-5-case-dump.md |
| Slice 6 | Spec catalog and manual editing | feat/slice-6-tui-specs | feat(slice-6): support manual TUI specifications | ops/default/report/slices/web-tui-parity/slice-6-specs.md |
| Slice 7 | Complete Quiz interaction | feat/slice-7-tui-quizzes | feat(slice-7): complete TUI quiz controls and history | ops/default/report/slices/web-tui-parity/slice-7-quizzes.md |
| Slice 8 | Plan and ticket preparation | feat/slice-8-tui-plans | feat(slice-8): prepare TUI plans and tickets together | ops/default/report/slices/web-tui-parity/slice-8-plans.md |
| Slice 9 | Ticket editing, comments and references | feat/slice-9-tui-tickets | feat(slice-9): add complete TUI ticket interactions | ops/default/report/slices/web-tui-parity/slice-9-tickets.md |
| Slice 10 | Responsive Board and retained workflows | feat/slice-10-tui-board-workflows | feat(slice-10): expose TUI board and retained workflows | ops/default/report/slices/web-tui-parity/slice-10-board-workflows.md |
| Slice 11 | Wizard entry and Dump in shared tabs | feat/slice-11-tui-wizard-entry | feat(slice-11): start TUI wizards inside shared tabs | ops/default/report/slices/web-tui-parity/slice-11-wizard-entry.md |
| Slice 12 | Wizard Quiz and preparation | feat/slice-12-tui-wizard-authoring | feat(slice-12): reuse TUI quiz and preparation in wizards | ops/default/report/slices/web-tui-parity/slice-12-wizard-authoring.md |
| Slice 13 | Wizard execution, results and recovery | feat/slice-13-tui-wizard-results | feat(slice-13): complete TUI wizard delivery and recovery | ops/default/report/slices/web-tui-parity/slice-13-wizard-results.md |
| Slice 14 | Parity acceptance and capability inventory | test/slice-14-tui-acceptance | test(slice-14): verify complete Web and TUI workflow parity | ops/default/report/slices/web-tui-parity/slice-14-acceptance.md |

## Tracking contract

The map must remain identical to the plan. Tasks below are canonical. Record actual commits and exact validation outcomes; planned checks are not passing evidence. Each slice has one ticket and report; move the ticket with matching status and retain its filename. Looper owns execution and conditional PR integration into dev. Passive observations, document preparation and PR readiness do not establish delivery.

## Tasks

### Slice 1: Safe queue selection and canonical moves

Ticket: `TKT-20261029101136` — `ops/default/ticket/solved/20261029101136-web-tui-parity-safe-actions.md`
State: delivered
PR: [#455](https://forge.adrianpk.com/adrianpk/chr/pulls/455)
Report: `ops/default/report/slices/web-tui-parity/slice-1-safe-actions.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T1.1 | Reproduce Dash/Board target divergence, legacy review movement and busy-key swallowing through production adapters. | done | 61a30044; 450d25ab; 3b8baf49 | Reproduced wrong Dash target, accepted legacy review move and swallowed busy Ctrl+C before fixes; named regressions now pass. |
| T1.2 | Give Dash an independent stable selected key and bind preview/start to that key and the observed revision. | done | 61a30044 | Dash key/navigation/refresh tests and real native adapter preview/stale-confirmation regressions pass. |
| T1.3 | Expose the six canonical Board targets through the existing guarded ticket move operation; preserve managed-ticket and review-evidence restrictions. | done | 450d25ab | Web/TUI move matrix, stale document and active-owner regressions pass; all six palette destinations are canonical. |
| T1.4 | Route Ctrl+C through graceful UI shutdown during busy operations, preserving retained-work ownership and cancellation semantics. | done | 3b8baf49 | Busy-mode Ctrl+C tests and production PTY/provider shutdown pass; retained workflow inspection is unchanged. |
| T1.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | b97e2363; c8034b74; 16162289 | PR #455 merged into dev at 161622891a376efe97e3e2d0b3bbebaaf8235660; all four ordered checks passed in validation YCBB7LWNNQR5OLKQDKXKLQ62HN. |

### Slice 2: Case-oriented shell and contextual help

Ticket: `TKT-20261029101137` — `ops/default/ticket/solved/20261029101137-web-tui-parity-navigation.md`
State: delivered
PR: [#456](https://forge.adrianpk.com/adrianpk/chr/pulls/456)
Report: `ops/default/report/slices/web-tui-parity/slice-2-navigation.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T2.1 | Align tabs and initial view by View ID: Cases, Dumps, Quizzes, Specs, Plans, Board, Dash; F1-F7 and F10 Exit. | done | f2d7961b; 5dbd0c2b | Seven View IDs own startup, F1-F7 and actual footer clicks; focused shell and full catalog regressions pass. |
| T2.2 | Keep Project in header context and an alt+p action; expose real ? Help outside text modes and retain palette access. | done | f2d7961b; 5dbd0c2b | Contextual Help returns to the same pane; text, function-key and mouse guards retain drafts; production Harness read/save/refresh tests pass through alt+p. |
| T2.3 | Propagate functional case/run binding between tabs and show active-view actions in the real shell footer. | done | f2d7961b; 5dbd0c2b | Case-scoped artifact and Board selection, retained-run binding, missing-case recovery and passive draft preservation regressions pass. |
| T2.4 | Inventory useful legacy views and retain workspace-index health, runtime information and existing commands through Project, Dash or palette. | done | f2d7961b; 5dbd0c2b | Useful-capability inventory is in the Slice 2 report; production loader and PTY verify auxiliary access and unchanged canonical documents. |
| T2.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | 5dbd0c2b; b797ef9c; 61e2398c; 322e8f0a | PR #456 integrated into dev at 322e8f0a7a7b17536f1cbe72b72e8023d2a4efb0; all four ordered checks passed in HCHOQPMEAG7XZXQ6XGFNLY3BFU. |


### Slice 3: Shared operation feedback and draft retention

Ticket: `TKT-20261029101138` — `ops/default/ticket/solved/20261029101138-web-tui-parity-feedback.md`
State: delivered
PR: [#457](https://forge.adrianpk.com/adrianpk/chr/pulls/457)
Report: `ops/default/report/slices/web-tui-parity/slice-3-feedback.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T3.1 | Reuse neutral application outcomes and map ok, validation_failed, stale and blocked into the existing Action overlay. | done | eb484dd3 | Shared application outcomes preserve Web status strings and typed errors; neutral classification and adapter mapping regressions pass. |
| T3.2 | Show summary, field findings, changed files and eligible reload/retry action with existing TUI styling. | done | 8046584d; bc8e124d | Action panel renders distinct outcomes, findings, changed sources and current eligible recovery; long findings scroll. |
| T3.3 | Keep user input on validation/stale/error and require confirmation before destructive draft reload or discard. | done | 8046584d; bc8e124d | Sentinel and production IO/PTY regressions retain failed input; reload/discard require confirmation and canonical sources remain unchanged on cancellation. |
| T3.4 | Bind pending results to captured operation, selection and revision; ignore late results for a newer interaction. | done | 8046584d; bc8e124d | Queued/reordered replies, editor generations, original retries and native revision regressions preserve current drafts and confirmations. |
| T3.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | eb2a245a; 2214eecf; c91ba4fe | PR #457 integrated into dev at c91ba4feac5b4296548a854e26b8d9cb59170ed7; all four ordered checks passed in 7IDEK3RU26UI7CUQSTNJ7T3NAK. |


### Slice 4: Shared queue observation and stable refresh

Ticket: `TKT-20261029101139` — `ops/default/ticket/reviewing/20261029101139-web-tui-parity-dash.md`
State: reviewing
PR: [#458](https://forge.adrianpk.com/adrianpk/chr/pulls/458)
Report: `ops/default/report/slices/web-tui-parity/slice-4-dash.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T4.1 | Project ObserveWork into a terminal-neutral queue, counts, readiness reasons, exclusions and pagination. | done | 0910c75b | Native/legacy adapter observations agree with Web on eligibility, reasons, exclusions and empty pages; priority/order precede bounded paging. |
| T4.2 | Keep Dash queue selection and Board selection separately by key across refresh, sorting and removal. | done | 863d09bd; 63f66c32 | Reorder/removal, stale-query and full 130-ticket Board catalog regressions pass; selection remains independent and never starts work. |
| T4.3 | Render current work, recent activity and results in Dash; preserve filter, order, Run all ready and running-work details. | done | fed68882 | Bounded context/event/result rendering and actual production-loader PTY/SQLite observation pass without canonical writes. |
| T4.4 | Expose Preview Work and Workflow details with exact scope/policy confirmation and cancellation. | done | f265da0e | Read-only Preview/Workflow PTY, cancellation, late replies, exact scope/policy conflicts and guarded host-checkpoint regressions pass. |
| T4.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | in_progress | 65032c6b | PR #458 prepared with the immutable report; ordered exact-HEAD checks are retained by Looper before controller-owned integration. |


### Slice 5: Case and Dump forms

Ticket: `TKT-20261029101140` — `ops/default/ticket/ready/20261029101140-web-tui-parity-case-dump.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-5-case-dump.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T5.1 | Create cases through Name, Description and Context before one Create; retain existing metadata editing and unresolved references. | pending | — | Planned: Cancel creates no blank case; entered fields survive validation failure. |
| T5.2 | Create and edit dumps through scope, title and body using existing canonical operations. | pending | — | Planned: Project-wide and case-bound dumps preserve actual user content without fixed placeholder text. |
| T5.3 | Expose AppendDump with free text and make external-editor shortcuts and displayed hints agree in textarea mode. | pending | — | Planned: The reachable append action uses the selected dump; editor round trips preserve edits. |
| T5.4 | Exercise stale-save, confirmed reload/discard and unsent draft retention with owned Web and TUI fixtures. | pending | — | Planned: Runtime disposition and evidence for G-018 are recorded; no lost draft or silent overwrite. |
| T5.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 6: Spec catalog and manual editing

Ticket: `TKT-20261029101141` — `ops/default/ticket/ready/20261029101141-web-tui-parity-specs.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-6-specs.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T6.1 | Read the same canonical Spec catalog as Web, including manual, generated and project-wide records. | pending | — | Planned: Owned fixture proves all three record kinds appear, with stable functional selection. |
| T6.2 | Expose shared terminal forms for title, create-time case binding, body and draft state. | pending | — | Planned: Create, edit and cancel use production service operations; no generated-only assumptions. |
| T6.3 | Apply canonical revision checks, field validation and confirmed reload while keeping unsent text. | pending | — | Planned: Stale save cannot overwrite a concurrent change; case identity is not silently rebound. |
| T6.4 | Retain generation, acceptance-feature generation and document inspection and prepare the same form for wizard mode. | pending | — | Planned: Manual/wizard presentations share the editor owner rather than copying markup or update handlers. |
| T6.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 7: Complete Quiz interaction

Ticket: `TKT-20261029101142` — `ops/default/ticket/ready/20261029101142-web-tui-parity-quizzes.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-7-quizzes.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T7.1 | Make Deep/Fast selection visible before start and mark targets with pending answers. | pending | — | Planned: Mode reaches the shared service; targets are scoped to the functional case. |
| T7.2 | Show answer history alongside the current question and retain answer, skip, done, cancel and resume. | pending | — | Planned: Multiple rounds preserve history and do not lose the active question. |
| T7.3 | Bind each submission to its exact session, current question, revision and input digest; retain draft after errors. | pending | — | Planned: Stale/double submission cannot answer another question; cancel/discard has confirmation. |
| T7.4 | Replace the always-failing context-aware quiz entry with the existing supported operation and real scope input, or explicit pre-action unavailability if unsupported. | pending | — | Planned: No confirmation invites an operation known to fail; no invented generator or new domain workflow. |
| T7.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 8: Plan and ticket preparation

Ticket: `TKT-20261029101143` — `ops/default/ticket/ready/20261029101143-web-tui-parity-plans.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-8-plans.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T8.1 | Route manual Generate Plan through GeneratePlanDelivery, matching the Web plan surface. | pending | — | Planned: One action retains the plan and creates the expected native tickets; execution has not started. |
| T8.2 | Reconcile a retained partial preparation by its original identity rather than regenerating plans or duplicate tickets. | pending | — | Planned: Injected partial failure followed by retry produces exactly one intended ticket set. |
| T8.3 | Preserve explicit Regenerate confirmation and any independent tracking CLI capability without exposing a misleading second required step. | pending | — | Planned: Existing-plan handling and regeneration preserve source/revision and scope checks. |
| T8.4 | Offer Open document and View tickets after successful Spec/Plan generation with context-preserving navigation. | pending | — | Planned: Actions appear only when their document or ticket result exists. |
| T8.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 9: Ticket editing, comments and references

Ticket: `TKT-20261029101144` — `ops/default/ticket/ready/20261029101144-web-tui-parity-tickets.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-9-tickets.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T9.1 | Share neutral canonical ticket metadata options and read-only rules with the existing application owner; map Web and TUI separately. | pending | — | Planned: No TUI dependency on completeTicketEditor Web types; options and field restrictions agree. |
| T9.2 | Create and edit tickets with user title, description, case and metadata; show phase/slice and read-only explanations. | pending | — | Planned: No fixed TUI-created body/area; revision and managed-ticket restrictions remain effective. |
| T9.3 | Expose free-text comment input and retained history instead of the fixed action comment. | pending | — | Planned: Round trip multiple distinct comments and reject empty/stale submissions without losing input. |
| T9.4 | Expose AttachReviewRef with user input and preserve external editing, checklist display and canonical review guards. | pending | — | Planned: Review reference is validated and attached to the selected ticket; append wiring from Slice 5 remains reachable. |
| T9.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 10: Responsive Board and retained workflows

Ticket: `TKT-20261029101145` — `ops/default/ticket/ready/20261029101145-web-tui-parity-board-workflows.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-10-board-workflows.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T10.1 | Show all Board columns and Attention by default when width permits; retain a usable narrow view and maximize. | pending | — | Planned: Wide/narrow resize and focus preserve key selection and canonical column meaning. |
| T10.2 | Add a retained workflow list/detail surface within Dash with pagination, members, evidence and current eligible actions. | pending | — | Planned: The list handles empty/large sets and does not infer delivery from worker exit or an open PR. |
| T10.3 | Expose bounded work preview for ticket, slice, set, case and loose scope with canonical filters and ordering. | pending | — | Planned: Selection, later-ticket inclusion, policy and confirmation match the observed application snapshot. |
| T10.4 | Expose allowed lifecycle controls, allowance extension and integration-decision answers, plus visible case/ticket links. | pending | — | Planned: Stale/revision/budget/approval guards remain host-owned; unavailable controls explain why. |
| T10.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 11: Wizard entry and Dump in shared tabs

Ticket: `TKT-20261029101146` — `ops/default/ticket/ready/20261029101146-web-tui-parity-wizard-entry.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-11-wizard-entry.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T11.1 | Expose Wizard and Open workflow in Cases and start with one Name/Create through StartNamedGuidedUseCase. | pending | — | Planned: No random-ID input ceremony or second name step; existing case/run re-entry is deterministic. |
| T11.2 | Replace the detached wizard presentation with case/run binding in the existing shell, stage indicator and shared forms. | pending | — | Planned: Stages map to existing tabs; no parallel menu, duplicated editor or raw namespace/UUID prompt. |
| T11.3 | Use the existing Dump form with a single Continue to quiz submission of the current content. | pending | — | Planned: Runtime G-008 evidence proves one exact retained input and retry reconciliation without duplicates. |
| T11.4 | Keep non-current wizard tabs read-only with Continue to the current stage; isolate manual and wizard controls. | pending | — | Planned: Manual save cannot accidentally break wizard progression; viewing another tab never submits work. |
| T11.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 12: Wizard Quiz and preparation

Ticket: `TKT-20261029101147` — `ops/default/ticket/ready/20261029101147-web-tui-parity-wizard-authoring.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-12-wizard-authoring.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T12.1 | Use the same Quiz component and controls in manual and wizard mode, bound to the retained run and current question. | pending | — | Planned: Deep/Fast, answer history, skip, answer, done and cancel remain functional without duplicate form logic. |
| T12.2 | Render specification and preparation stages in their existing tabs from the shared guided operation. | pending | — | Planned: Specification auto-approval, plan/ticket preparation and committed original inputs remain domain-owned. |
| T12.3 | Show stage-specific processing, suspended, pending-operation and blocker feedback with eligible actions. | pending | — | Planned: Partial result and timeout keep the original operation identity and saved input visible. |
| T12.4 | Retry the retained operation and handle stale/newer revisions without UI-owned orchestration or duplicate preparation. | pending | — | Planned: Restart/re-entry and out-of-order responses preserve exact run state and finite attempts. |
| T12.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 13: Wizard execution, results and recovery

Ticket: `TKT-20261029101148` — `ops/default/ticket/ready/20261029101148-web-tui-parity-wizard-results.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-13-wizard-results.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T13.1 | Use the case-filtered Board for wizard execution, progress, review links and eligible lifecycle controls. | pending | — | Planned: Counts distinguish validated, confirmed and remaining work using retained receipts. |
| T13.2 | Render structured terminal results with summary, branch/PR, delivery evidence and validation commands/exit codes. | pending | — | Planned: Open PR or successful worker cannot be shown as confirmed delivery; JSON export remains available. |
| T13.3 | Confirm cancellation with the same saved-document/completed-change warning and show Pause only where domain policy permits it. | pending | — | Planned: Owned runtime evidence for G-013 proves confirmation and retained artifacts; no implicit rollback. |
| T13.4 | Keep polling read-only, preserve unsent drafts and navigation, and retain later ideas/originals as existing secondary capabilities. | pending | — | Planned: Owned runtime evidence for G-014 proves no redirect, draft replacement or hidden progression while typing. |
| T13.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

### Slice 14: Parity acceptance and capability inventory

Ticket: `TKT-20261029101149` — `ops/default/ticket/ready/20261029101149-web-tui-parity-acceptance.md`
State: ready
PR: pending
Report: `ops/default/report/slices/web-tui-parity/slice-14-acceptance.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T14.1 | Reconcile all 38 gap rows and four incidental findings with exact tests and dispositions; close runtime uncertainties without changing source claims into assumed results. | pending | — | Planned: Every coverage row links a delivered slice/report and passing evidence or an explicit maintainer-approved exception. |
| T14.2 | Reconcile the palette and legacy capability inventory; gate unsupported safe checklist writers before confirmation without inventing a new domain or filesystem writer. | pending | — | Planned: No advertised unconditional failure remains; useful baseline extras retain tested access. |
| T14.3 | Extend owned production PTY and Web reference fixtures to exercise manual forms and the complete guided journey, including focus, resize, cancel and restart. | pending | — | Planned: Real Bubble Tea process plus real HTTP/Chromium, filesystem, Git and SQLite; deterministic model/provider responses only. |
| T14.4 | Update the User Guide and consolidate the delivered user-visible changelog outcomes; run final focused acceptance and publish the cycle report. | pending | — | Planned: No claims based on screenshots or reducer-only tests; no broad suite, new dependency or unapproved live project mutation. |
| T14.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | pending | — | Planned: make verify; focused named regressions; git diff --check; canonical merged PR/hash |

## Blockers and decisions

None recorded. Five source runtime uncertainties are assigned to explicit implementation/acceptance tasks; they are not execution blockers. Any missing canonical operation or incompatible guard must be investigated in the owning slice, with out-of-scope domain changes returned for maintainer decision.

## Final evidence

Not run. No runtime implementation, integrated slice or passing acceptance is claimed at planning. Final controller gate: `make verify`. Focused acceptance selection and fixture boundaries are defined in the plan. Cycle report is created at actual closure.
