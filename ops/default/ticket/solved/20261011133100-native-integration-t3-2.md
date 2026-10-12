---
id: TKT-20261011133100
title: "T3.2: Persist integration entries through WorkflowStore"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-3
source: chat
reported_at: 2026-10-11T13:31:00Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-11T22:17:47Z
resolution: fixed
closed_at: 2026-10-12T02:42:44Z
branch: feat/slice-3-native-integration-records
pr: 377
reviewed_at: 2026-10-11T22:47:37Z
commits: [9a004ba55c6b72ac69aa0322fba531eebcc257d7]
---

# T3.2: Persist integration entries through WorkflowStore

## Delivery Binding

Delivery set: native-integration-queue

Slice: 3 — Durable integration records

Task: T3.2

Plan: [Slice 3](../../plan/native-integration-queue.md#slice-3-durable-integration-records)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Implement idempotent unit entry identity, queue sequence allocation and guarded entry/invocation transitions with the existing file transaction owner.

## Acceptance

Duplicate admission consumes one entry; interrupted transitions preserve authoritative records and replay safely.

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

Implemented guarded queue admission and phase transitions through the existing workflow transaction. Concurrent retry, cross-namespace ordering, capacity and arithmetic bounds, interrupted replacement/replay, stale callback, candidate ownership, malformed bytes, stop/resume and uncertain publication checks pass.

Successful task checks:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeEntry|^TestNativeRecord|^TestNativeReceipt' -count=1`
- `git diff --check`


Canonical PR: [#377](https://forge.adrianpk.com/adrianpk/chr/pulls/377). All four ordered Looper checks passed on `ef72c8cbabf06816a93c293161c12d03211979ba`. Canonical Rebase + Fast-forward integration into `dev` was verified at 2026-10-12T02:39:12Z; Acta merge record: `01M4PZMZF2HN2XM577G70T3DYB`.
