---
id: TKT-20261011133109
title: "T5.3: Enforce non-no-op and signing policy"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-5
source: chat
reported_at: 2026-10-11T13:31:09Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T5.3: Enforce non-no-op and signing policy

## Delivery Binding

Delivery set: native-integration-queue

Slice: 5 — Method preparation and provenance

Task: T5.3

Plan: [Slice 5](../../plan/native-integration-queue.md#slice-5-method-preparation-and-provenance)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Reprepare source-equal candidates within the approved bound and sign rewritten commits through the trusted host where required.

## Acceptance

No artificial unrelated commit is inserted; missing certified signing capability blocks before validation/publication.

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
