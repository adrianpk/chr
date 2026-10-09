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

## Implementation Notes

The application service owns explicit extent admission, status projection and
manual handoff orchestration. Adapters parse bounded arguments and render the
shared results. A read-only target port uses remote observation without fetching.
Inspection holds repository exclusion but refuses recovery writes. Cleanup
releases only quiescent host descriptors and never changes task authority.

## Contracts Added or Changed

- `chr work slice`, `slice-set`, `use-case` and `loose` extend the explicit native
  command surface; `ticket` uses the same path under tickets authority.
- Retained `status`, `continue`, `pause`, `cancel`, `resume`, `approve` and
  `reconcile` controls support read-only previews and bounded JSON output.
- Snapshot hashes bind preview pagination and admission; status revisions bind
  member pagination. Approval uses the retained exact review-input hash.
- A manual harness supplies one JSON candidate object within 4096 bytes. Its
  frozen input hash, full branch, commit and tree must pass host verification.

## Files of Interest

- `internal/app/service_workflow_controls.go`, `service_workflow_status.go` and
  `service_workflow_manual.go`: adapter-neutral operations and bounded views.
- `internal/app/work_workflow_cli.go` and `work_workflow_render.go`: CLI parsing,
  dispatch and output.
- `internal/core/workflow_inspection.go`, `workflow_target_git.go` and
  `workflow_claim_lock.go`: read-only inspection/target IO and owned cleanup.
- `internal/app/work_workflow_*_test.go`: real parser, HTTPS Git preview,
  lifecycle, exact approval and supervised manual boundary tests.
- `docs/reference/workflow-controls.md`, the User Guide and `CHANGELOG.md`:
  delivered command behavior and operational guidance.

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

## Risks and Follow-ups

T10.2–T10.4 remain pending: TUI, Web/browser and daemon controls must exercise the
same service operations before this slice is ready for its canonical PR.
