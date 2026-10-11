---
id: TKT-20261011133111
title: "T6.1: Define versioned merger input and result"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-6
source: chat
reported_at: 2026-10-11T13:31:11Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T6.1: Define versioned merger input and result

## Delivery Binding

Delivery set: native-integration-queue

Slice: 6 — Bounded merger role

Task: T6.1

Plan: [Slice 6](../../plan/native-integration-queue.md#slice-6-bounded-merger-role)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Bind intent, target changes, conflict evidence, scope, replay checkpoint and finite allowances to one immutable role input.

## Acceptance

Malformed, oversized or stale results and escaped evidence paths are rejected before candidate acceptance.

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
