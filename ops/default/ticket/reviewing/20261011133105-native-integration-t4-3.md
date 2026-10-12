---
id: TKT-20261011133105
title: "T4.3: Normalize target and coordinator identity"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-4
source: chat
reported_at: 2026-10-11T13:31:05Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T04:35:19Z
branch: feat/slice-4-native-queue-scheduling
pr: pending
reviewed_at: 2026-10-12T05:44:40Z
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

Approved: 2026-10-11; the maintainer authorized this corrective set through the controller.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only the controller submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Implemented and validated on the recorded Slice 4 branch. Safe public repository URL spellings resolve to the same target key across Git remote names, namespaces and linked checkouts. Registered coordinator ownership is checked before domain mutation or transition replay; copied and disconnected authority cannot reuse a native journal. Consistent records remain readable after registration loss. Actual native task execution, worker and prepared-target capacity, retained native publication barriers and byte preservation after rejected clone work are verified. The task commit and Acta reference follow in the documentation binding commit; the canonical slice PR remains pending T4.4.

Successful task checks:

- `make verify`
- `go test -race ./internal/core -run '^TestWorkflowClaimForeignClone$|^TestNativeRepositoryAliases$|^TestNativeTarget|^TestNativeCoordinator|^TestWorkflowCoordinator|^TestWorkflowEntryRepository$' -count=1`
- `go test -race ./internal/core ./internal/app -run '^TestNativeRepository|^TestNativeTarget|^TestNativeCoordinator|^TestNativeSchedule|^TestNativeRole|^TestNativePrepared|^TestNativeEntry|^TestNativeQueue|^TestNativeSubmission|^TestWorkflowCoordinator|^TestWorkflowCapacity|^TestWorkflowEntry|^TestWorkflowStore|^TestWorkflowTransition|^TestWorkflowAdopt|^TestWorkflowUpgrade|^TestWorkflowMaterial' -count=1`
- `git diff --check`

Strict lint: zero issues. The final full normal core suite passed in 1279.578 seconds and app in 113.484 seconds, with all other packages passing. The focused race regression passed in 24.308 seconds; expanded race checks passed for core in 276.815 seconds and app in 3.579 seconds. The canonical [Slice 4 report](../../report/slices/native-integration-queue/slice-4-queue-scheduling.md) retains corrected development failures and the task validation boundary.
