---
id: TKT-20261011133127
title: "T10.1: Verify immutable dual-ref receipts"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-10
source: chat
reported_at: 2026-10-11T13:31:27Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T08:20:18Z
reviewed_at: 2026-10-14T09:35:54Z
branch: feat/slice-10-native-receipt-closure
pr: 384
resolution: fixed
closed_at: 2026-10-14T18:42:26Z
commits:
  - b0a64a02dbc13ad61220cde8502fca66f7a332f9
---

# T10.1: Verify immutable dual-ref receipts

## Delivery Binding

Delivery set: native-integration-queue

Slice: 10 — Receipts fanout and closure

Task: T10.1

Plan: [Slice 10](../../plan/native-integration-queue.md#slice-10-receipts-fanout-and-closure)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Bind the exact candidate/submission/profile, original S, published P, validated B, method history and same canonical merged PR.

## Acceptance

An incompatible merge result, missing proof or closed-only PR cannot be accepted.

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

The receipt operation binds the exact retained candidate, submission and profile
with actual Git application and the latest original canonical merged PR proof.
Lost responses remain unknown transport evidence; a separately retained Git
application observation may prove publication without replacing the raw result.
Pending, closed-unmerged and incompatible results cannot create receipts.

Initial focused receipt/restart checks passed before the substituted-application
proof cases were added. Strict lint passed with zero issues on the current
nine-file source. Broader normal/race checks passed in 375.131 s and 603.402 s on that unchanged
source; both check processes and their owned fixtures finished. Acta activity:
`01M4WSYK9EZM99G585ZDXFTXRP`. Required pre-commit `make verify` passed on that same source with zero lint
issues; core 2401.672 s, app 107.021 s and all other packages passed. The owned
service finished with exit status zero and no live check process. All owned
check fixtures finished.

Evidence: [Slice 10 report](../../report/slices/native-integration-queue/slice-10-receipt-closure.md).

Implementation: [b0a64a02dbc13ad61220cde8502fca66f7a332f9](https://forge.adrianpk.com/adrianpk/chr/commit/b0a64a02dbc13ad61220cde8502fca66f7a332f9).
Acta commit activity: `01M4WWAFZQ5NQSPN1ZRKS10XTK`.

Canonical slice PR: [PR #384](https://forge.adrianpk.com/adrianpk/chr/pulls/384).
Acta opened activity: `01M4XF8TXK6RCHBWCJ5B2NKJZ8`. All four configured
ordered exact-HEAD checks passed; task status is solved.

Verified integration: [#384](https://forge.adrianpk.com/adrianpk/chr/pulls/384),
`9726200044d15b4e66bacfe1334de4a9410f4bdc` on `dev` at 2026-10-14T18:36:18Z. Controller Acta merge:
`01M4XV6YGZZBS8CGG3AJ2EHKT7`.
