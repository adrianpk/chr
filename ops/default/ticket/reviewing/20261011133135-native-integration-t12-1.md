---
id: TKT-20261011133135
title: "T12.1: Certify Forgejo merge rebase and squash"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: ops
tags: native-integration, slice-12
source: chat
reported_at: 2026-10-11T13:31:35Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T23:55:10Z
reviewed_at: 2026-10-15T00:53:43Z
branch: test/slice-12-native-provider-certification
commits: []
---

# T12.1: Certify Forgejo merge rebase and squash

## Delivery Binding

Delivery set: native-integration-queue

Slice: 12 — Live provider certification

Task: T12.1

Plan: [Slice 12](../../plan/native-integration-queue.md#slice-12-live-provider-certification)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Run all three exact-result methods against designated disposable Forgejo repositories, including source/target contention and metadata convergence.

## Acceptance

Each case records canonical PR, actual Git history/tree, policy/credential role and compatible receipt proof.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through the delivery controller.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only the delivery controller submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

All nine real Forgejo merge, rebase and squash publication/source-race/target-race
cases passed. Positive cases retain the same original PR observed merged, exact
remote Git history/tree and compatibility with native receipt proof validation.
Races were introduced after advertisement and preserved the other canonical ref.
The actual provider merge-policy field mapping was corrected with a red/green
regression covering enabled, denied, absent and wrong-provider fields.

Required pre-commit `make verify` passed with zero lint issues: core 2747.252 s,
app 197.636 s and all TUI/Web packages passed. The complete live matrix passed
in 51.915 s; focused race checks passed in 10.642 s. Acta
`01M4YGRAHG25NY28CDGH3JQKKN`. The owned unit exited successfully on unchanged
seven-file source `2acd17be65196ad0f760c11bb43ff4a380fe502b8be0150260e6299cdc474f53`.
See the [Slice 12 report](../../report/slices/native-integration-queue/slice-12-provider-certification.md).

Canonical delivery PR and exact-HEAD final validation remain pending.
