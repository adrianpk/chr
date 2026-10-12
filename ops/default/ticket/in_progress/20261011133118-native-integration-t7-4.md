---
id: TKT-20261011133118
title: "T7.4: Handle external target and source movement"
status: in_progress
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-7
source: chat
reported_at: 2026-10-11T13:31:18Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T22:18:48Z
branch: feat/slice-7-native-candidate-review
commits: []
---

# T7.4: Handle external target and source movement

## Delivery Binding

Delivery set: native-integration-queue

Slice: 7 — Combined validation and review

Task: T7.4

Plan: [Slice 7](../../plan/native-integration-queue.md#slice-7-combined-validation-and-review)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Reprepare on a changed target, verify changed source ownership/scope and release publication exclusion during repair.

## Acceptance

Serial and parallel runs tolerate external writers without reusing stale combined evidence or silently expanding scope.

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

Implementation, validation, PR and verified integration evidence are pending.
