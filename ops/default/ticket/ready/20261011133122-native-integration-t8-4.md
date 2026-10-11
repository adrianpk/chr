---
id: TKT-20261011133122
title: "T8.4: Verify process and budget recovery"
status: ready
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-8
source: chat
reported_at: 2026-10-11T13:31:22Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T8.4: Verify process and budget recovery

## Delivery Binding

Delivery set: native-integration-queue

Slice: 8 — Persistent limits and decisions

Task: T8.4

Plan: [Slice 8](../../plan/native-integration-queue.md#slice-8-persistent-limits-and-decisions)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Exercise launch/timeout/crash/unknown-usage checkpoints and slot retention with real owned subprocesses.

## Acceptance

A surviving child or uncertain usage prevents hidden replacement and refunds, while serial mode remains one active agent.

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
