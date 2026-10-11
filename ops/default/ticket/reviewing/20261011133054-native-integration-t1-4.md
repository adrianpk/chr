---
id: TKT-20261011133054
title: "T1.4: Seal exact source/base/result publication proof"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: infra
tags: native-integration, slice-1
source: chat
reported_at: 2026-10-11T13:30:54Z
ready_at: 2026-10-11T13:31:37Z
reviewed_at: 2026-10-11T14:02:50Z
started_at: 2026-10-11T14:00:01Z
branch: feat/slice-1-native-git-transport
pr: 375
commits: [cbd4fb8a6e9ab0a1471cb09db7f4a78d9b3d304d]
---

# T1.4: Seal exact source/base/result publication proof

## Delivery Binding

Delivery set: native-integration-queue

Slice: 1 — Guarded Git transport

Task: T1.4

Plan: [Slice 1](../../plan/native-integration-queue.md#slice-1-guarded-git-transport)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Bind the port to immutable prepared candidate identity, fast-forward ancestry, explicit expected refs, original/published source distinction and no-op defenses.

## Acceptance

P equal to S, an already-P advertisement, fork sources, invalid refs and non-fast-forward target updates cannot be partially published.

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

feat(workflow): bind publication to exact ownership and history proof
Focused check passed: `go test ./internal/core -run 'TestWorkflowGit(Atomic|Candidate|AlreadyPrepared|FastForward)$' -count=1`.
The package linter passed with zero issues. Canonical PR: [#375](https://forge.adrianpk.com/adrianpk/chr/pulls/375). Verified integration remains pending the controller.
