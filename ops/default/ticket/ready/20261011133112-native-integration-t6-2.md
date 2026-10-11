---
id: TKT-20261011133112
title: "T6.2: Execute the merger in an owned checkout"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-6
source: chat
reported_at: 2026-10-11T13:31:12Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T6.2: Execute the merger in an owned checkout

## Delivery Binding

Delivery set: native-integration-queue

Slice: 6 — Bounded merger role

Task: T6.2

Plan: [Slice 6](../../plan/native-integration-queue.md#slice-6-bounded-merger-role)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Run the configured harness only for actionable conflicts/regressions with shared capacity, local tool permissions and host-owned lifecycle.

## Acceptance

The role cannot push, merge, approve, change scope or forge receipts, and timeout preserves ownership of a live child.

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
