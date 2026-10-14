---
id: TKT-20261011133132
title: "T11.2: Present CLI and TUI integration status"
status: solved
kind: task
severity: unclassified
priority: normal
scope: api
tags: native-integration, slice-11
source: chat
reported_at: 2026-10-11T13:31:32Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T20:13:34Z
resolution: fixed
closed_at: 2026-10-14T23:54:45Z
reviewed_at: 2026-10-14T21:13:02Z
branch: feat/slice-11-native-integration-surfaces
pr: 385
commits:
  - 18b307141f84534b78d38429db1f79a380417c1a
---

# T11.2: Present CLI and TUI integration status

## Delivery Binding

Delivery set: native-integration-queue

Slice: 11 — Shared operation surfaces

Task: T11.2

Plan: [Slice 11](../../plan/native-integration-queue.md#slice-11-shared-operation-surfaces)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Show serial/parallel lanes, target queue position, active role and actionable blockers through shared service projections.

## Acceptance

Read-only views do not retry, repair or mutate records and a serial user sees one active agent.

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

CLI/TUI text and CLI JSON expose the shared retained integration projection.
The views identify canonical queues, original/prepared identity, actual serial
and parallel roles, bounded accounting, questions and actionable blockers.
Actual prepared-candidate and owned-role reads preserve authoritative bytes.

Focused app regressions passed in 33.812 s and strict lint reports zero issues;
Acta `01M4Y19DFC1ADK6N1T1ZWWYEH6`. The five-file source is frozen at
`3d98c77273f1916314042ea00886b32695cb45b5068c69dbf01cff119069f44a`. Focused race checks passed: app 64.400 s and TUI 1.023 s; Acta
`01M4Y1EB7NM6TFQ16NCK6CEYN3`. Required pre-commit `make verify` passed with zero lint issues, core 2809.193 s,
app 128.702 s and all TUI/Web packages passed; Acta `01M4Y44X498KK6Z58E062WY9MX`.
The owned validation service exited successfully:
`chr-native-slice11-t112-validation-3d98c772.service`. The existing
[Slice 11 report](../../report/slices/native-integration-queue/slice-11-integration-surfaces.md)
records exact commands, fixture corrections and retained process evidence.

Task commit: [18b307141f84534b78d38429db1f79a380417c1a](https://forge.adrianpk.com/adrianpk/chr/commit/18b307141f84534b78d38429db1f79a380417c1a). Acta commit activity: `01M4Y4616JTEX8TRQ2J10X1GKA`.

Canonical PR: [#385](https://forge.adrianpk.com/adrianpk/chr/pulls/385). PR activity: `01M4Y9KVFXPP2RAVFDH4SY62CV`.
Final ordered clean-HEAD validation is recorded through the controller after
PR-reference commits. Verified integration is recorded below.

Verified integration: [PR #385](https://forge.adrianpk.com/adrianpk/chr/pulls/385) merged into `dev` at `12094bddb2d298f6c60fbfcb49f513b6178ff595` on 2026-10-14T23:48:48Z. All four configured checks passed on that exact clean HEAD. Controller merge activity: `01M4YD33D6YKR6AASWSDNZEEV0`; final validation activity: `01M4YD0NT7KDTBNP3ARA2RNMD8`.
