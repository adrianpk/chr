---
id: TKT-20261011133123
title: "T9.1: Fence and journal authorized dispatch"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-9
source: chat
reported_at: 2026-10-11T13:31:23Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T18:18:51Z
reviewed_at: 2026-10-13T19:13:48Z
branch: feat/slice-9-native-publication-recovery
commits: ["e4531667cbbeb82922d79500cb7bd79c72e2efb9"]
---

# T9.1: Fence and journal authorized dispatch

## Delivery Binding

Delivery set: native-integration-queue

Slice: 9 — Publication and uncertainty recovery

Task: T9.1

Plan: [Slice 9](../../plan/native-integration-queue.md#slice-9-publication-and-uncertainty-recovery)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Acquire canonical target exclusion, re-observe current candidate/policy, persist intent and dispatch identity before external IO.

## Acceptance

Stale callbacks cannot publish and an earlier potentially live dispatcher prevents replacement publication.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through the delivery controller.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only the delivery controller submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

The host dispatch boundary is implemented with durable intent and dispatcher
identity before actual atomic Git IO. Stale owner and policy checks refuse IO;
restart and overlapping live dispatch preserve the same exclusion and budget.

Focused normal and race checks passed in 254.104 s and 381.919 s; strict lint
reports zero issues. The required pre-commit `make verify` check passed with seven unchanged Go
files: core 2243.062 s, app 108.440 s and all other packages passed. Canonical
PR and verified integration remain pending.

Evidence: [Slice 9 report](../../report/slices/native-integration-queue/slice-9-publication-recovery.md).

Implementation: [e4531667cbbeb82922d79500cb7bd79c72e2efb9](https://forge.adrianpk.com/adrianpk/chr/commit/e4531667cbbeb82922d79500cb7bd79c72e2efb9).
Acta commit activity: `01M4VAYVXVK7AMTY6J2RA7TYC1`.
