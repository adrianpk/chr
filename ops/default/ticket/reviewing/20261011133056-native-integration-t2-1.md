---
id: TKT-20261011133056
title: "T2.1: Admit an explicit owned-source publication profile"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-2
source: chat
reported_at: 2026-10-11T13:30:56Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-11T19:01:24Z
reviewed_at: 2026-10-11T19:39:13Z
commits: [131dd1733a087585add492b68ba5fa3177ef3ec7]
---

# T2.1: Admit an explicit owned-source publication profile

## Delivery Binding

Delivery set: native-integration-queue

Slice: 2 — Forge publication profiles

Task: T2.1

Plan: [Slice 2](../../plan/native-integration-queue.md#slice-2-forge-publication-profiles)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add opt-in profile configuration, current capability/ownership/policy fingerprints, unsupported reasons and original/published source-aware binding.

## Acceptance

Upgrade preserves policy and keeps the profile disabled; incompatible protections, native approval and fork ownership block before publication.

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

Implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/131dd1733a087585add492b68ba5fa3177ef3ec7).

Focused configuration/upgrade/native checks, real default rebase on both designated forges and the complete pre-commit `make verify` passed.

Report: [Slice 2](../../report/slices/native-integration-queue/slice-2-forge-profiles.md).

Canonical PR: pending. Final ordered checks are recorded against the clean PR head by Looper; verified integration is pending.
