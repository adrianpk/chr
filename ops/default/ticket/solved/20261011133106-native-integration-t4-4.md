---
id: TKT-20261011133106
title: "T4.4: Verify real concurrent scheduling"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-4
source: chat
reported_at: 2026-10-11T13:31:06Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T05:55:32Z
resolution: fixed
closed_at: 2026-10-12T07:35:05Z
branch: feat/slice-4-native-queue-scheduling
pr: 378
reviewed_at: 2026-10-12T06:05:37Z
commits: [c08b6d6590342a45dcb21fce12d05b740f98ef6b]
---

# T4.4: Verify real concurrent scheduling

## Delivery Binding

Delivery set: native-integration-queue

Slice: 4 — Canonical queue scheduling

Task: T4.4

Plan: [Slice 4](../../plan/native-integration-queue.md#slice-4-canonical-queue-scheduling)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Exercise subprocess overlap, per-target exclusion, independent targets, dynamic capacity and unresolved publication barriers.

## Acceptance

Two independent lanes overlap within bounds; one uncertain target does not block unrelated targets or read-only status.

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

Implemented and validated on the recorded Slice 4 branch. Actual native-selector subprocesses prove default serial execution, overlap at capacity two, isolated checkouts, third-lane and legacy exclusion, and reduced capacity retaining live owners until natural completion. Every worker commits source and passes its host task gate. Independent existing publication owners overlap and advance only their own actual targets; a retained unknown native dispatch permits unrelated native execution and status without changing its journal or owning records.

Successful task checks:

- `make lint-strict`
- `go test -race ./internal/core -run '^TestNativeProcess|^TestWorkflowParallelExecution$|^TestWorkflowParallelBlocker$|^TestWorkflowParallelPublication$|^TestNativeTargetBarrier$|^TestNativeSubmissionUnknown$|^TestNativePreparedBound$' -count=1`
- `go test ./internal/core -run '^TestNativeProcessTargets$' -count=1`
- `git diff --check`

Strict lint: zero issues. Focused race regression: 132.880 seconds. Final target fixture normal check: 6.538 seconds. Implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/c08b6d6590342a45dcb21fce12d05b740f98ef6b). Acta: `01M4QBFSQJAKPEM2CEPRJVN2BR`. Required ordered clean-HEAD candidate evidence is recorded through the controller after PR-reference commits.

Canonical slice PR: [#378](https://forge.adrianpk.com/adrianpk/chr/pulls/378).

All four ordered controller checks passed on `bb3a88888a47040f4ded39807f79623bb86d8ab7`. Canonical Rebase + Fast-forward integration into `dev` was verified at 2026-10-12T07:28:55Z; Acta merge record: `01M4QG7EMHKEPQB0G8SHBCVD1T`.
