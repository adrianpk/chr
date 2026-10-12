---
id: TKT-20261011133114
title: "T6.4: Verify host acceptance of merger changes"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-6
source: chat
reported_at: 2026-10-11T13:31:14Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T16:43:20Z
branch: feat/slice-6-native-merger-role
reviewed_at: 2026-10-12T17:40:46Z
commits: []
---

# T6.4: Verify host acceptance of merger changes

## Delivery Binding

Delivery set: native-integration-queue

Slice: 6 — Bounded merger role

Task: T6.4

Plan: [Slice 6](../../plan/native-integration-queue.md#slice-6-bounded-merger-role)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Compare actual diffs and method proof, commit through the host port and retain attributed evidence of accepted/rejected resolutions.

## Acceptance

Blanket side selection, weakened requirements and stale or out-of-scope changes cannot yield authorization.

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

Implemented generation-bound retained checkpoints, independently reconstructed
Git diffs, owned declared regeneration, exact host commits and accepted/rejected
resolution evidence. Whole-side copies, altered scope/gates, aliases, unsupported
file edits, stale checkpoint/target callbacks and failed procedures cannot create
publication authorization. A bounded host wait confirms merger group quiescence.

Successful focused checks:

- `go test ./internal/core -run '^TestNative(Merger|Combined|Replay)' -count=1` — pass, 122.094 seconds.
- `go test -race ./internal/core -run '^TestNative(Merger|Combined|Replay)' -count=1` — pass, 199.013 seconds.
- `go test ./internal/core -run '^TestNativeMergerDescendants$' -count=3` — pass, 18.082 seconds.
- `make verify` — pass; zero lint issues, core 1328.210 seconds, app 93.104 seconds and all other packages passed.
- `git diff --check` — pass.

Task commit, Acta and canonical PR references follow after commit creation.
The complete candidate still requires ordered exact-HEAD controller validation.
Verified integration is pending the external controller.
