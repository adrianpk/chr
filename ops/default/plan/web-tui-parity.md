# Web and TUI parity delivery

Status: Approved for execution
Date: 2026-10-29
Delivery set: web-tui-parity
Approved: maintainer request, 2026-10-29; convert the supplied Web/TUI parity material to specification, slices, tickets and Looper execution within the recorded scope.
Implementation authority: execute these 14 slices in order through one Looper run.
Merge authority: Looper may merge only these 14 recorded slice PRs into dev using Rebase + Fast-forward after required exact-candidate checks pass.
Baseline: `4e2a556cbc34129f7ab91f76252bcb7ae80315ca` on dev.
Active slice: Slice 1 — Safe queue selection and canonical moves.

Spec: [Requirements](../spec/web-tui-parity.md)
Tracker: [Tasks and evidence](../tracker/web-tui-parity.md)
Source: [Assessed inventory](../report/web-tui-parity-discovery.md)
Coverage: [All source IDs](web-tui-parity-coverage.csv)

## Outcome and slicing strategy

Slice strategy: behavior-first.

Deliver all source outcomes in one plan, one tracker, 14 slice tickets and one Looper cycle. Safety precedes presentation, shared manual forms precede wizard reuse, and final acceptance verifies the integrated journeys. Each slice compiles, has useful behavior and focused regression evidence; no empty scaffolding or intentionally failing intermediate slice is authorized. The ordered map is fixed for this cycle; changing scope, adding slices or broadening recovery requires maintainer approval.

No model companion is needed: existing durable entities, status machines, persistence, command ownership and schemas are reused. The specification's view map and draft/operation binding rules define the transient presentation contract.

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

## Execution contract

- Commit and push this specification, assessed inventory, coverage ledger, plan, tracker and ready tickets on dev before creating a slice branch.
- Execute in order. Each slice starts from current clean dev after its predecessor is integrated, uses its exact branch/title and opens one PR to dev.
- Use one Conventional Commit per task when practical. Preserve unrelated changes and active runtime/controller processes. Central application operations own shared behavior; UI-specific DTOs stay in adapters.
- Capture immutable command inputs and typed operation outcomes; do not read mutable Model state from delayed commands. Keep revision/question/scope binding, request identities, finite budgets and host-owned integration guards.
- Update the owning ticket through ready, in_progress, reviewing and solved in the corresponding operation. Tickets mirror delivery; the tracker remains the canonical task/evidence ledger. Do not admit these repository tickets to a second chr runtime execution.
- Create the one report only when implementing its slice; do not prefill delivered reports. Use required report sections and statuses. Link its immutable introduction commit from the PR.
- PR bodies contain exactly Summary, Changes and Validation. Use the recorded Conventional title. No WIP prefix, process disclaimer or fourth section. List configured required commands first in their configured order, then focused checks and diff validation.
- Looper alone owns conditional slice integration. Verify actual canonical merge and dev hash before closing slice records. Continue the next recorded slice without a fresh authorization request.
- Freeze the external Looper configuration after the first checkpoint. Keep one run, its conversation, ownership and finite counters on recovery. Default budgets are 200 worker turns and 5 stalled turns; no unbounded loop or automatic successor is authorized.

## Validation contract

The configured local final gate is exactly `make verify`, following `.agents/LINTING.md`. Each slice requires that same development gate and `git diff --check`, plus the smallest relevant named production regression selection. Changed IO/workflow behavior must be tested without `-short`; changed asynchronous ownership/cancellation must receive the focused `-race` selection. Do not add complete normal/race suites, nightly execution or live-forge certification to this cycle.

Use the existing `internal/app` production fixtures and `internal/tui` component tests. Preserve meaningful explanatory test comments, concise names and named table cases; do not add tests that only restate a new helper. Prefer the existing owned filesystem, Git, SQLite, HTTP, PTY and Chromium boundaries. Model/provider responses may be deterministic; command selection, persistence, revisions and terminal input must be real. No live project or account mutation is required.

Final Slice 14 explicitly selects and records the existing guided/native, real terminal, and owned browser entry points, extended for the delivered behavior:

```sh
go test ./internal/app -run '^Test(GuidedUseCaseParityNative|GuidedUseCaseTUITerminal|TUIWebParity)$' -count=1 -timeout=5m
CHR_WEB_BROWSER=/usr/bin/chromium go test ./internal/app -run '^Test(GuidedUseCaseWebBrowser|WebQuizBrowser|WebSpecBrowser|WebDumpEditorBrowser|WebBoardBrowser|WebDashBrowser)$' -count=1 -timeout=5m
```

