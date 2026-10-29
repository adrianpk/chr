# Web and TUI parity tracker

Status: In progress
Date: 2026-10-29
Delivery set: web-tui-parity
Approved: maintainer request, 2026-10-29; convert the supplied Web/TUI parity material to specification, slices, tickets and Looper execution within the recorded scope.
Implementation authority: execute these 14 slices in order through one Looper run.
Merge authority: Looper may merge only these 14 recorded slice PRs into dev using Rebase + Fast-forward after required exact-candidate checks pass.
Baseline: `4e2a556cbc34129f7ab91f76252bcb7ae80315ca` on dev.
Active slice: Slice 12 — Wizard Quiz and preparation.
Completed slices: 11/14
Completed tasks: 55/70

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

Ticket: `TKT-20261029101139` — `ops/default/ticket/solved/20261029101139-web-tui-parity-dash.md`
State: delivered
PR: [#458](https://forge.adrianpk.com/adrianpk/chr/pulls/458)
Report: `ops/default/report/slices/web-tui-parity/slice-4-dash.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T4.1 | Project ObserveWork into a terminal-neutral queue, counts, readiness reasons, exclusions and pagination. | done | 0910c75b | Native/legacy adapter observations agree with Web on eligibility, reasons, exclusions and empty pages; priority/order precede bounded paging. |
| T4.2 | Keep Dash queue selection and Board selection separately by key across refresh, sorting and removal. | done | 863d09bd; 63f66c32; e242a9a6 | Reorder/removal, stale-query and full 130-ticket Board catalog regressions pass; selection remains independent and never starts work. |
| T4.3 | Render current work, recent activity and results in Dash; preserve filter, order, Run all ready and running-work details. | done | fed68882 | Bounded context/event/result rendering and actual production-loader PTY/SQLite observation pass without canonical writes. |
| T4.4 | Expose Preview Work and Workflow details with exact scope/policy confirmation and cancellation. | done | f265da0e; e242a9a6 | Read-only Preview/Workflow PTY, cancellation, late replies, exact scope/policy conflicts and guarded host-checkpoint regressions pass. |
| T4.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | 65032c6b; 78f08a32; c4451b76 | PR #458 integrated at c4451b766e3a5a92e3b9be4a4dfc1a952401b00e; all four ordered checks passed in MQWCF6KW2NQXNMRL3FEXHGLGIH. |


### Slice 5: Case and Dump forms

Ticket: `TKT-20261029101140` — `ops/default/ticket/solved/20261029101140-web-tui-parity-case-dump.md`
State: delivered
PR: [#459](https://forge.adrianpk.com/adrianpk/chr/pulls/459)
Report: `ops/default/report/slices/web-tui-parity/slice-5-case-dump.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T5.1 | Create cases through Name, Description and Context before one Create; retain existing metadata editing and unresolved references. | done | 8a569106 | Real metadata validation and production PTY retain Description/Context after missing Name; confirmed cancellation leaves canonical files unchanged. |
| T5.2 | Create and edit dumps through scope, title and body using existing canonical operations. | done | d6f933ab | Project/case capture, selected-result editing and production PTY preserve entered body and retained title/scope. |
| T5.3 | Expose AppendDump with free text and make external-editor shortcuts and displayed hints agree in textarea mode. | done | cf8c5e4b | Captured-target append and private external-editor IO preserve user text, revision, error/size limits and late-reply ownership. |
| T5.4 | Exercise stale-save, confirmed reload/discard and unsent draft retention with owned Web and TUI fixtures. | done | fbd0b6d5 | Owned PTY concurrent Save rejects stale input; cancelled reload retains the draft and canonical bytes, confirmed reload permits the current revision. Web HTTP references pass. |
| T5.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | bb8eef11; 1bc34eda | PR #459 integrated at 1bc34edaac9b751810f776efd5ea89535908760f; all four ordered checks passed in YV7TF3QYUUICJ24UVUIWR7TTC3. |

### Slice 6: Spec catalog and manual editing

Ticket: `TKT-20261029101141` — `ops/default/ticket/solved/20261029101141-web-tui-parity-specs.md`
State: delivered
PR: [#460](https://forge.adrianpk.com/adrianpk/chr/pulls/460)
Report: `ops/default/report/slices/web-tui-parity/slice-6-specs.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T6.1 | Read the same canonical Spec catalog as Web, including manual, generated and project-wide records. | done | b4cbad75 | Canonical manual/generated/project-wide IO and Web HTTP catalog agreement pass; refresh preserves the selected concrete path. |
| T6.2 | Expose shared terminal forms for title, create-time case binding, body and draft state. | done | 0e2d48d2 | Project/case forms, actual title/body/state saves, validation retention and cancellation pass through canonical owners without generation. |
| T6.3 | Apply canonical revision checks, field validation and confirmed reload while keeping unsent text. | done | ab927ea4 | Production PTY concurrent-save conflict, cancelled/confirmed reload and fresh Save pass; captured retries, late replies and immutable case guards pass. |
| T6.4 | Retain generation, acceptance-feature generation and document inspection and prepare the same form for wizard mode. | done | df73bed5 | Palette keys reuse the direct Spec form; generation, feature generation, inspection and complete catalog frames pass. Unchanged committed-result retry creates no duplicate. |
| T6.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | 7b7718f6; 81e4331d | PR #460 integrated at 81e4331da3a711dd3dd8f1fb5e2bbd93083b85c3; all four ordered checks passed in BZNEGXNC3AUIASKTADDVRE26ZI. |

### Slice 7: Complete Quiz interaction

Ticket: `TKT-20261029101142` — `ops/default/ticket/solved/20261029101142-web-tui-parity-quizzes.md`
State: delivered
PR: [#461](https://forge.adrianpk.com/adrianpk/chr/pulls/461)
Report: `ops/default/report/slices/web-tui-parity/slice-7-quizzes.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T7.1 | Make Deep/Fast selection visible before start and mark targets with pending answers. | done | 6d12928d | Reproduced forced Deep mode; real adapter and production PTY now submit the visible Fast choice. Pending targets and scoped selection are visible. |
| T7.2 | Show answer history alongside the current question and retain answer, skip, done, cancel and resume. | done | 552751e3; 4fa89c65; 8958f020 | Neutral history agrees with Web ordering; real saved IO, history paging/resize, Cancel, Resume, Skip and Done pass. Narrow final-page and target-offset corrections preserve complete answers. |
| T7.3 | Bind each submission to its exact session, current question, revision and input digest; retain draft after errors. | done | 15fbe9b9 | Reproduced missing absent-session revision. Captured binding, queued stale answer, duplicate reconciliation, retained prompt/input, confirmed reload and actual concurrent-writer PTY pass. |
| T7.4 | Replace the always-failing context-aware quiz entry with the existing supported operation and real scope input, or explicit pre-action unavailability if unsupported. | done | 8466be11 | The contextual-authoring entry is explicitly unavailable before confirmation/executor calls. Real palette PTY compares unchanged canonical files; supported scoped runtime questions remain available. |
| T7.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | 3a6b0165; abc766b3 | PR #461 integrated at abc766b3e15ed14cd16d215ac823d1a16d25e582; all four ordered checks passed in JRP44LHIFOTCAKMSHR3ELP542E. |

### Slice 8: Plan and ticket preparation

Ticket: `TKT-20261029101143` — `ops/default/ticket/solved/20261029101143-web-tui-parity-plans.md`
State: delivered
PR: [#462](https://forge.adrianpk.com/adrianpk/chr/pulls/462)
Report: `ops/default/report/slices/web-tui-parity/slice-8-plans.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T8.1 | Route manual Generate Plan through GeneratePlanDelivery, matching the Web plan surface. | done | 446998e8 | Real adapter/HTTP regressions retain the plan and native ticket through GeneratePlanDelivery; captured case, revision and operation ID reach the shared owner. |
| T8.2 | Reconcile a retained partial preparation by its original identity rather than regenerating plans or duplicate tickets. | done | 0f98d041 | Provider failure, original retry, duplicate replay and reopened preparation retain one plan and ticket set; queued commands keep their observed target. |
| T8.3 | Preserve explicit Regenerate confirmation and any independent tracking CLI capability without exposing a misleading second required step. | done | 77f84ecb | Explicit Regenerate retains confirmation and source/revision guards; canonical catalog preference and independent tracking service regressions pass. |
| T8.4 | Offer Open document and View tickets after successful Spec/Plan generation with context-preserving navigation. | done | 1523b132 | Production PTY and IO checks open actual documents and a populated case Board without writes; cancelled replacement and idle native task assertions pass. |
| T8.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | 72bae04e; 6a59cd2d; 59108dea | PR #462 integrated at 59108deadad1848bc0ffbfda9def9fb97ea81f52; all four ordered checks passed in 32RLE7DHDBWJYZRIOXRECPQMSH. |


### Slice 9: Ticket editing, comments and references

Ticket: `TKT-20261029101144` — `ops/default/ticket/solved/20261029101144-web-tui-parity-tickets.md`
State: delivered
PR: [#463](https://forge.adrianpk.com/adrianpk/chr/pulls/463)
Report: `ops/default/report/slices/web-tui-parity/slice-9-tickets.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T9.1 | Share neutral canonical ticket metadata options and read-only rules with the existing application owner; map Web and TUI separately. | done | fc8cd8e0 | Neutral owner and separate Web/TUI mappings pass real editor-option and Web regressions; no service imports a UI. |
| T9.2 | Create and edit tickets with user title, description, case and metadata; show phase/slice and read-only explanations. | done | f2ddfdfd | Legacy/native forms round-trip entered title/body/metadata; cancel writes nothing, stale input remains, confirmed reload and managed restrictions pass. |
| T9.3 | Expose free-text comment input and retained history instead of the fixed action comment. | done | 29095bef | Distinct comments and retained history round-trip on both authorities; empty/stale input and comment-mode reload checks pass. |
| T9.4 | Expose AttachReviewRef with user input and preserve external editing, checklist display and canonical review guards. | done | ad327372 | Entered references, HTTP validation, stale/active-claim/receipt guards, native reference retention, source editing, checklist and dump-append regressions pass; production PTY journey passes. |
| T9.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | 52a38b57; be5cfd9a | PR #463 merged at be5cfd9ae8a1fa2899517feddcf1b87dc32ecea6; validation RNRFX3HTEIWFXQJHIKQSD4HET4 passed all five ordered checks. |


### Slice 10: Responsive Board and retained workflows

Ticket: `TKT-20261029101145` — `ops/default/ticket/solved/20261029101145-web-tui-parity-board-workflows.md`
State: delivered
PR: [#464](https://forge.adrianpk.com/adrianpk/chr/pulls/464)
Report: `ops/default/report/slices/web-tui-parity/slice-10-board-workflows.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T10.1 | Show all Board columns and Attention by default when width permits; retain a usable narrow view and maximize. | done | 2e6f8938 | Focused TUI resize/focus/filter tests and strict lint passed. |
| T10.2 | Add a retained workflow list/detail surface within Dash with pagination, members, evidence and current eligible actions. | done | 8d348e4d | Real empty/17-run paging, retained members and late refresh checks passed. |
| T10.3 | Expose bounded work preview for ticket, slice, set, case and loose scope with canonical filters and ordering. | done | 67456536 | Real bounded scope form/preview cancellation and canonical stale anchor checks passed. |
| T10.4 | Expose allowed lifecycle controls, allowance extension and integration-decision answers, plus visible case/ticket links. | done | 6f391411 | Real grant/decision/native approval guards, retained drafts/links and production PTY stop controls passed. |
| T10.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | 600fe4b1; 728f5dd5 | PR #464 merged at 728f5dd5bcc59f172ba63b9e62e386a084f86c2d; validation JQCPMJ55MJPOB64VPBLMV6N4Q3 passed all four ordered checks. |


### Slice 11: Wizard entry and Dump in shared tabs

Ticket: `TKT-20261029101146` — `ops/default/ticket/solved/20261029101146-web-tui-parity-wizard-entry.md`
State: delivered
PR: [#465](https://forge.adrianpk.com/adrianpk/chr/pulls/465)
Report: `ops/default/report/slices/web-tui-parity/slice-11-wizard-entry.md`

| Task | Work | State | Commits | Validation |
| --- | --- | --- | --- | --- |
| T11.1 | Expose Wizard and Open workflow in Cases and start with one Name/Create through StartNamedGuidedUseCase. | done | d0516dc7 | Named Create, replay and case-bound re-entry pass through the real owner and production terminal. |
| T11.2 | Replace the detached wizard presentation with case/run binding in the existing shell, stage indicator and shared forms. | done | a1ed8105 | Existing tab, progress, captured input/retry and committed-original inspection checks pass. |
| T11.3 | Use the existing Dump form with a single Continue to quiz submission of the current content. | done | 7524663d | Shared Dump Continue seals exact current content once; real lost-ack replay and production PTY assertions pass. |
| T11.4 | Keep non-current wizard tabs read-only with Continue to the current stage; isolate manual and wizard controls. | done | 06d019e5 | Non-current tabs and manual writers cannot mutate; navigation, provenance/byte identity and native parity checks pass. |
| T11.5 | Validate, report, open the recorded PR, verify dev integration and close this slice. | done | f51e7e39; 6ee42aac | PR #465 merged at 6ee42aac701f6b2e6f479cbd639601965b29602b; validation COHC2KP7RZY5ONJTIC4UAQFWZJ passed all four ordered checks. |


### Slice 12: Wizard Quiz and preparation

Ticket: `TKT-20261029101147` — `ops/default/ticket/in_progress/20261029101147-web-tui-parity-wizard-authoring.md`
State: in_progress
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
