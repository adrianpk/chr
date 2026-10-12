---
id: TKT-20261011133110
title: "T5.4: Verify preparation isolation and method proofs"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-5
source: chat
reported_at: 2026-10-11T13:31:10Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T11:02:54Z
reviewed_at: 2026-10-12T11:29:36Z
branch: feat/slice-5-native-method-preparation
pr: 379
commits: [a6a4d6ec83d8dc15f4bdf05712417567dc6145fc]
---

# T5.4: Verify preparation isolation and method proofs

## Delivery Binding

Delivery set: native-integration-queue

Slice: 5 — Method preparation and provenance

Task: T5.4

Plan: [Slice 5](../../plan/native-integration-queue.md#slice-5-method-preparation-and-provenance)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add real repository method/history fixtures and prove that worker lanes and user checkouts are untouched.

## Acceptance

Merge/rebase/squash proofs reject altered trees, mappings and provenance, including no-op edge cases.

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

Real merge/rebase/squash fixtures retain distinct dirty worker and user checkouts:
heads, indexes, staged and unstaged content, untracked original evidence and fetch
metadata remain unchanged. The isolation regression reproduced attachment
rewriting the original worker's `FETCH_HEAD`; the object import now disables
that metadata write while attaching the exact clean prepared lane.

Real altered-result objects with structurally consistent attribution and
recomputed history/candidate hashes fail host verification before canonical ref
updates. Forged source trees, prepared dates, signature hashes and source/base
no-op publication claims are rejected across all configured methods. Existing
ordered replay, parent and empty-replay mapping regressions pass in the same run.

Successful task checks:

- `make verify` — zero lint issues; core 1260.604 seconds, app 104.874 seconds and all other packages passed.
- `go test ./internal/core -run '^TestNativePreparationIsolation$|^TestNativeMethod|^TestNativeCombined|^TestNativeReplay|^TestNativePreparationCollision$|^TestNativeCollisionBudget$' -count=1` — pass, 11.638 seconds.
- `go test -race ./internal/core -run '^TestNativePreparationIsolation$|^TestNativeMethod|^TestNativeCombined|^TestNativeReplay|^TestNativePreparationCollision$|^TestNativeCollisionBudget$' -count=1` — pass, 17.664 seconds.
- `git diff --check` — pass.

Task commit: [a6a4d6ec83d8dc15f4bdf05712417567dc6145fc](https://forge.adrianpk.com/adrianpk/chr/commit/a6a4d6ec83d8dc15f4bdf05712417567dc6145fc).
Acta task record: `01M4QY05BSSYEYMEPHE06JK3VR`.
This task is validated locally. Required ordered clean-HEAD candidate checks use
the controller ledger after PR-reference commits. Integration remains controller-owned.

Canonical slice PR: [#379](https://forge.adrianpk.com/adrianpk/chr/pulls/379).
