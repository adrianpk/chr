---
id: TKT-20261011133128
title: "T10.2: Replay task finalization exactly once"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-10
source: chat
reported_at: 2026-10-11T13:31:28Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T09:37:23Z
reviewed_at: 2026-10-14T11:02:41Z
branch: feat/slice-10-native-receipt-closure
pr: 384
commits:
  - c96ecba537e1bfedf6925c0324f0d49bd8332d89
---

# T10.2: Replay task finalization exactly once

## Delivery Binding

Delivery set: native-integration-queue

Slice: 10 — Receipts fanout and closure

Task: T10.2

Plan: [Slice 10](../../plan/native-integration-queue.md#slice-10-receipts-fanout-and-closure)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Use existing receipt/member journal keys for all slice tasks and independent tickets and expose finalizing until fanout completes.

## Acceptance

Interrupted fanout resumes without duplicate completion or missing task receipts.

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

Native receipt projection and existing-journal member fanout are implemented.
Real host crashes at each changed member, set, queue and cursor image recover
without duplicate completion or missing receipts. Independent-ticket, concurrent
finalizer, stop, corrupted-proof and unfinished-fanout barriers are covered.

Focused checks passed in 124.779 s; broader race checks passed in 843.435 s.
Required pre-commit `make verify` passed with zero lint issues on frozen source
`d4e17bdb213e312940bf2259465493988cd43927c4644241bb01a9f9005149ba`.
The owned validation service exited successfully with no live check process.
See the existing [Slice 10 report](../../report/slices/native-integration-queue/slice-10-receipt-closure.md).

Implementation: [c96ecba537e1bfedf6925c0324f0d49bd8332d89](https://forge.adrianpk.com/adrianpk/chr/commit/c96ecba537e1bfedf6925c0324f0d49bd8332d89).
Acta commit activity: `01M4X1988DMQYREF39TJTH8BHV`.

Canonical slice PR: [PR #384](https://forge.adrianpk.com/adrianpk/chr/pulls/384).
Acta opened activity: `01M4XF8TXK6RCHBWCJ5B2NKJZ8`. Final ordered clean-HEAD validation and
verified integration remain pending; task status is reviewing.
