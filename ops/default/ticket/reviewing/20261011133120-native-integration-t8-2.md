---
id: TKT-20261011133120
title: "T8.2: Reserve and settle optional token ceilings"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-8
source: chat
reported_at: 2026-10-11T13:31:20Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T08:57:28Z
reviewed_at: 2026-10-13T10:10:31Z
branch: feat/slice-8-native-integration-limits
commits: ["e1e574d4c031a25452de66da886be0e1df7271a9"]
---

# T8.2: Reserve and settle optional token ceilings

## Delivery Binding

Delivery set: native-integration-queue

Slice: 8 — Persistent limits and decisions

Task: T8.2

Plan: [Slice 8](../../plan/native-integration-queue.md#slice-8-persistent-limits-and-decisions)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Implement checked invocation-wide reservations across all roles with certified whole-call harness enforcement and compatible usage accounting.

## Acceptance

Concurrent calls cannot oversubscribe; output-only metering rejects finite-token mode and unknown usage stays charged.

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

Implemented invocation admission ceilings, certified executor/reviewer launch
paths, sealed attempt-owned reservations, compatible merger accounting units,
and trusted settlement into the existing invocation ledger. Uncertain usage
keeps the full grant across restart; raw commands cannot bypass finite mode.

Development regression and race checks passed, as recorded in the
[Slice 8 report](../../report/slices/native-integration-queue/slice-8-integration-limits.md).
Strict lint passed with zero issues. Corrected-source pre-commit `make verify`
passed with all 24 changed Go files unchanged; core 2182.690 s, app 125.596 s
and every other package passed. Late-ceiling and process-permission regressions
passed, and focused race checks passed after the correction. Owned check
services finished naturally.

Implementation: [e1e574d4c031a25452de66da886be0e1df7271a9](https://forge.adrianpk.com/adrianpk/chr/commit/e1e574d4c031a25452de66da886be0e1df7271a9).
Acta commit activity: `01M4TBXAE0MJ2MGP0TE9CH7CXF`.
The Slice 8 PR and verified integration remain pending.
