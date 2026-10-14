---
id: TKT-20261011133125
title: "T9.3: Observe canonical PR convergence"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-9
source: chat
reported_at: 2026-10-11T13:31:25Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T20:43:00Z
reviewed_at: 2026-10-13T21:50:58Z
branch: feat/slice-9-native-publication-recovery
pr: 383
commits:
  - 48a9539d023e63c96f04aa61ecb6baa46d77f9e2
---

# T9.3: Observe canonical PR convergence

## Delivery Binding

Delivery set: native-integration-queue

Slice: 9 — Publication and uncertainty recovery

Task: T9.3

Plan: [Slice 9](../../plan/native-integration-queue.md#slice-9-publication-and-uncertainty-recovery)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Poll bounded forge/Git observations, preserve the target barrier for pending metadata and handle external PR edits or target divergence.

## Acceptance

Closed-unmerged and inconsistent metadata cannot create receipts or trigger automatic target rollback.

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

Bounded canonical observation and polling retain the same submission, original
PR and deadline without publishing or issuing a receipt. Pending metadata and
PR/target inconsistencies retain target exclusion; stale callbacks cannot
rewrite the common journal. Prior exact source-P and target-history evidence
attributes lost transport output after later source work without fabricating a
successful response.

Development checks passed with nine unchanged Go files:

- `make verify` — zero lint issues; core 2394.437 s, app 108.657 s and all other packages passed.
- `go test ./internal/core -run '^(TestNativeCanonical.*|TestNativeLaterSource|TestNativeObservation.*|TestNativeSubmission.*|TestNativeGit.*|TestNativeDispatch.*|TestNativeReceipt.*)$' -count=1` — 162.892 s.
- `go test -race ./internal/core -run '^(TestNativeCanonical.*|TestNativeLaterSource|TestNativeObservation.*|TestNativeSubmission.*|TestNativeGit.*|TestNativeDispatch.*|TestNativeReceipt.*)$' -count=1` — 271.828 s.
- `git diff --check` — passed.

Source digest: `f3d69423789f84537de347f594c0701a76bd2a3bcb43bb5040465b8381a4bcbc`.
The owned validation service finished with exit status zero and no live check
process. All owned check fixtures finished. Focused normal/race Acta activity:
`01M4VHAK680FBXG2P887XS9Q90`, `01M4VHM9JAY36FRCT1QB06Z57X`.

The controller records ordered exact-HEAD validation separately before
integration. See the [Slice 9 report](../../report/slices/native-integration-queue/slice-9-publication-recovery.md).

Implementation: `48a9539d023e63c96f04aa61ecb6baa46d77f9e2`. Acta commit activity: `01M4VKZWK3H5X4ZDRMEDMHBP6J`.

Canonical slice PR: [PR #383](https://forge.adrianpk.com/adrianpk/chr/pulls/383). The controller records ordered exact-HEAD
validation separately before integration; task status is reviewing.
