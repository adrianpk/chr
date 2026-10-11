---
id: TKT-20261011133105
title: "T4.3: Normalize target and coordinator identity"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-4
source: chat
reported_at: 2026-10-11T13:31:05Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T4.3: Normalize target and coordinator identity

## Delivery Binding

Delivery set: native-integration-queue

Slice: 4 — Canonical queue scheduling

Task: T4.3

Plan: [Slice 4](../../plan/native-integration-queue.md#slice-4-canonical-queue-scheduling)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Resolve aliases/worktrees/namespaces to common repository/target keys and enforce registered coordination across clones/hosts.

## Acceptance

Aliases cannot bypass capacity or locks; unsupported coordinator connectivity blocks before mutable work.

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
