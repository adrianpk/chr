# Slice 7: Ordered set and use-case continuation

Status: reviewing

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-7-workflow-continuation`

PR: pending

## Purpose

Continue one frozen invocation through ordered slice PRs, receipt-confirmed
frontiers and exact set closure, preserving accepted work across restart.

## Delivered Behavior

A later slice cannot start before its predecessor has a verified shared receipt
and fresh canonical PR/Git proof. Each slice owns a distinct lane and PR.
Frontier release archives the previous delivery image, clears active bindings
and hydrates the next lane from the observed canonical target without moving
the host checkout. Unsliced use-case tickets remain excluded. Set boundaries
retain a separate closing barrier before another associated set can claim work.

Equivalent requests attach to the same invocation and frozen policy. Bounded
continuation checkpoints preserve validated and receipt-confirmed cursors,
request identity, consumed attempts and immutable evidence. Restart skips
accepted task, PR and integration work. Identical waits reuse their checkpoint;
corrupt checkpoints, explicit stops and changed frozen scope block successors.

Default admission excludes later work. Explicit accepted later admission adds
matching ready work strictly after the confirmed frontier. Quiescent approved
amendments preserve existing order, execution contracts and delivered images,
allocate native task keys and retain immutable plan/amendment evidence.
Retroactive work and external reassociation reject. Explicit reopening preserves
prior closure bytes and terminal invocations.

Closing observes every actual member, including tickets excluded by frozen
admission. All tasks and slices must have verified integration receipts.
The host freezes their revisions and hashes, approved plan and closing contract,
canonical target commit/tree and method-specific proofs. Gate commands run on
an isolated exact target; changed members, evidence or target prevent closure.
Failure retries closure alone without replaying integrated tasks. Empty sets
still require an explicit gate or approved waiver. Confirmed closure releases
the next set from its refreshed base; the final set makes the invocation terminal.

## Implementation Notes

`WorkflowStore` owns traversal, membership, immutable evidence and replayable
record transitions. The central application service delegates the same
operations through trusted canonical provider and Git ports. Workers do not
receive host ownership credentials or choose their publication target.

Closing admission and confirmation observations have finite three-minute host
deadlines. Required gate commands share their contract timeout; explicit waivers
retain a one-minute closing lease. Closing uses existing repository serial
capacity, operation locks, PID birth fencing and descendant supervision.
Pause/cancel intent stays active until process quiescence. Expired owner recovery
retains interruption and adopted operation-lock identity without confirming a gate.

Each set invocation retains 128 closing attempts across retry and restart.
Distinct explicit invocations have distinct budgets while preserving set-owned
history, bounded to 4096 attempts. Continuation permits 1–128 accepted steps per
call and 4096 immutable checkpoints. Files and journals retain the 1 MiB bound;
commands retain existing argv and 16 KiB per-stream output limits.

## Contracts Added or Changed

- Immutable frontier images and `confirmed_cursor` retain prior receipts, PRs and fresh target proof while `execution_target` supplies the next lane base.
- `chr-continuation-checkpoint/v1` binds equivalent request identity, progress, accepted steps and parent evidence without a scope override.
- `IncludeLaterTickets` stays frozen at admission. Approved `chr-set-amendment/v1` files preserve original plan bytes and all previously delivered projections.
- `chr-set-closing-input/v1` freezes all actual members, ordered slices, plan/contract hashes and exact canonical target/tree.
- `chr-set-closing-attempt/v1` retains fenced ownership, per-invocation budget, finite lease, process identities and hashed logs/results.
- `chr-set-closure/v1`, the completed set projection and invocation `set_closures` frontier share one replayable confirmation. Reopening retains historical receipts.

## Files of Interest

- [Verified frontier release](../../../../../internal/core/workflow_frontier.go).
- [Bounded continuation](../../../../../internal/core/workflow_continuation.go).
- [Approved amendments](../../../../../internal/core/workflow_amendment.go).
- [Exact set closing](../../../../../internal/core/workflow_closing.go).
- [Closing owner recovery](../../../../../internal/core/workflow_closing_recovery.go).
- [Central closing service](../../../../../internal/app/service_workflow_closing.go).
- [Real ordered set advancement](../../../../../internal/core/workflow_closing_sets_test.go).
- [Execution reference](../../../../../docs/reference/workflow-execution.md).

## Validation

All four task pre-commit `make verify` checks passed with the retained owned CLI
curation fixture described in the [Slice 2 report](slice-2-workflow-materialization.md).
Strict lint reported zero issues. Gates use a scoped uncached 30-minute Go test
timeout. Successful focused race checks:

- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowFrontier|TestServiceWorkflowFrontier)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowContinuation|TestServiceWorkflowContinuation)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowAmendment|TestWorkflowMembership|TestWorkflowReopening|TestServiceWorkflowMembership)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowClosing|TestWorkflowContinuation|TestWorkflowFrontier|TestWorkflowMembership|TestServiceWorkflowClosing|TestServiceWorkflowContinuation|TestServiceWorkflowFrontier)' -count=1`.
- `git diff --check`.

Actual bare-Git fixtures and controlled canonical PR observations cover two sets,
three distinct PR identities, receipt barriers, fresh bases, frozen exclusions,
approved later work, exact closing snapshots, closure-only retries and explicit
empty decisions. Journal interruption, expired owners, descendant stops, budget
exhaustion, corrupted evidence and prior-plan/closure/terminal immutability are
covered through native operations. Shared service tests use the same contracts.

The controller records final clean-head PR evidence for this ordered list:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

These candidate checks are distinct from the delivery set's later full gate.

## Risks and Follow-ups

None recorded.