`TestTUIWebParity` is the planned named acceptance family, not a baseline test claimed to exist. Slice 14 must ensure the selected family and existing PTY entry point actually run and do not skip. The explicit Chromium binary is available at planning; verify fixture access before claiming browser acceptance. Keep tests scoped to these workflows, not an expanded certification suite. Adapt tests for the new UI interactions while preserving the old safety assertions.

Runtime uncertainties G-008, G-013, G-014, G-018 and G-024 are resolved in their owning slices and cross-checked in final acceptance. Record what was observed in Web and TUI, exact commands, pass/fail, revisions and fixture boundaries. A static finding or screenshot cannot be reported as runtime proof.

## Slice work and acceptance

### Slice 1: Safe queue selection and canonical moves

Ticket: `TKT-20261029101136` — `ops/default/ticket/ready/20261029101136-web-tui-parity-safe-actions.md`
Findings: G-026, G-034, I-001.
Prerequisite: committed planning artifacts.

- **T1.1:** Reproduce Dash/Board target divergence, legacy review movement and busy-key swallowing through production adapters.
  Acceptance: Distinct Board/queue keys; alias review target without evidence; active busy operation.
- **T1.2:** Give Dash an independent stable selected key and bind preview/start to that key and the observed revision.
  Acceptance: Arrow movement, empty queue, reorder/removal and stale confirmation cannot run another ticket.
- **T1.3:** Expose the six canonical Board targets through the existing guarded ticket move operation; preserve managed-ticket and review-evidence restrictions.
  Acceptance: Web and TUI reject identical forbidden/stale moves; successful writes persist canonical status.
- **T1.4:** Route Ctrl+C through graceful UI shutdown during busy operations, preserving retained-work ownership and cancellation semantics.
  Acceptance: Owned terminal process exits; no implicit retained-work cancellation or goroutine leak.
- **T1.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 2: Case-oriented shell and contextual help

Ticket: `TKT-20261029101137` — `ops/default/ticket/ready/20261029101137-web-tui-parity-navigation.md`
Findings: G-001, G-002, G-003, G-033, I-002.
Prerequisite: Slice 1 integrated into dev.

- **T2.1:** Align tabs and initial view by View ID: Cases, Dumps, Quizzes, Specs, Plans, Board, Dash; F1-F7 and F10 Exit.
  Acceptance: Navigation, click targets and restored selection survive the reordered view list.
- **T2.2:** Keep Project in header context and an alt+p action; expose real ? Help outside text modes and retain palette access.
  Acceptance: No text-field key interception; project editing and Harness selection remain accessible.
- **T2.3:** Propagate functional case/run binding between tabs and show active-view actions in the real shell footer.
  Acceptance: Case context survives navigation without submitting or mutating documents.
- **T2.4:** Inventory useful legacy views and retain workspace-index health, runtime information and existing commands through Project, Dash or palette.
  Acceptance: Every baseline extra has a reachable destination; no duplicate legacy tab bar.
- **T2.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 3: Shared operation feedback and draft retention

Ticket: `TKT-20261029101138` — `ops/default/ticket/ready/20261029101138-web-tui-parity-feedback.md`
Findings: G-038.
Prerequisite: Slice 2 integrated into dev.

- **T3.1:** Reuse neutral application outcomes and map ok, validation_failed, stale and blocked into the existing Action overlay.
  Acceptance: No TUI import of Web DTOs and no domain import of terminal types.
- **T3.2:** Show summary, field findings, changed files and eligible reload/retry action with existing TUI styling.
  Acceptance: Success, validation rejection, stale revision and retained pending operation remain distinguishable.
- **T3.3:** Keep user input on validation/stale/error and require confirmation before destructive draft reload or discard.
  Acceptance: Sentinel unsent text survives errors, focus changes and delayed responses.
- **T3.4:** Bind pending results to captured operation, selection and revision; ignore late results for a newer interaction.
  Acceptance: Out-of-order completion cannot overwrite another form or apply a stale confirmation.
- **T3.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 4: Shared queue observation and stable refresh

Ticket: `TKT-20261029101139` — `ops/default/ticket/ready/20261029101139-web-tui-parity-dash.md`
Findings: G-032, G-035, G-036, G-037.
Prerequisite: Slice 3 integrated into dev.

