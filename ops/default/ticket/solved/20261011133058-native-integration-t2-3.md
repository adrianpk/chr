---
id: TKT-20261011133058
title: "T2.3: Certify GitHub default rebase observation"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-2
source: chat
reported_at: 2026-10-11T13:30:58Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-11T19:01:24Z
reviewed_at: 2026-10-11T19:39:13Z
resolution: fixed
closed_at: 2026-10-11T21:42:36Z
branch: feat/slice-2-native-forge-profiles
pr: 376
commits: [a7e90f2ee18ba5d09e46677c101b38f981916b1b]
---

# T2.3: Certify GitHub default rebase observation

## Delivery Binding

Delivery set: native-integration-queue

Slice: 2 — Forge publication profiles

Task: T2.3

Plan: [Slice 2](../../plan/native-integration-queue.md#slice-2-forge-publication-profiles)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Implement the equivalent existing GitHub adapter contract and prove one default rebase in a designated disposable repository.

## Acceptance

The same source/base/result and canonical merged-PR proof holds without depending on a provider merge queue.

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

Implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/a7e90f2ee18ba5d09e46677c101b38f981916b1b).

Focused configuration/upgrade/native checks, real default rebase on both designated forges and the complete pre-commit `make verify` passed.

Report: [Slice 2](../../report/slices/native-integration-queue/slice-2-forge-profiles.md).

Canonical PR: [#376](https://forge.adrianpk.com/adrianpk/chr/pulls/376). All five ordered Looper checks passed on `3129b5ce312df22287a4e7b78e08bf41f5958377`; canonical merge into `dev` was verified on 2026-10-11.
