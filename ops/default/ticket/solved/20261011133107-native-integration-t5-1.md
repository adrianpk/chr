---
id: TKT-20261011133107
title: "T5.1: Prepare exact merge and squash candidates"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-5
source: chat
reported_at: 2026-10-11T13:31:07Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T07:40:03Z
branch: feat/slice-5-native-method-preparation
pr: 379
reviewed_at: 2026-10-12T08:11:05Z
resolution: fixed
closed_at: 2026-10-12T12:58:41Z
commits: [d6532b26691da0df0002f3d29753a9ae0a52be11]
---

# T5.1: Prepare exact merge and squash candidates

## Delivery Binding

Delivery set: native-integration-queue

Slice: 5 — Method preparation and provenance

Task: T5.1

Plan: [Slice 5](../../plan/native-integration-queue.md#slice-5-method-preparation-and-provenance)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Extend deterministic preparation to preserve B/S parents or the one-parent squash proof with complete source attribution.

## Acceptance

Host-verified parents, combined diff and result tree match the immutable prepared candidate.

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

Implemented complete merge/squash source attribution and exact result provenance. Actual Git objects retain B/S merge parents or the single squash parent B, source authors/dates/message hashes, prepared committer metadata and deterministic combined diff. Host guards reject altered metadata even with recomputed enclosing hashes. Original source evidence and canonical fixture refs remain unchanged; historical proof hashes remain verifiable.

Successful task checks:

- `make verify`
- `go test -race ./internal/core -run '^TestNativeCombined|^TestNativeRecord|^TestWorkflowPreparation|^TestWorkflowIntegration|^TestWorkflowGit' -count=1`
- `git diff --check`

Strict lint reported zero issues. The required pre-commit gate passed with core in 1269.073 seconds, app in 105.618 seconds and all other packages passing. The focused race regression passed in 262.157 seconds. [Slice 5 report](../../report/slices/native-integration-queue/slice-5-method-preparation.md) retains development failures and evidence boundaries. All four Slice 5 tasks are locally validated; the controller records required ordered clean-HEAD candidate checks after PR-reference updates.

Task implementation: [commit](https://forge.adrianpk.com/adrianpk/chr/commit/d6532b26691da0df0002f3d29753a9ae0a52be11). Acta task record: `01M4QJN40VG8QKAKS12Z40FY3T`.

Canonical slice PR: [#379](https://forge.adrianpk.com/adrianpk/chr/pulls/379).

All four ordered controller checks passed on `771ae3d603f34e6c12d16243e258a125e5db4f9b`. Canonical Rebase + Fast-forward integration into `dev` was verified at 2026-10-12T12:53:28Z; Acta merge record: `01M4R2SQDTK2WNACPWQRJ7RAQR`.
