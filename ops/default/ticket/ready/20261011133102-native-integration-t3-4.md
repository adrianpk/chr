---
id: TKT-20261011133102
title: "T3.4: Rebuild indexes and validate migrations"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-3
source: chat
reported_at: 2026-10-11T13:31:02Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T3.4: Rebuild indexes and validate migrations

## Delivery Binding

Delivery set: native-integration-queue

Slice: 3 — Durable integration records

Task: T3.4

Plan: [Slice 3](../../plan/native-integration-queue.md#slice-3-durable-integration-records)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add rebuildable queue queries, strict cross-record validation and additive migration with malformed-byte preservation.

## Acceptance

Index loss is recoverable; stale or inconsistent records block mutation without manufacturing execution state.

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
