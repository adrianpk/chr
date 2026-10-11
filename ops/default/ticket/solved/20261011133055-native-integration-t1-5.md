---
id: TKT-20261011133055
title: "T1.5: Verify protocol races and review the transport boundary"
status: solved
kind: task
severity: unclassified
priority: normal
scope: infra
tags: native-integration, slice-1
source: chat
reported_at: 2026-10-11T13:30:55Z
ready_at: 2026-10-11T13:31:37Z
reviewed_at: 2026-10-11T14:09:27Z
started_at: 2026-10-11T14:05:06Z
resolution: fixed
closed_at: 2026-10-11T18:20:54Z
branch: feat/slice-1-native-git-transport
pr: 375
commits: [b7790de38a5473f3f0f3c6e5187ececb7f11089d]
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

test(workflow): certify guarded transport contention and lost responses
Focused check passed: `go test ./internal/core -run 'TestWorkflowGit' -count=1`.
The package linter passed with zero issues. Canonical PR: [#375](https://forge.adrianpk.com/adrianpk/chr/pulls/375). Canonical integration verified at `79dfe32380dae9058bde15ba50b8f30c30cbaf3d` on `dev`; the controller merged PR #375 on 2026-10-11.
