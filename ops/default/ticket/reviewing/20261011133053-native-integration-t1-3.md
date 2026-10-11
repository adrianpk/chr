---
id: TKT-20261011133053
title: "T1.3: Support host-authenticated HTTPS publication"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: infra
tags: native-integration, slice-1
source: chat
reported_at: 2026-10-11T13:30:53Z
ready_at: 2026-10-11T13:31:37Z
reviewed_at: 2026-10-11T13:58:41Z
started_at: 2026-10-11T13:54:15Z
branch: feat/slice-1-native-git-transport
commits: []
---

# T1.3: Support host-authenticated HTTPS publication

## Delivery Binding

Delivery set: native-integration-queue

Slice: 1 — Guarded Git transport

Task: T1.3

Plan: [Slice 1](../../plan/native-integration-queue.md#slice-1-guarded-git-transport)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Connect the same exchange through smart HTTPS with host credential integration, canonical identity checks and bounded packet/object processing.

## Acceptance

The positive and stale-ref cases pass over a real HTTP receive-pack fixture; redirects and partial responses cannot become success.

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

feat(workflow): publish guarded refs through smart HTTPS
Focused check passed: `go test ./internal/core -run 'TestWorkflow(GitHTTPS|HTTPIdentity)$' -count=1`.
The package linter passed with zero issues. PR and integration remain pending.