- **T4.1:** Project ObserveWork into a terminal-neutral queue, counts, readiness reasons, exclusions and pagination.
  Acceptance: Compare the same application observation across adapters, including ineligible and empty pages.
- **T4.2:** Keep Dash queue selection and Board selection separately by key across refresh, sorting and removal.
  Acceptance: Selection follows the same key or a deterministic visible fallback without starting work.
- **T4.3:** Render current work, recent activity and results in Dash; preserve filter, order, Run all ready and running-work details.
  Acceptance: Existing queue extras remain usable; large/empty event collections are bounded.
- **T4.4:** Expose Preview Work and Workflow details with exact scope/policy confirmation and cancellation.
  Acceptance: Changed snapshot or scope invalidates confirmation; no execution from passive refresh.
- **T4.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 5: Case and Dump forms

Ticket: `TKT-20261029101140` — `ops/default/ticket/ready/20261029101140-web-tui-parity-case-dump.md`
Findings: G-015, G-017, G-018, G-019.
Prerequisite: Slice 4 integrated into dev.

- **T5.1:** Create cases through Name, Description and Context before one Create; retain existing metadata editing and unresolved references.
  Acceptance: Cancel creates no blank case; entered fields survive validation failure.
- **T5.2:** Create and edit dumps through scope, title and body using existing canonical operations.
  Acceptance: Project-wide and case-bound dumps preserve actual user content without fixed placeholder text.
- **T5.3:** Expose AppendDump with free text and make external-editor shortcuts and displayed hints agree in textarea mode.
  Acceptance: The reachable append action uses the selected dump; editor round trips preserve edits.
- **T5.4:** Exercise stale-save, confirmed reload/discard and unsent draft retention with owned Web and TUI fixtures.
  Acceptance: Runtime disposition and evidence for G-018 are recorded; no lost draft or silent overwrite.
- **T5.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 6: Spec catalog and manual editing

Ticket: `TKT-20261029101141` — `ops/default/ticket/ready/20261029101141-web-tui-parity-specs.md`
Findings: G-023, G-024.
Prerequisite: Slice 5 integrated into dev.

- **T6.1:** Read the same canonical Spec catalog as Web, including manual, generated and project-wide records.
  Acceptance: Owned fixture proves all three record kinds appear, with stable functional selection.
- **T6.2:** Expose shared terminal forms for title, create-time case binding, body and draft state.
  Acceptance: Create, edit and cancel use production service operations; no generated-only assumptions.
- **T6.3:** Apply canonical revision checks, field validation and confirmed reload while keeping unsent text.
  Acceptance: Stale save cannot overwrite a concurrent change; case identity is not silently rebound.
- **T6.4:** Retain generation, acceptance-feature generation and document inspection and prepare the same form for wizard mode.
  Acceptance: Manual/wizard presentations share the editor owner rather than copying markup or update handlers.
- **T6.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 7: Complete Quiz interaction

Ticket: `TKT-20261029101142` — `ops/default/ticket/ready/20261029101142-web-tui-parity-quizzes.md`
Findings: G-020, G-021, G-022.
Prerequisite: Slice 6 integrated into dev.

- **T7.1:** Make Deep/Fast selection visible before start and mark targets with pending answers.
  Acceptance: Mode reaches the shared service; targets are scoped to the functional case.
- **T7.2:** Show answer history alongside the current question and retain answer, skip, done, cancel and resume.
  Acceptance: Multiple rounds preserve history and do not lose the active question.
- **T7.3:** Bind each submission to its exact session, current question, revision and input digest; retain draft after errors.
  Acceptance: Stale/double submission cannot answer another question; cancel/discard has confirmation.
- **T7.4:** Replace the always-failing context-aware quiz entry with the existing supported operation and real scope input, or explicit pre-action unavailability if unsupported.
  Acceptance: No confirmation invites an operation known to fail; no invented generator or new domain workflow.
- **T7.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 8: Plan and ticket preparation

Ticket: `TKT-20261029101143` — `ops/default/ticket/ready/20261029101143-web-tui-parity-plans.md`
Findings: G-004, G-025.
Prerequisite: Slice 7 integrated into dev.

- **T8.1:** Route manual Generate Plan through GeneratePlanDelivery, matching the Web plan surface.
  Acceptance: One action retains the plan and creates the expected native tickets; execution has not started.
- **T8.2:** Reconcile a retained partial preparation by its original identity rather than regenerating plans or duplicate tickets.
  Acceptance: Injected partial failure followed by retry produces exactly one intended ticket set.
