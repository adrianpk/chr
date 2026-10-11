---
id: TKT-20261011133101
title: "T3.3: Extend attempt and budget journals"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-3
source: chat
reported_at: 2026-10-11T13:31:01Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T3.3: Extend attempt and budget journals

## Delivery Binding

Delivery set: native-integration-queue

Slice: 3 — Durable integration records

Task: T3.3

Plan: [Slice 3](../../plan/native-integration-queue.md#slice-3-durable-integration-records)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add merger/drift/token reservations and explicit Git-applied/forge-pending submission facts, preserving existing consumed counters.

## Acceptance

Restart and candidate replacement cannot refill counters or release unknown reservations.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through Looper.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only Looper submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Implementation, validation, PR and verified integration evidence are pending.
