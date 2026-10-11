---
id: TKT-20261011133055
title: "T1.5: Verify protocol races and review the transport boundary"
status: ready
kind: task
severity: unclassified
priority: normal
scope: infra
tags: native-integration, slice-1
source: chat
reported_at: 2026-10-11T13:30:55Z
ready_at: 2026-10-11T13:31:37Z
commits: []
---

# T1.5: Verify protocol races and review the transport boundary

## Delivery Binding

Delivery set: native-integration-queue

Slice: 1 — Guarded Git transport

Task: T1.5

Plan: [Slice 1](../../plan/native-integration-queue.md#slice-1-guarded-git-transport)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add meaningful real-process fault fixtures for advertisement/receive races, atomic rejection, malformed packets and lost response; record protocol/security review.

## Acceptance

Both refs participate; external source/target writers are preserved and SSH/HTTPS behavior agrees without an unguarded fallback.

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
