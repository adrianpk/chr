---
id: TKT-20261011133135
title: "T12.1: Certify Forgejo merge rebase and squash"
status: ready
kind: task
severity: unclassified
priority: normal
scope: ops
tags: native-integration, slice-12
source: chat
reported_at: 2026-10-11T13:31:35Z
ready_at: 2026-10-11T13:31:37Z
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

Approved: 2026-10-11; the maintainer authorized this corrective set through Looper.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only Looper submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Implementation, validation, PR and verified integration evidence are pending.
