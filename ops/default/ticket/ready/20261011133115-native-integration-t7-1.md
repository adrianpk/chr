---
id: TKT-20261011133115
title: "T7.1: Run required gates on the exact prepared result"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-7
source: chat
reported_at: 2026-10-11T13:31:15Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T7.1: Run required gates on the exact prepared result

## Delivery Binding

Delivery set: native-integration-queue

Slice: 7 — Combined validation and review

Task: T7.1

Plan: [Slice 7](../../plan/native-integration-queue.md#slice-7-combined-validation-and-review)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Execute repository-owned combined validation on clean P and retain result-tree/history-bound evidence independently of agent reports.

## Acceptance

A textually clean semantic regression fails the combined gate instead of being published.

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
