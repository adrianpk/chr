---
id: TKT-20261011133059
title: "T3.1: Extend candidate and receipt schemas"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-3
source: chat
reported_at: 2026-10-11T13:30:59Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-11T21:44:33Z
branch: feat/slice-3-native-integration-records
reviewed_at: 2026-10-11T22:10:52Z
commits: []
---

# T3.1: Extend candidate and receipt schemas

## Delivery Binding

Delivery set: native-integration-queue

Slice: 3 — Durable integration records

Task: T3.1

Plan: [Slice 3](../../plan/native-integration-queue.md#slice-3-durable-integration-records)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Add original/published source identities, method manifests and versioned capability/transport references without rewriting historical receipts.

## Acceptance

Old schemas remain verifiable and new readers reject missing required proof or unsupported versions.

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

Implemented versioned native candidate/receipt bindings, owned hashed source/history/capability/validation and transaction/observation references. Historical receipt reads preserve their exact bytes. Unsupported or incomplete native proof blocks creation and reads.

The focused preparation/integration/finalization/submission regressions, native record/receipt race checks, pinned strict lint and pre-commit `make verify` passed. Final slice exact-HEAD checks and canonical PR integration remain pending.

Report: [Slice 3](../../report/slices/native-integration-queue/slice-3-integration-records.md).
