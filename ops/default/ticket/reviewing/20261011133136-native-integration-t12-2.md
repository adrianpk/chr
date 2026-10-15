---
id: TKT-20261011133136
title: "T12.2: Certify GitHub merge rebase and squash"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: ops
tags: native-integration, slice-12
source: chat
reported_at: 2026-10-11T13:31:36Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-15T00:54:01Z
reviewed_at: 2026-10-15T01:10:13Z
branch: test/slice-12-native-provider-certification
commits: []
---

# T12.2: Certify GitHub merge rebase and squash

## Delivery Binding

Delivery set: native-integration-queue

Slice: 12 — Live provider certification

Task: T12.2

Plan: [Slice 12](../../plan/native-integration-queue.md#slice-12-live-provider-certification)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Run the equivalent three cases against designated disposable GitHub repositories using the existing adapter.

## Acceptance

Each selected method publishes exact P with the same PR observed merged and no native-queue prerequisite.

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

All nine real GitHub merge/rebase/squash publication and post-advertisement
source/target contention cases passed under the race detector. Positive cases
publish exact P, retain original S, verify actual remote history/tree and observe
the same original PR merged without a provider merge queue. Evidence is retained
for fixture PRs #13-21 and passes native receipt compatibility checks.

Strict lint passed with zero issues. The complete live race command passed in
221.147 s (package) and 222.095 s (command) on unchanged eight-file source
`d4dc984fbd33e29ff8f0d6612c6bfe8bd07073b7b60191d5879affae715692e5`.
Acta `01M4YHQ5TQDSP5NXAZ10HHQVTZ`. The owned unit exited successfully. Its first
run failed before PR creation due to the local tool-manager wrapper; the passing
run selected the same installed GitHub CLI directly in the fixture PATH.

See the [Slice 12 report](../../report/slices/native-integration-queue/slice-12-provider-certification.md).
Canonical delivery PR and exact-HEAD final validation remain pending.
