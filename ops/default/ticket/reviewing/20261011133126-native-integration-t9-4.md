---
id: TKT-20261011133126
title: "T9.4: Verify crash and stop boundaries"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-9
source: chat
reported_at: 2026-10-11T13:31:26Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T21:52:36Z
reviewed_at: 2026-10-14T00:01:54Z
branch: feat/slice-9-native-publication-recovery
pr: 383
commits:
  - a06c392fc7a6b9169694989e436ea6afdec4010e
---

# T9.4: Verify crash and stop boundaries

## Delivery Binding

Delivery set: native-integration-queue

Slice: 9 — Publication and uncertainty recovery

Task: T9.4

Plan: [Slice 9](../../plan/native-integration-queue.md#slice-9-publication-and-uncertainty-recovery)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add fault injection around intent, dispatch, server apply, response, restart, lease expiry and stop intent.

## Acceptance

All uncertain outcomes retain exclusion and stopped work reconciles any already delivered result before terminal cancellation.

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

The initial stop probe reproduced terminal cancellation while native Git
application awaited canonical observation with no receipt. The implementation
now retains invocation/entry stop intent and permits observation before terminal
stop. Send admission rechecks stop before commands. Known changed-input
non-application may finalize retained stop; receipt confirmation alone cannot
skip required fanout.

Actual host subprocess fixtures cover intent, dispatch, send, server application,
response retention, restart and expired task leases without duplicate authority
or counter refill. Preliminary overlay checks and their Acta evidence are in the
[Slice 9 report](../../report/slices/native-integration-queue/slice-9-publication-recovery.md).

Initial focused repository tests passed in 57.018 s. The first broader check
exposed a receipt fixture referencing a different transaction than the retained
submission. Its exact-response correction passed the receipt case in 5.464 s;
strict lint and whitespace checks passed. Failure/correction evidence and Acta
activity are retained in the Slice 9 report.

On the prior six-file source, broader normal/race checks passed in 251.101 s
and 423.165 s. Pre-commit `make verify` then failed in the canonical deadline
case when a real Git read reached the persisted bound. The corrected assertion
accepts only the documented typed deadline error and preserves every retained
blocker/deadline, expired-IO and explicit same-PR recovery assertion. The narrow
deadline cases passed in 14.256 s; strict lint passed after extracting the
explicit recovery helper. Complete failed-run and Acta evidence is in the report.

Fresh normal/race checks passed in 247.969 s and 433.723 s on seven unchanged
Go files. Pre-commit `make verify` passed with zero lint issues; core 2473.043 s,
app 113.530 s and all other packages passed. The owned
`chr-native-slice9-t94-validation-e305eefe.service` finished with exit status
zero and no live check process. All owned check fixtures finished.
Source digest: `e305eefe0d88cebe669d08a4f8d2a7388251eec1b61a2793dd767627bcff3758`.

Implementation and development validation are complete. The controller records
ordered exact-HEAD slice validation separately before integration. Verified integration remains pending.

Implementation: [a06c392fc7a6b9169694989e436ea6afdec4010e](https://forge.adrianpk.com/adrianpk/chr/commit/a06c392fc7a6b9169694989e436ea6afdec4010e).
Acta commit activity: `01M4VVFZY92DBXEJYPP80BP2W8`.

Canonical slice PR: [PR #383](https://forge.adrianpk.com/adrianpk/chr/pulls/383). The controller records ordered exact-HEAD
validation separately before integration; task status is reviewing.
