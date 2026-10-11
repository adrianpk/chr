---
id: TKT-20261011133051
title: "T1.1: Execute an explicit atomic receive-pack transaction"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: infra
tags: native-integration, slice-1
source: chat
reported_at: 2026-10-11T13:30:51Z
ready_at: 2026-10-11T13:31:37Z
reviewed_at: 2026-10-11T13:47:31Z
started_at: 2026-10-11T13:44:35Z
branch: feat/slice-1-native-git-transport
commits: []
---

# T1.1: Execute an explicit atomic receive-pack transaction

## Delivery Binding

Delivery set: native-integration-queue

Slice: 1 — Guarded Git transport

Task: T1.1

Plan: [Slice 1](../../plan/native-integration-queue.md#slice-1-guarded-git-transport)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Implement the publication port with advertisement comparison, both explicit old/new ref commands, atomic negotiation, bounded report-status handling and a positive local receive-pack path.

## Acceptance

A real local repository updates S->P and B->P together; stale advertised values abort before commands and uncertain IO is retained.

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

Implemented explicit two-ref atomic commands, bounded advertisement/report/pack processing and owned local receive-pack.
Focused check passed: `go test ./internal/core -run 'TestWorkflowGit(Atomic|Advertisement)$' -count=1`.
The package linter passed with zero issues. PR and integration remain pending.
