---
id: TKT-20261011133129
title: "T10.3: Preserve descendant proof and report divergence"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-10
source: chat
reported_at: 2026-10-11T13:31:29Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T11:04:32Z
reviewed_at: 2026-10-14T13:21:05Z
branch: feat/slice-10-native-receipt-closure
commits: []
---

# T10.3: Preserve descendant proof and report divergence

## Delivery Binding

Delivery set: native-integration-queue

Slice: 10 — Receipts fanout and closure

Task: T10.3

Plan: [Slice 10](../../plan/native-integration-queue.md#slice-10-receipts-fanout-and-closure)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Verify receipt reachability under later target advancement and expose destructive remote rewrites without erasing history.

## Acceptance

A later descendant keeps valid delivery; removed proof creates a visible reconciliation blocker.

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

Confirmed receipts now retain later descendant proof and expose fetched history
divergence through the existing canonical observation and queue journal. Native
frontier continuation preserves cumulative role counts, immutable completed
members, consumed submission budgets and cancellation intent. Reconciliation
cannot be skipped through an optimistic phase callback.

Focused checks passed in 115.599 s; the corrected broader race check passed in
1218.447 s. Required pre-commit `make verify` passed with zero lint issues on
frozen 21-file source
`976a94dd3792872b7adf022924aa276b2706d79cad96494ae59c2c4eaa8c2b01`.
All checks and owned fixtures finished. See the existing
[Slice 10 report](../../report/slices/native-integration-queue/slice-10-receipt-closure.md).

Implementation commit and Acta references will be recorded in the follow-up.
PR and verified integration remain pending.