- **T8.3:** Preserve explicit Regenerate confirmation and any independent tracking CLI capability without exposing a misleading second required step.
  Acceptance: Existing-plan handling and regeneration preserve source/revision and scope checks.
- **T8.4:** Offer Open document and View tickets after successful Spec/Plan generation with context-preserving navigation.
  Acceptance: Actions appear only when their document or ticket result exists.
- **T8.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 9: Ticket editing, comments and references

Ticket: `TKT-20261029101144` — `ops/default/ticket/ready/20261029101144-web-tui-parity-tickets.md`
Findings: G-027, G-029, G-030, I-004.
Prerequisite: Slice 8 integrated into dev.

- **T9.1:** Share neutral canonical ticket metadata options and read-only rules with the existing application owner; map Web and TUI separately.
  Acceptance: No TUI dependency on completeTicketEditor Web types; options and field restrictions agree.
- **T9.2:** Create and edit tickets with user title, description, case and metadata; show phase/slice and read-only explanations.
  Acceptance: No fixed TUI-created body/area; revision and managed-ticket restrictions remain effective.
- **T9.3:** Expose free-text comment input and retained history instead of the fixed action comment.
  Acceptance: Round trip multiple distinct comments and reject empty/stale submissions without losing input.
- **T9.4:** Expose AttachReviewRef with user input and preserve external editing, checklist display and canonical review guards.
  Acceptance: Review reference is validated and attached to the selected ticket; append wiring from Slice 5 remains reachable.
- **T9.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 10: Responsive Board and retained workflows

Ticket: `TKT-20261029101145` — `ops/default/ticket/ready/20261029101145-web-tui-parity-board-workflows.md`
Findings: G-016, G-028, G-031.
Prerequisite: Slice 9 integrated into dev.

- **T10.1:** Show all Board columns and Attention by default when width permits; retain a usable narrow view and maximize.
  Acceptance: Wide/narrow resize and focus preserve key selection and canonical column meaning.
- **T10.2:** Add a retained workflow list/detail surface within Dash with pagination, members, evidence and current eligible actions.
  Acceptance: The list handles empty/large sets and does not infer delivery from worker exit or an open PR.
- **T10.3:** Expose bounded work preview for ticket, slice, set, case and loose scope with canonical filters and ordering.
  Acceptance: Selection, later-ticket inclusion, policy and confirmation match the observed application snapshot.
- **T10.4:** Expose allowed lifecycle controls, allowance extension and integration-decision answers, plus visible case/ticket links.
  Acceptance: Stale/revision/budget/approval guards remain host-owned; unavailable controls explain why.
- **T10.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 11: Wizard entry and Dump in shared tabs

Ticket: `TKT-20261029101146` — `ops/default/ticket/ready/20261029101146-web-tui-parity-wizard-entry.md`
Findings: G-005, G-006, G-007, G-008.
Prerequisite: Slice 10 integrated into dev.

- **T11.1:** Expose Wizard and Open workflow in Cases and start with one Name/Create through StartNamedGuidedUseCase.
  Acceptance: No random-ID input ceremony or second name step; existing case/run re-entry is deterministic.
- **T11.2:** Replace the detached wizard presentation with case/run binding in the existing shell, stage indicator and shared forms.
  Acceptance: Stages map to existing tabs; no parallel menu, duplicated editor or raw namespace/UUID prompt.
- **T11.3:** Use the existing Dump form with a single Continue to quiz submission of the current content.
  Acceptance: Runtime G-008 evidence proves one exact retained input and retry reconciliation without duplicates.
- **T11.4:** Keep non-current wizard tabs read-only with Continue to the current stage; isolate manual and wizard controls.
  Acceptance: Manual save cannot accidentally break wizard progression; viewing another tab never submits work.
- **T11.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 12: Wizard Quiz and preparation

Ticket: `TKT-20261029101147` — `ops/default/ticket/ready/20261029101147-web-tui-parity-wizard-authoring.md`
Findings: G-009, G-010.
Prerequisite: Slice 11 integrated into dev.

- **T12.1:** Use the same Quiz component and controls in manual and wizard mode, bound to the retained run and current question.
  Acceptance: Deep/Fast, answer history, skip, answer, done and cancel remain functional without duplicate form logic.
- **T12.2:** Render specification and preparation stages in their existing tabs from the shared guided operation.
  Acceptance: Specification auto-approval, plan/ticket preparation and committed original inputs remain domain-owned.
