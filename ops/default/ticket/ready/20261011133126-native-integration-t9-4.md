---
id: TKT-20261011133126
title: "T9.4: Verify crash and stop boundaries"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-9
source: chat
reported_at: 2026-10-11T13:31:26Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T9.4: Verify crash and stop boundaries

## Delivery Binding

Delivery set: native-integration-queue

Slice: 9 — Publication and uncertainty recovery

Task: T9.4

Plan: [Slice 9](../../plan/native-integration-queue.md#slice-9-publication-and-uncertainty-recovery)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add fault injection around intent, dispatch, server apply, response, restart, lease expiry and stop intent.

## Acceptance

All uncertain outcomes retain exclusion and stopped work reconciles any already delivered result before terminal cancellation.

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
