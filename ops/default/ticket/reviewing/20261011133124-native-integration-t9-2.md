---
id: TKT-20261011133124
title: "T9.2: Reconcile atomic rejection and lost responses"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-9
source: chat
reported_at: 2026-10-11T13:31:24Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T19:24:35Z
reviewed_at: 2026-10-13T20:39:33Z
branch: feat/slice-9-native-publication-recovery
pr: 383
commits: ["ae54f89aff1248e7cee90c5e033ac923ec7c77c7"]
---

# T9.2: Reconcile atomic rejection and lost responses

## Delivery Binding

Delivery set: native-integration-queue

Slice: 9 — Publication and uncertainty recovery

Task: T9.2

Plan: [Slice 9](../../plan/native-integration-queue.md#slice-9-publication-and-uncertainty-recovery)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Classify definitive not-applied, unknown and Git-applied outcomes with actual refs and method proof instead of command exit alone.

## Acceptance

No lost response triggers another submission or partial ref update.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through the delivery controller.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only the delivery controller submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Implementation and development validation are complete on the configured Slice
9 branch. Durable actual advertisement/command facts precede IO; returned host
completion and transport responses stay in the same common journal. Independent
actual ref/history observations classify Git application and definitive
non-application. Unknown outcomes retain exclusion and cannot resubmit or refill
counters. Git classification creates no forge receipt.

Development checks passed on fourteen unchanged Go files:

- `make verify` — zero lint issues; core 2311.960 s, app 110.734 s and all other packages passed.
- `go test ./internal/core -run '^(TestNativeDispatch.*|TestNativeGit.*|TestNativeSubmission.*|TestNativeReview.*|TestNativePreparation.*|TestNativeBudget.*|TestWorkflowGit.*)$' -count=1` — passed in 355.662 s.
- `go test -race ./internal/core -run '^(TestNativeDispatch.*|TestNativeGit.*|TestNativeSubmission.*|TestNativeReview.*|TestNativePreparation.*|TestNativeBudget.*|TestWorkflowGit.*)$' -count=1` — passed in 560.525 s.
- `make lint-strict` — zero issues.
- `git diff --check` — passed.

The initial actual-dispatch evidence test reproduced absent transport retention.
The current real SSH fixture truncates the receive-pack report after both refs
apply; actual Git proof recovers application while the response stays unknown.
Definitive rejection and source/target races cannot partially update refs.
Unchanged and moved-ref unknown results stay excluded, and stale observations
leave the common journal intact.

Source digest: `839b196f9e05fd84bab9e3e0e0f88df0fceb627442bea018de2de3e53905f53d`.
Command/process evidence: `/tmp/chr-native-slice9-t9-2-*`.
All owned check processes and fixtures finished. Focused-check Acta activity:
`01M4VCSGYGAEN61C0KYC9XAK7W`; broader normal/race activity:
`01M4VDRC8Z5TQAX002QF2DW5TW`.

Implementation: [ae54f89aff1248e7cee90c5e033ac923ec7c77c7](https://forge.adrianpk.com/adrianpk/chr/commit/ae54f89aff1248e7cee90c5e033ac923ec7c77c7).
Acta commit activity: `01M4VFXT6NZ0S5TT0V9BJ8W6PF`.
Verified integration remains pending.

Canonical slice PR: [PR #383](https://forge.adrianpk.com/adrianpk/chr/pulls/383). The controller records ordered exact-HEAD
validation separately before integration; task status is reviewing.