- **T12.3:** Show stage-specific processing, suspended, pending-operation and blocker feedback with eligible actions.
  Acceptance: Partial result and timeout keep the original operation identity and saved input visible.
- **T12.4:** Retry the retained operation and handle stale/newer revisions without UI-owned orchestration or duplicate preparation.
  Acceptance: Restart/re-entry and out-of-order responses preserve exact run state and finite attempts.
- **T12.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 13: Wizard execution, results and recovery

Ticket: `TKT-20261029101148` — `ops/default/ticket/ready/20261029101148-web-tui-parity-wizard-results.md`
Findings: G-011, G-012, G-013, G-014.
Prerequisite: Slice 12 integrated into dev.

- **T13.1:** Use the case-filtered Board for wizard execution, progress, review links and eligible lifecycle controls.
  Acceptance: Counts distinguish validated, confirmed and remaining work using retained receipts.
- **T13.2:** Render structured terminal results with summary, branch/PR, delivery evidence and validation commands/exit codes.
  Acceptance: Open PR or successful worker cannot be shown as confirmed delivery; JSON export remains available.
- **T13.3:** Confirm cancellation with the same saved-document/completed-change warning and show Pause only where domain policy permits it.
  Acceptance: Owned runtime evidence for G-013 proves confirmation and retained artifacts; no implicit rollback.
- **T13.4:** Keep polling read-only, preserve unsent drafts and navigation, and retain later ideas/originals as existing secondary capabilities.
  Acceptance: Owned runtime evidence for G-014 proves no redirect, draft replacement or hidden progression while typing.
- **T13.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

### Slice 14: Parity acceptance and capability inventory

Ticket: `TKT-20261029101149` — `ops/default/ticket/ready/20261029101149-web-tui-parity-acceptance.md`
Findings: I-003.
Prerequisite: Slice 13 integrated into dev.

- **T14.1:** Reconcile all 38 gap rows and four incidental findings with exact tests and dispositions; close runtime uncertainties without changing source claims into assumed results.
  Acceptance: Every coverage row links a delivered slice/report and passing evidence or an explicit maintainer-approved exception.
- **T14.2:** Reconcile the palette and legacy capability inventory; gate unsupported safe checklist writers before confirmation without inventing a new domain or filesystem writer.
  Acceptance: No advertised unconditional failure remains; useful baseline extras retain tested access.
- **T14.3:** Extend owned production PTY and Web reference fixtures to exercise manual forms and the complete guided journey, including focus, resize, cancel and restart.
  Acceptance: Real Bubble Tea process plus real HTTP/Chromium, filesystem, Git and SQLite; deterministic model/provider responses only.
- **T14.4:** Update the User Guide and consolidate the delivered user-visible changelog outcomes; run final focused acceptance and publish the cycle report.
  Acceptance: No claims based on screenshots or reducer-only tests; no broad suite, new dependency or unapproved live project mutation.
- **T14.5:** Run the required and focused checks; update the ticket/tracker, create the slice report, open the recorded PR and verify controller-owned dev integration. Record exact commands, HEAD and PR, then close the slice.

## Completion and closure

All 14 slices must be integrated into dev; every task and ticket must have verified disposition; every report must be delivered with its canonical PR and integration hash. Update each coverage row with evidence/report references and a final disposition; retain the static source assessment unchanged. No unresolved source item may be silently labelled complete. A runtime check that disproves a suspected gap may close it as already satisfied only with recorded production evidence; other exclusions or deferrals require maintainer approval.

Publish `ops/default/report/web-tui-parity-cycle.md` with the baseline, integrated candidate, PR/report mapping, coverage totals, acceptance commands/results and residual limitations. Include the final gate's exact candidate separately from any closure-only commit. Update the User Guide and user-visible Unreleased notes as behavior becomes usable in each owning slice; final acceptance consolidates them without inventing a version or release. Do not modify main or mirrors.

Looper closure-only writes are limited to this set's spec, plan, tracker, coverage ledger, cycle report, 14 slice reports and exact ticket lifecycle paths. Source implementation belongs to recorded slice PRs. There is no blanket source-edit or final-gate repair authority. Preserve a genuine blocker with diagnostics and report it instead of bypassing checks or broadening the plan.

Related existing ticket: `TKT-20261023210422` (`ops/default/ticket/open/20261023210422-standardize-web-feedback-messages.md`) owns broader Web feedback consolidation; this set reuses neutral outcomes for the TUI and does not claim that separate Web work is resolved.
