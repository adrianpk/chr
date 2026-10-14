---
id: TKT-20261011133130
title: "T10.4: Enforce set closure barriers"
status: in_progress
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-10
source: chat
reported_at: 2026-10-11T13:31:30Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T13:22:10Z
branch: feat/slice-10-native-receipt-closure
commits: []
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

Implementation, validation, PR and verified integration evidence are pending.
