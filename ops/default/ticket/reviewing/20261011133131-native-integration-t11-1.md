---
id: TKT-20261011133131
title: "T11.1: Expose shared integration service operations"
status: reviewing
kind: task
severity: unclassified
priority: normal
scope: api
tags: native-integration, slice-11
source: chat
reported_at: 2026-10-11T13:31:31Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T18:42:26Z
reviewed_at: 2026-10-14T20:12:12Z
branch: feat/slice-11-native-integration-surfaces
pr: 385
commits:
  - 98915f97987104ae0cb3bc67ddcf3f8f91dbab0d
---

# T11.1: Expose shared integration service operations

## Delivery Binding

Delivery set: native-integration-queue

Slice: 11 — Shared operation surfaces

Task: T11.1

Plan: [Slice 11](../../plan/native-integration-queue.md#slice-11-shared-operation-surfaces)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Extend the existing application service with observed queue/phase/role/budget facts and revision-bound review/decision/grant actions.

## Acceptance

Surface callers cannot introduce alternative state transitions or optimistic delivery status.

Acceptance IDs and invariants: use the owning slice's explicit coverage map in the approved plan.

## Validation

- `make verify`
- `go test ./internal/app ./internal/tui ./internal/webui -count=1`
- `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`
- `git diff --check`

## Execution Gate

Approved: 2026-10-11; the maintainer authorized this corrective set through the delivery controller.

Execute only when this slice is active and its predecessors and external prerequisites are satisfied.
Keep one task commit and one canonical slice PR; only the delivery controller submits development merges.
Use the existing Markdown ticket lifecycle; this file does not create a product runtime invocation.

## Evidence

Retained native queue, candidate, role, publication and budget observations now
flow through the shared status operation. Displayed action revisions fence
approval, bounded grants and explicit answers; the host retains claim generation.
Read-only decision inspection rejects a pending transition without replaying it.

Focused regressions passed: core 69.682 s and app 33.102 s. Strict lint passed
with zero issues. Acta activity: `01M4XXSWZ91P9JV7JBWDQ20S78`. The 11-file
source is frozen at `8ddc53dce797218763bebd341c36bd8578efd8b5e4e46bfa598a2caf727428ad`.
Broader race passed: core 115.109 s and app 53.292 s, recorded as Acta
`01M4XY9J5W70A20HWQ4D5VXTXM`. Required pre-commit `make verify` passed with zero lint issues, core 2754.086 s,
app 126.105 s and all TUI/Web packages passed; Acta `01M4Y0NX6YKJTA466DP5TVQ8DA`.
The owned validation service exited successfully:
`chr-native-slice11-t111-validation-8ddc53dc.service`. See the existing
[Slice 11 report](../../report/slices/native-integration-queue/slice-11-integration-surfaces.md).

Task commit: [98915f97987104ae0cb3bc67ddcf3f8f91dbab0d](https://forge.adrianpk.com/adrianpk/chr/commit/98915f97987104ae0cb3bc67ddcf3f8f91dbab0d). Acta commit activity: `01M4Y0R3R8REK566SDYW7AMC0M`.

Canonical PR: [#385](https://forge.adrianpk.com/adrianpk/chr/pulls/385). PR activity: `01M4Y9KVFXPP2RAVFDH4SY62CV`.
Final ordered clean-HEAD validation is recorded through the controller after
PR-reference commits. Verified integration remains pending.
