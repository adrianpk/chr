# Slice 5: Slice candidates and delivery receipts

Status: delivered

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-5-workflow-delivery`

PR: [#368](https://forge.adrianpk.com/adrianpk/chr/pulls/368)

## Purpose

Deliver one aggregate candidate, canonical PR and verified shared receipt for an
ordered slice while preserving task validation and recoverable host ownership.

## Delivered Behavior

Remaining tasks execute on one retained cumulative lane. A separate host gate
binds aggregate acceptance to immutable task/input/contract and actual Git
images. Aggregate failure retains task evidence; explicit repair preserves the
unchanged prefix and invalidates only the affected suffix.

Canonical PR intent precedes provider IO. All tasks share one set-owned branch,
candidate and PR. Closed, mismatched or ambiguous identities block; lost create
responses and host restarts retain observation ownership without another create.

External integration requires matching canonical merged PR metadata and a
fetched target containing its result. Merge verifies source/base parents and the
combined tree; rebase verifies every ordered linear result tree; squash verifies
one combined commit on its base. Moved bases and later target commits are proved
through actual history, rather than worker claims or mutable PR base metadata.

One journal creates the immutable shared receipt and finalizes every task,
the owning slice and invocation. Actual host death during fan-out replays once.
Retries and concurrent callers reuse the same receipt, retain task evidence and
allocate no extra worker or PR. Later set members and set closure remain at their
recorded boundaries.

## Implementation Notes

`WorkflowStore` owns frozen membership, fenced aggregate validation, immutable
candidate/provider/delivery evidence and guarded multi-image transitions. The
central application service delegates these operations through configured
provider and trusted host Git ports. Constructors perform no IO.

Git verification uses an owned temporary bare repository, disables replacement
objects, retains trusted host credentials and rechecks candidate images after IO.
It bounds execution to three minutes, Git/proof output to 1 MiB and ordered
history to 4096 commits. Temporary repositories are removed after observation.
Finalization uses the existing 1 MiB journal bound and fails before completion
writes when the complete fan-out exceeds it.

## Contracts Added or Changed

- `chr-slice-candidate/v1` binds aggregate host acceptance to ordered task proof.
- `chr-slice-pr/v1` and immutable provider observations retain canonical PR identity and uncertain creation.
- `chr-slice-delivery-proof/v1` binds external provider facts to actual method-specific target/tree/history proof.
- Invocation candidate, PR, delivery-proof and receipt references include immutable image hashes.
- Every integrated task retains one common `slice_delivery_ref` and typed delivery projection; the set-owned PR and receipt remain authoritative.
- Repair, observation, receipt introduction and task/set/run fan-out use guarded replay rather than status inference.

## Files of Interest

- [Aggregate execution and repair](../../../../../internal/core/workflow_slice.go).
- [Canonical PR lifecycle](../../../../../internal/core/workflow_pull_request.go).
- [External integration proof](../../../../../internal/core/workflow_integration.go).
- [Git method proof](../../../../../internal/core/workflow_integration_history.go).
- [Shared receipt finalization](../../../../../internal/core/workflow_finalization.go).
- [Host-crash fan-out checks](../../../../../internal/core/workflow_finalization_recovery_test.go).
- [Execution reference](../../../../../docs/reference/workflow-execution.md).

## Validation

All task pre-commit `make verify` checks passed with the retained owned CLI
curation fixture documented in the [Slice 2 report](slice-2-workflow-materialization.md).
Strict lint reported zero issues. Successful focused race checks:

- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowSlice|TestServiceWorkflowSlice)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowSlicePR|TestWorkflowPR|TestServiceWorkflowSlicePR)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowIntegration|TestServiceWorkflowIntegration)' -count=1`.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowFinalization|TestServiceWorkflowFinalization)' -count=1`.
- `git diff --check`.

Real subprocess, linked-worktree, bare-remote and crash fixtures cover aggregate
repair/ownership, lost PR responses, concurrent callers, wrong provider/branch
identity, forged final/intermediate trees, each receipt fan-out replacement and
restart reuse. Central service checks exercise the same operations. No subjective
review pass was requested.

Final PR-head evidence is recorded through the controller for this ordered list:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

All four commands passed on clean head `6cfaedd9c12f7a6dcfc562804917c32c5f6daaca`,
recorded in the controller state directory as the matching `checks/<HEAD>.json`
and log. Canonical PR #368 merged that exact head through Rebase + Fast-forward
into dev; canonical and local dev match. Acta merge event:
`01M4ESBJ8W5EGFTJR23A3CCXN3`. The four task tickets are solved. These candidate
checks are distinct from the delivery set's later full gate.

## Risks and Follow-ups

None recorded.
