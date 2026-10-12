---
id: TKT-20261011133102
title: "T3.4: Rebuild indexes and validate migrations"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-3
source: chat
reported_at: 2026-10-11T13:31:02Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-11T23:45:04Z
resolution: fixed
closed_at: 2026-10-12T02:42:44Z
branch: feat/slice-3-native-integration-records
pr: 377
reviewed_at: 2026-10-12T00:32:59Z
commits: [c7ee88bb92a27c520489effb717cecde34f2dbdf]
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

Implemented bounded authoritative queue/submission/barrier/capacity/closure queries, disposable SQLite queue projection and explicit revision-guarded additive migration. Cross-record validation rejects malformed or unsupported source records, stale generation or ledger bindings and inconsistent publication/receipt proof without replacing source bytes or creating execution state.

Successful task checks:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeQueue|^TestNativeBudget|^TestNativeSubmission|^TestNativeEntry|^TestNativeRecord|^TestNativeReceipt|^TestWorkflowIndex|^TestWorkflowCacheDeletion|^TestWorkflowSubmissionExclusion' -count=1`
- `go test ./internal/core ./internal/app -run '^TestMaintenanceDBStatus|^TestMigrationStatus' -count=1`
- `go test ./internal/core -run '^TestWorkflowSubmissionExclusion|^TestNativeQueueCorruption' -count=1`
- `git diff --check`


Task commit: `c7ee88bb92a27c520489effb717cecde34f2dbdf`. Acta commit record: `01M4PREJHWRNP8W33M3X20ZJ31`.

Canonical PR: [#377](https://forge.adrianpk.com/adrianpk/chr/pulls/377). All four ordered Looper checks passed on `ef72c8cbabf06816a93c293161c12d03211979ba`. Canonical Rebase + Fast-forward integration into `dev` was verified at 2026-10-12T02:39:12Z; Acta merge record: `01M4PZMZF2HN2XM577G70T3DYB`.
