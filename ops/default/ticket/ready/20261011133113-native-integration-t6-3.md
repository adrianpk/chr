---
id: TKT-20261011133113
title: "T6.3: Preserve semantic intent during repair"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-6
source: chat
reported_at: 2026-10-11T13:31:13Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T6.3: Preserve semantic intent during repair

## Delivery Binding

Delivery set: native-integration-queue

Slice: 6 — Bounded merger role

Task: T6.3

Plan: [Slice 6](../../plan/native-integration-queue.md#slice-6-bounded-merger-role)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Support per-conflict decisions and correct replay-step repair, declared regeneration, explicit integration fixes and ambiguity requests.

## Acceptance

Conflicting product requirements or unsupported binary/submodule/LFS resolution stop with a concrete decision request.

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
