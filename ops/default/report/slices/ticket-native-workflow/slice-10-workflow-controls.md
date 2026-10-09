# Slice 10: Shared operation controls

Status: drafting

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-10-workflow-controls`

PR: pending

## Purpose

Expose the shared ticket-native work operations through CLI, TUI, Web and daemon
controls with consistent selection, retained progress, recovery and approval.

## Delivered Behavior

T10.1 adds explicit ticket, slice, set, use-case and loose CLI extents with bounded
previews and continuation. Preview pages bind selected records to the observed
remote target and leave local Git objects, refs and operational files unchanged.
Retained status distinguishes host task validation from confirmed ticket
integration and reports member revisions, policy, candidate, PR, approval and
delivery evidence. Pending journals block inspection without implicit replay.

CLI pause, cancel, resume, approval and reconciliation delegate to the existing
domain operations. Approval requires the displayed exact binding and rejects a
changed candidate, PR head, target or policy. Manual reconciliation supervises a
verbatim harness argv, freezes public inputs and host-validates one bounded exact
Git report. Failed reports retain checkpoints without claiming task success.

T10.2 adds selected native ticket, slice, set, use-case and independent-work TUI
previews. Confirmation retains the selected task revision, complete snapshot and
target. Retained lifecycle controls pass the displayed invocation revision and
approval hash unchanged. Stale failures keep current progress visible. The Dash
shows bounded retained pages, waits, blocker/next-action owner, member order and
revision, candidate/PR/approval, continuation and delivery evidence. Status and
stop controls remain available during execution; a second execution is not queued.

T10.3 adds native ticket/case Web previews, exact snapshot/revision
confirmations and shared lifecycle controls. The Board exposes native tickets
without requiring the legacy ticket directory. Retained views show bounded
progress, ordered members, policy, waits, evidence and approval bindings. Failed
writes render current status with a visible failure. Actual browser forms execute
a task checkpoint, continue its candidate and enforce ownership and review waits.

## Implementation Notes

The application service owns explicit extent admission, status projection and
manual handoff orchestration. Adapters parse bounded arguments and render the
shared results. A read-only target port uses remote observation without fetching.
Inspection holds repository exclusion but refuses recovery writes. Cleanup
releases only quiescent host descriptors and never changes task authority.

TUI mappings stay in adapter files. Bound admission validates the selected task
and snapshot under shared domain exclusion. Lifecycle writes compare invocation
revisions under that same exclusion; continuation retains acceptance evidence
before executing a checkpoint. Asynchronous reads cannot replace a newer TUI
revision. Inspection checks all linked namespaces/worktrees before discovery,
preventing journal replay during status or dashboard reads.

Web mappings remain in adapter files and use the same bound admission/control
service. Read-only previews retain their confirmation inputs; revision-guarded
member pages reject stale reads. The smoke script uses an owned executable/server
PID and real browser actions; screenshots supplement those actions. Temporary
HTTP, Git, worker and browser fixtures are cleaned after their checks.

## Contracts Added or Changed

- `chr work slice`, `slice-set`, `use-case` and `loose` extend the explicit native
  command surface; `ticket` uses the same path under tickets authority.
- Retained `status`, `continue`, `pause`, `cancel`, `resume`, `approve` and
  `reconcile` controls support read-only previews and bounded JSON output.
- Snapshot hashes bind preview pagination and admission; status revisions bind
  member pagination. Approval uses the retained exact review-input hash.
- A manual harness supplies one JSON candidate object within 4096 bytes. Its
  frozen input hash, full branch, commit and tree must pass host verification.

- TUI native confirmation binds task ID/revision and the full preview snapshot.
  Invocation controls bind ID/revision; approval also binds its exact input hash.
- Retained TUI status uses pages of 16 invocations, eight visible members and
  recent evidence rows, and one checkpoint per continue activation. Status
  exposes frozen policy, selected task/slice, capacity resume phase, continuation
  evidence and next-action owner without host credentials.
- Guarded continuation acceptance is a durable checkpoint; terminal continuation
  remains observational and does not add attempts or acceptance evidence.

- `GET /workflows` and `/workflows/fragment` observe previews or retained status;
  `POST /workflows/actions` confirms a snapshot or invocation revision.
- Native Web forms reject duplicate fields and exceedance of 32 KiB bodies,
  8 KiB confirmation bindings, or bounded control fields.
- Retained Web status remains observable under tracker compatibility; new native
  admission requires tickets authority. Current views use 16 invocation choices,
  eight members/evidence rows and at most one shared checkpoint per activation.

## Files of Interest

- `internal/app/service_workflow_controls.go`, `service_workflow_status.go` and
  `service_workflow_manual.go`: adapter-neutral operations and bounded views.
- `internal/app/work_workflow_cli.go` and `work_workflow_render.go`: CLI parsing,
  dispatch and output.
- `internal/core/workflow_inspection.go`, `workflow_target_git.go` and
  `workflow_claim_lock.go`: read-only inspection/target IO and owned cleanup.
- `internal/app/work_workflow_*_test.go`: real parser, HTTPS Git preview,
  lifecycle, exact approval and supervised manual boundary tests.
- `internal/app/tui_adapters_workflow*.go`, `internal/tui/workflow_actions.go`
  and `workflow_commands.go`: mapping, selection, confirmation and progress.
- `internal/core/workflow_control_guard.go` and the TUI/core boundary tests:
  locked revision checks, acceptance evidence and real race/failure checks.
- `docs/reference/workflow-controls.md`, the User Guide and `CHANGELOG.md`:
  delivered command behavior and operational guidance.

- `internal/app/web_workflow*.go`, `internal/webui/workflow.go` and the embedded
  workflow template: native projection, guarded actions and owned browser tests.
- `scripts/qa/web-ui-smoke.sh`: actual activation/conflict/review/continuation
  browser assertions alongside existing page, fragment and screenshot checks.

## Validation

T10.1 passed:

- `make verify`
- `go test ./internal/app -run '^TestWorkflowCLI(Start|Admission)$' -count=1`
- `go test -race ./internal/app ./internal/core -run '^(TestWorkflowCLI.*|TestWorkflowInspectionJournal|TestWorkflowClaimClose|TestServiceWorkNative|TestWorkTicketRunsSelectedTicket)$' -count=1`
- `git diff --check`

Go 1.27.1, the retained CLI curation fixture and a scoped uncached 30-minute test
timeout were used for the repository gate. Strict lint reported zero issues and
all packages passed; core tests took 781.229 seconds and app tests 68.590 seconds.
Focused race checks passed in 24.465 seconds for app and 1.393 seconds for core.
Real root parsing and byte snapshots prove nonmutating previews; moved targets
and PR heads reject stale bindings; actual manual harnesses require host proof.
The final ordered controller checks for the canonical PR head remain pending.

T10.2 passed:

- `make verify`
- `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`
- `go test -race ./internal/core -run '^(TestWorkflowControlCAS|TestWorkflowGuardedAdmission|TestWorkflowInspection.*|TestWorkflowTerminalControl)$' -count=1`
- `git diff --check`

Go 1.27.1, the retained CLI curation fixture and scoped uncached 30-minute timeout
were used. Strict lint reported zero issues; all repository packages passed.
App tests took 74.313 seconds and core tests 768.568 seconds. Adapter race checks
passed in 184.775 seconds for app, 1.190 seconds for TUI and 1.189 seconds for Web;
focused domain race checks passed in 1.882 seconds. The 37 gate input files stayed
unchanged. Real adapter/model checks preserve selected bindings, show stale
failures, execute one actual task/continuation checkpoint, reject a moved PR head,
and keep readonly previews/journals byte-identical. Local documentation links and
whitespace passed. Final PR-head controller evidence remains pending.

T10.3 passed:

- `make verify`
- `go test ./internal/app ./internal/webui -run '^TestWebWorkflow' -count=1`
- `go test -race ./internal/app ./internal/webui -run '^TestWebWorkflow' -count=1`
- `scripts/qa/web-ui-smoke.sh`
- `git diff --check`

Go 1.27.1 and the retained curation fixture were used. The repository gate used
the scoped uncached 30-minute test timeout; strict lint reported zero issues and
all packages passed. App tests took 89.521 seconds and core tests 826.869 seconds.
The 18 gate inputs stayed unchanged. Focused race checks used the installed
browser and passed for app in 29.978 seconds; the smoke script passed real browser
activation, stale ownership/revisions, review wait, moved-head rejection, exact
approval and continuation. Handler checks also cover pause/resume/cancel, input
bounds, native ticket/case controls, byte-identical previews and retained status
under compatibility mode. Owned browser/server fixtures finished and stopped.
Documentation links and whitespace passed. Final PR-head controller evidence
remains pending.

## Risks and Follow-ups

T10.4 remains pending: daemon lifecycle, manual status parity and the notification
boundary must be delivered before this slice is ready for its canonical PR.
