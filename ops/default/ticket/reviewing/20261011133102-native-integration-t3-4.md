---
id: TKT-20261011133102
title: "T3.4: Rebuild indexes and validate migrations"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-3
source: chat
reported_at: 2026-10-11T13:31:02Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-11T23:45:04Z
branch: feat/slice-3-native-integration-records
reviewed_at: 2026-10-12T00:32:59Z
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

Implemented bounded authoritative queue/submission/barrier/capacity/closure queries, disposable SQLite queue projection and explicit revision-guarded additive migration. Cross-record validation rejects malformed or unsupported source records, stale generation or ledger bindings and inconsistent publication/receipt proof without replacing source bytes or creating execution state.

Successful task checks:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeQueue|^TestNativeBudget|^TestNativeSubmission|^TestNativeEntry|^TestNativeRecord|^TestNativeReceipt|^TestWorkflowIndex|^TestWorkflowCacheDeletion|^TestWorkflowSubmissionExclusion' -count=1`
- `go test ./internal/core ./internal/app -run '^TestMaintenanceDBStatus|^TestMigrationStatus' -count=1`
- `go test ./internal/core -run '^TestWorkflowSubmissionExclusion|^TestNativeQueueCorruption' -count=1`
- `git diff --check`

The required pre-commit gate passed with zero strict-lint issues, core tests in 975.617 seconds and app tests in 103.692 seconds. The corrected focused race check passed in 132.700 seconds. [Slice 3 report](../../report/slices/native-integration-queue/slice-3-integration-records.md) retains the earlier alias-scope and migration-count failures and their passing corrections. Canonical PR preparation and the controller's ordered exact-HEAD checks follow the task/reference commits. Verified integration remains pending.
