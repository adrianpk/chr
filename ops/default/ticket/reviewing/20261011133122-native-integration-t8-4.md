---
id: TKT-20261011133122
title: "T8.4: Verify process and budget recovery"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-8
source: chat
reported_at: 2026-10-11T13:31:22Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T11:49:45Z
reviewed_at: 2026-10-13T12:08:00Z
branch: feat/slice-8-native-integration-limits
commits: ["b0574cca109c65d46a00bbbfb0b426f35fe751d1"]
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

Added real owned driver and host-loss fixtures for crash, timeout, missing usage,
process birth before permission, and a surviving launched driver. Restart and
linked worktree aliases retain the serial agent slot until the driver stops;
unknown usage retains the complete token grant and rejects a hidden replacement.

Focused normal and race checks passed in 38.816 s and 74.448 s. Strict lint
reports zero issues. The two changed test files were unchanged throughout those
checks. All owned fixtures and check processes finished. Required exact-head
slice validation, canonical PR and verified integration remain pending.

Evidence: [Slice 8 report](../../report/slices/native-integration-queue/slice-8-integration-limits.md).

Implementation: [b0574cca109c65d46a00bbbfb0b426f35fe751d1](https://forge.adrianpk.com/adrianpk/chr/commit/b0574cca109c65d46a00bbbfb0b426f35fe751d1).
Acta commit activity: `01M4TJKYTBQ2KTKQWSNH7NNCGN`.
