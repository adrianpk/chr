---
id: TKT-20261011133052
title: "T1.2: Support host-authenticated SSH publication"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: infra
tags: native-integration, slice-1
source: chat
reported_at: 2026-10-11T13:30:52Z
ready_at: 2026-10-11T13:31:37Z
reviewed_at: 2026-10-11T13:52:59Z
started_at: 2026-10-11T13:49:06Z
branch: feat/slice-1-native-git-transport
commits: [6df8a1f73dbc0f54d80ee15c7697ecd61068ced2]
---

# T1.2: Support host-authenticated SSH publication

## Delivery Binding

Delivery set: native-integration-queue

Slice: 1 — Guarded Git transport

Task: T1.2

Plan: [Slice 1](../../plan/native-integration-queue.md#slice-1-guarded-git-transport)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Connect the guarded exchange through SSH using host-owned authentication and argument vectors, with subprocess ownership and bounded IO.

## Acceptance

The same positive and stale-ref cases pass over a real SSH fixture without credentials entering evidence.

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

feat(workflow): publish guarded refs through host-authenticated SSH
Focused check passed: `go test ./internal/core -run 'TestWorkflow(GitSSH|SSHArguments)$' -count=1`.
The package linter passed with zero issues. PR and integration remain pending.
