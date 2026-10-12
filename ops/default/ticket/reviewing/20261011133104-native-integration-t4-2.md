---
id: TKT-20261011133104
title: "T4.2: Share role capacity and prevent starvation"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-4
source: chat
reported_at: 2026-10-11T13:31:04Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T03:34:55Z
branch: feat/slice-4-native-queue-scheduling
pr: pending
reviewed_at: 2026-10-12T04:23:09Z
commits: []
---

# T4.2: Share role capacity and prevent starvation

## Delivery Binding

Delivery set: native-integration-queue

Slice: 4 — Canonical queue scheduling

Task: T4.2

Plan: [Slice 4](../../plan/native-integration-queue.md#slice-4-canonical-queue-scheduling)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Route executor, merger, reviewer and closing launches through the same repository pool; bound prepared candidates and prefer finishable integration.

## Acceptance

Default mode runs one agent and sustained implementation cannot starve eligible integration.

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

Implemented and validated on the recorded Slice 4 branch. Eligible validation, preparation and review receive shared repository capacity ahead of new implementation. Existing role launches and merger-purpose reservations use the same pool. Prepared-target capacity defaults to two, remains retained during pause/block/restart, and is released through guarded terminal transitions. Human approval and full/unknown target barriers do not retain runnable integration priority. Exact task commit and Acta reference follow this code commit; the canonical slice PR remains pending T4.3 and T4.4.

Successful task checks:

- `make verify`
- `go test -race ./internal/core ./internal/app -run '^TestNativeRole|^TestNativePrepared|^TestNativeSchedule|^TestNativeEntry|^TestNativeQueue|^TestNativeSubmission|^TestNativeBudget|^TestNativeProfile|^TestNativeLegacy|^TestWorkflowCapacity|^TestWorkflowPreparation|^TestWorkflowReview|^TestWorkflowConfig|^TestWorkflowUpgrade|^TestInitWorkflow' -count=1`
- `git diff --check`

Strict lint: zero issues. The full normal core suite passed in 1195.940 seconds and app in 139.514 seconds. Expanded focused race checks passed for core in 680.292 seconds and app in 8.638 seconds. The canonical [Slice 4 report](../../report/slices/native-integration-queue/slice-4-queue-scheduling.md) retains corrected development failures and the task validation boundary.
