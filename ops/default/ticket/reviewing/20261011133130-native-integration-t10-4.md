---
id: TKT-20261011133130
title: "T10.4: Enforce set closure barriers"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-10
source: chat
reported_at: 2026-10-11T13:31:30Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T13:22:10Z
reviewed_at: 2026-10-14T15:05:14Z
branch: feat/slice-10-native-receipt-closure
commits:
  - a9b33ea1f09e16fa5167c0fe75b2a4218b7a8d47
---

# T10.4: Enforce set closure barriers

## Delivery Binding

Delivery set: native-integration-queue

Slice: 10 — Receipts fanout and closure

Task: T10.4

Plan: [Slice 10](../../plan/native-integration-queue.md#slice-10-receipts-fanout-and-closure)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Require intended receipts, full-gate evidence, settled finalization and no unknown publication before closure.

## Acceptance

Original and corrective delivery closure remain separate verified outcomes with preserved historical evidence.

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

Native set closure now reads the existing receipt, fanout archive and confirmed
common publication journal. It requires all intended members and receipts,
freezes native proof hashes and rechecks canonical PR/Git evidence after the
configured gate. Unresolved publications across namespaces block closure;
original and corrective sets retain separate gate attempts and closure receipts.

Focused native/legacy closure, frontier and finalization checks passed in
315.057 s. The broader race check passed in 541.991 s on frozen 11-file source
`c8933a047c6f9dfe4e0b4b887cf6e2da8baf77bd7c9e8528a072ec338dd442f0`.
Acta race activity: `01M4XC51WKBG0ACGS0Q39BVECA`. Required pre-commit
`make verify` passed with zero lint issues; core tests passed in 3078.743 s,
application tests passed in 121.224 s and all other packages passed. All checks
and owned fixtures finished; the 11 Go source files remained unchanged.
Pre-commit gate activity: `01M4XF2XHX2N5XQ3KZJW7GW6BZ`. See the existing
[Slice 10 report](../../report/slices/native-integration-queue/slice-10-receipt-closure.md).

Implementation: [a9b33ea1f09e16fa5167c0fe75b2a4218b7a8d47](https://forge.adrianpk.com/adrianpk/chr/commit/a9b33ea1f09e16fa5167c0fe75b2a4218b7a8d47).
Acta commit activity: `01M4XF61EDBVN4CXKBQYSR8H53`.
Canonical PR, final ordered clean-HEAD validation and verified integration remain
pending.
