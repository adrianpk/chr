---
id: TKT-20261011133134
title: "T11.4: Document and verify profile setup"
status: ready
kind: task
severity: unclassified
priority: normal
scope: api
tags: native-integration, slice-11
source: chat
reported_at: 2026-10-11T13:31:34Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T11.4: Document and verify profile setup

## Delivery Binding

Delivery set: native-integration-queue

Slice: 11 — Shared operation surfaces

Task: T11.4

Plan: [Slice 11](../../plan/native-integration-queue.md#slice-11-shared-operation-surfaces)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add user-facing setup/reference guidance for source rewriting, policy support and optional bounded parallel/token operation.

## Acceptance

Upgrade preserves old configuration, explicit opt-in is checked before work admission and supported boundaries match the executable.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/app ./internal/tui ./internal/webui -count=1`
- `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through Looper.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only Looper submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Implementation, validation, PR and verified integration evidence are pending.
