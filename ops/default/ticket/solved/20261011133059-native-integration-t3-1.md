---
id: TKT-20261011133059
title: "T3.1: Extend candidate and receipt schemas"
status: solved
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-3
source: chat
reported_at: 2026-10-11T13:30:59Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-11T21:44:33Z
resolution: fixed
closed_at: 2026-10-12T02:42:44Z
branch: feat/slice-3-native-integration-records
pr: 377
reviewed_at: 2026-10-11T22:10:52Z
commits: [64369b81579c8b4136609383568731067bc7bdfd]
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

Implementation: [task commit](https://forge.adrianpk.com/adrianpk/chr/commit/64369b81579c8b4136609383568731067bc7bdfd).

Implemented versioned native candidate/receipt bindings, owned hashed source/history/capability/validation and transaction/observation references. Historical receipt reads preserve their exact bytes. Unsupported or incomplete native proof blocks creation and reads.


Report: [Slice 3](../../report/slices/native-integration-queue/slice-3-integration-records.md).

Canonical PR: [#377](https://forge.adrianpk.com/adrianpk/chr/pulls/377). All four ordered Looper checks passed on `ef72c8cbabf06816a93c293161c12d03211979ba`. Canonical Rebase + Fast-forward integration into `dev` was verified at 2026-10-12T02:39:12Z; Acta merge record: `01M4PZMZF2HN2XM577G70T3DYB`.
