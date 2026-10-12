---
id: TKT-20261011133111
title: "T6.1: Define versioned merger input and result"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-6
source: chat
reported_at: 2026-10-11T13:31:11Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T13:02:05Z
reviewed_at: 2026-10-12T13:41:56Z
branch: feat/slice-6-native-merger-role
pr: 380
commits: [dd376916016bc85c9f20677e03997e0b7f6a2a58]
---

# T6.1: Define versioned merger input and result

## Delivery Binding

Delivery set: native-integration-queue

Slice: 6 — Bounded merger role

Task: T6.1

Plan: [Slice 6](../../plan/native-integration-queue.md#slice-6-bounded-merger-role)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Bind intent, target changes, conflict evidence, scope, replay checkpoint and finite allowances to one immutable role input.

## Acceptance

Malformed, oversized or stale results and escaped evidence paths are rejected before candidate acceptance.

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

Versioned merger input/result contracts and host evidence checks are implemented.
Inputs seal exact entry/unit/candidate/generation, source/base/refs, method, scope,
owned workspace, intent, target changes, replay checkpoint and finite allowances.
Malformed or stale proposals, ambiguous JSON, escaped or altered evidence,
forged approval/publication fields and incomplete intent declarations are rejected.

Successful task checks:

- `make verify` — zero lint issues; core 1282.104 seconds, app 94.013 seconds and all other packages passed.
- `go test ./internal/core -run '^TestNativeMerger' -count=1` — pass, 0.132 seconds.
- `go test -race ./internal/core -run '^TestNativeMerger' -count=1` — pass, 1.263 seconds.
- `go test ./internal/core -run '^$' -fuzz '^FuzzNativeMergerJSON$' -fuzztime=10s -parallel=2` — pass, 37022 executions, 11.044 seconds.
- `git diff --check` — pass.

The protocol validates proposals without launching an agent or granting candidate
approval. T6.2–T6.4 remain pending before the canonical slice PR. Required ordered
clean-HEAD candidate checks will use the controller ledger after PR-reference
commits are pushed. Owned test fixtures completed naturally and were cleaned
through the test lifecycle.

Task commit: [dd376916016bc85c9f20677e03997e0b7f6a2a58](https://forge.adrianpk.com/adrianpk/chr/commit/dd376916016bc85c9f20677e03997e0b7f6a2a58).
Acta task record: `01M4R5KBYTW86RNNWE19EJ9ZJJ`.

Canonical PR: [#380](https://forge.adrianpk.com/adrianpk/chr/pulls/380). Exact-HEAD candidate validation is
pending in the controller ledger; verified integration remains pending.
