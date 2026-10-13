# Slice 8: Persistent limits and decisions

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-8-native-integration-limits`
PR: pending

## Purpose

Retain finite integration and invocation budgets, share optional whole-call
reservations across agent roles, and require explicit current operator decisions
or bounded extensions before exhausted work can continue.

## Delivered Behavior

T8.1 adds admission-time merger and drift policy limits. Omitted values preserve
three merger launches and five drift revisions; explicit values are bounded from
one through ten. Existing worker, review, closing and submission limits remain
part of the retained invocation ledger. Every read verifies that admission
ceilings still agree with the frozen execution policy; positive grants retain
separate history and do not replace that admission image.

Call reservations require a future deadline within the retained worker timeout.
This applies to metered executor, reviewer, closing and merger reservations and
the existing unmetered merger mode. Invalid deadlines fail before consumption
changes. Retrying a reservation cannot replace its original deadline. Restart
and a new queue generation retain all previously charged counters and identities.

## Implementation Notes

The new optional policy fields use omitted zero values, so existing default
policy encodings remain compatible. Effective defaults are recorded in the
budget at admission. Entry retries retain explicit merger/drift policy values;
changed requested limits cannot silently replace frozen invocation authority.
The existing ledger, mutation exclusion, counter arithmetic and grant history
remain the owners of accounting.

## Contracts Added or Changed

- `WorkflowExecutionPolicy` accepts `max_merger_attempts` and `max_drift_revisions` with finite automatic bounds.
- Budget admission limits must match the invocation's frozen execution policy.
- Agent-call reservation deadlines cannot exceed the retained worker timeout or be extended by retry identity.

## Files of Interest

- `internal/core/workflow_integration_limits.go`: default and bounded admission policy.
- `internal/core/workflow_integration_limits_test.go`: persisted policy, exhaustion and deadline guards.
- `internal/core/workflow_integration_budget_store.go`: frozen-policy authority checks.
- `internal/core/workflow_integration_budget_reserve.go`: pre-charge deadline admission.
- `internal/core/workflow_invocation.go` and `internal/core/workflow_entry.go`: policy bounds and compatible entry restoration.

## Validation

The initial deadline regression failed for every role before correction:
`go test ./internal/core -run '^TestNativeBudgetDeadline$' -count=1` accepted
one-minute reservations under a five-second frozen worker timeout. The corrected
admission refuses those calls without changing retained budget state.

Development checks passed with Go 1.27.1:

- `make verify` — passed with zero lint issues; core 1853.274 s, app 99.190 s and all other packages passed.
- `go test ./internal/core -run '^(TestNativeBudgetDeadline|TestNativeLimitPolicy|TestNativeLimitExhaustion|TestNativeDeadlineRetry|TestWorkflowStartBounds|TestWorkflowEntry)$' -count=1` — passed in 2.308 s.
- `go test -race ./internal/core -run '^(TestNativeBudget|TestNativeLimit|TestNativeDeadline|TestNativeMergerUnmetered|TestWorkflowStartBounds|TestWorkflowEntry)' -count=1` — passed in 14.199 s after the test helper correction.
- `make lint-strict` — passed with zero issues.
- `git diff --check` — passed.

Initial linting required a smaller exhaustion test helper and explicit whitespace;
the corrected source retains the same assertions. The repository pre-commit
`make verify` check passed with all seven changed Go files unchanged throughout
validation. Owned checks exited naturally. Exact-head slice
validation will be recorded separately after all tasks and PR references are
committed and pushed.

T8.1 implementation: [94ea7e86745e9ab30d55bc9fa931fd23a91826bc](https://forge.adrianpk.com/adrianpk/chr/commit/94ea7e86745e9ab30d55bc9fa931fd23a91826bc).
Acta commit activity: `01M4T0JW66X49EPFWK9J7TQV4Q`.

## Risks and Follow-ups

T8.1 is implemented and checked. T8.2, T8.3 and T8.4 remain
recorded tasks for role-wide token enforcement, operator decisions/extensions
and owned-process recovery. The canonical Slice 8 PR remains pending.
