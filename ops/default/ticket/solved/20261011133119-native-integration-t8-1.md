---
id: TKT-20261011133119
title: "T8.1: Enforce merger drift and submission counters"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-8
source: chat
reported_at: 2026-10-11T13:31:19Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T06:10:27Z
reviewed_at: 2026-10-13T06:51:34Z
branch: feat/slice-8-native-integration-limits
pr: 382
resolution: fixed
closed_at: 2026-10-13T18:12:40Z
commits: ["94ea7e86745e9ab30d55bc9fa931fd23a91826bc"]
---

# T8.1: Enforce merger drift and submission counters

## Delivery Binding

Delivery set: native-integration-queue

Slice: 8 — Persistent limits and decisions

Task: T8.1

Plan: [Slice 8](../../plan/native-integration-queue.md#slice-8-persistent-limits-and-decisions)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Apply approved default/bounded limits and retained worker timeout across unit and invocation lifecycle.

## Acceptance

Restart, rebase and queue re-entry cannot reset consumed attempts or extend deadlines.

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

Implementation and development validation are retained below; Final exact-head slice validation and verified integration are recorded below.

## Evidence

T8.1 applies frozen default or explicitly bounded merger/drift limits, refuses
budget images that disagree with admitted policy and checks role-call deadlines
before consumption changes. Merger, drift and submission reservations remain
charged through restart and replacement generations; retry cannot extend an
accepted call deadline.

Development checks passed with Go 1.27.1:

- `make verify` — zero lint issues; all packages passed, including core in 1853.274 s and app in 99.190 s.
- `go test ./internal/core -run '^(TestNativeBudgetDeadline|TestNativeLimitPolicy|TestNativeLimitExhaustion|TestNativeDeadlineRetry|TestWorkflowStartBounds|TestWorkflowEntry)$' -count=1` — passed in 2.308 s.
- `go test -race ./internal/core -run '^(TestNativeBudget|TestNativeLimit|TestNativeDeadline|TestNativeMergerUnmetered|TestWorkflowStartBounds|TestWorkflowEntry)' -count=1` — passed in 14.199 s after the helper correction.
- `make lint-strict` — passed with zero issues.
- `git diff --check` — passed.

The oversized-deadline regression failed before correction. The existing Slice 8
report retains that reproduction, corrected checks and task checks. Final exact-head controller evidence passed and is retained below.

T8.1 implementation: [94ea7e86745e9ab30d55bc9fa931fd23a91826bc](https://forge.adrianpk.com/adrianpk/chr/commit/94ea7e86745e9ab30d55bc9fa931fd23a91826bc).
Acta commit activity: `01M4T0JW66X49EPFWK9J7TQV4Q`.

Canonical slice PR: [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382). The controller records ordered exact-HEAD
validation separately before integration; task status is solved.

Verified integration: [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382),
`0096f6e960ecc25e7aae72a8a97fdc53c689418c` on `dev` at 2026-10-13T18:09:40Z. Controller Acta merge:
`01M4V79DVNVSXZG5D9TMZ0NMSX`. All four configured ordered exact-head checks passed; this task is solved.
