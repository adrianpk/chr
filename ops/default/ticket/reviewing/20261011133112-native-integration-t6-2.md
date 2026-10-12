---
id: TKT-20261011133112
title: "T6.2: Execute the merger in an owned checkout"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: domain
tags: native-integration, slice-6
source: chat
reported_at: 2026-10-11T13:31:12Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-12T13:55:30Z
reviewed_at: 2026-10-12T15:28:14Z
branch: feat/slice-6-native-merger-role
pr: 380
commits: [78793f6ce58fc5fabedcece6e9338698713d8739]
---

# T6.2: Execute the merger in an owned checkout

## Delivery Binding

Delivery set: native-integration-queue

Slice: 6 — Bounded merger role

Task: T6.2

Plan: [Slice 6](../../plan/native-integration-queue.md#slice-6-bounded-merger-role)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Run the configured harness only for actionable conflicts/regressions with shared capacity, local tool permissions and host-owned lifecycle.

## Acceptance

The role cannot push, merge, approve, change scope or forge receipts, and timeout preserves ownership of a live child.

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

Implemented owned merger execution through a certified JSON-line model-tool broker.
The driver runs in a private Git checkout and uses the shared capacity pool,
process birth identity and persistent attempt/token ledger. Scoped file tools
reject control paths and filesystem aliases; shell and local validation processes
are denied. Proposals retain evidence without candidate/publication authority.

Focused subprocess/Git checks passed:

- `go test ./internal/core -run '^TestNativeMerger' -count=1` — 24.264 seconds.
- `go test -race ./internal/core -run '^TestNativeMerger' -count=1` — 39.962 seconds.
- Strict lint — zero issues.

Required pre-commit `make verify` passed: zero lint issues; core 1252.301 seconds,
app 95.194 seconds and all other packages passed. `git diff --check` passed.
Task commit: [78793f6ce58fc5fabedcece6e9338698713d8739](https://forge.adrianpk.com/adrianpk/chr/commit/78793f6ce58fc5fabedcece6e9338698713d8739).
Acta task record: `01M4RBPR0K3FB8PTX766B0K7ZH`.
Canonical PR and verified integration evidence remain pending.

The driver is a trusted host component with built-in tools disabled, not an
operating-system sandbox. Fixtures use a certified test driver, not a live model.
Semantic repair and host diff acceptance remain owned by T6.3 and T6.4.

Canonical PR: [#380](https://forge.adrianpk.com/adrianpk/chr/pulls/380). Exact-HEAD candidate validation is
pending in the controller ledger; verified integration remains pending.
