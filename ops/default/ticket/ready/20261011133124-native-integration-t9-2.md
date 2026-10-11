---
id: TKT-20261011133124
title: "T9.2: Reconcile atomic rejection and lost responses"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-9
source: chat
reported_at: 2026-10-11T13:31:24Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T9.2: Reconcile atomic rejection and lost responses

## Delivery Binding

Delivery set: native-integration-queue

Slice: 9 — Publication and uncertainty recovery

Task: T9.2

Plan: [Slice 9](../../plan/native-integration-queue.md#slice-9-publication-and-uncertainty-recovery)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Classify definitive not-applied, unknown and Git-applied outcomes with actual refs and method proof instead of command exit alone.

## Acceptance

No lost response triggers another submission or partial ref update.

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
