---
id: TKT-20261011133121
title: "T8.3: Persist operator decisions and bounded grants"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-8
source: chat
reported_at: 2026-10-11T13:31:21Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-13T10:19:42Z
reviewed_at: 2026-10-13T11:43:31Z
branch: feat/slice-8-native-integration-limits
pr: 382
commits: ["cb24ddffc886af7db8fe2f746fdd0d9982bbcdd8"]
---

# T8.3: Persist operator decisions and bounded grants

## Delivery Binding

Delivery set: native-integration-queue

Slice: 8 — Persistent limits and decisions

Task: T8.3

Plan: [Slice 8](../../plan/native-integration-queue.md#slice-8-persistent-limits-and-decisions)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Expose candidate-bound ambiguity resolution and explicit positive budget extensions using existing revision guards.

## Acceptance

Silence does not approve; stale answers fail and grants preserve all prior consumption.

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

Implemented immutable entry-owned questions, explicit guarded answers and
candidate-bound operator evidence for the existing merger. Positive grants work
before queue admission and preserve policy, consumed attempts and unknown usage.
The common operation service delegates these actions to the existing owner.

Development normal/race checks passed for decision, grant, budget and real merger
fixtures; strict lint reports zero issues. Required pre-commit `make verify`
passed with 27 unchanged Go files: core 2136.587 s, app 115.808 s and all other
packages passed. Implementation and development validation are complete; final exact-head slice
validation and verified integration remain pending.

Evidence: [Slice 8 report](../../report/slices/native-integration-queue/slice-8-integration-limits.md).

Implementation: [cb24ddffc886af7db8fe2f746fdd0d9982bbcdd8](https://forge.adrianpk.com/adrianpk/chr/commit/cb24ddffc886af7db8fe2f746fdd0d9982bbcdd8).
Acta commit activity: `01M4TH7SCDF346JDG9Y5J8GKJQ`.

Canonical slice PR: [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382). The controller records ordered exact-HEAD
validation separately before integration; task status remains reviewing.
