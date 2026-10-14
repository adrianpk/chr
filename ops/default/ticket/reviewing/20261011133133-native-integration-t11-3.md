---
id: TKT-20261011133133
title: "T11.3: Present Web daemon and manual controls"
status: reviewing
reviewed_at: 2026-10-14T22:42:31Z
kind: task
severity: unclassified
priority: normal
scope: api
tags: native-integration, slice-11
source: chat
reported_at: 2026-10-11T13:31:33Z
ready_at: 2026-10-11T13:31:37Z
started_at: 2026-10-14T21:13:39Z
branch: feat/slice-11-native-integration-surfaces
commits:
  - e37e275b49c8a16f1349c916be9f5045b4be811f
---

# T11.3: Present Web daemon and manual controls

## Delivery Binding

Delivery set: native-integration-queue

Slice: 11 — Shared operation surfaces

Task: T11.3

Plan: [Slice 11](../../plan/native-integration-queue.md#slice-11-shared-operation-surfaces)

Tracker: [Native Integration Tracker](../../tracker/native-integration-queue.md)

Specification: [Native Integration Queue and Merger](../../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../../spec/stable/native-integration-queue-model.md)

## Work

Map existing Web/daemon/manual entrypoints to the same candidate-bound actions and evidence with stale-action rejection.

## Acceptance

All surfaces agree on phase and blocker; private tokens, raw lease fields and credentials remain hidden.

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

Web, work/daemon commands and manual service callers now submit displayed
invocation, entry, budget, candidate and question guards through the shared
control operation. Explicit grants retain consumption; resolve/defer answers
retain their declared choice and authority without granting publication.
Private claim generations remain host-side.

Actual prepared candidates, sealed questions and operator actions passed the
focused surface regressions: app 107.195 s, TUI 0.007 s and Web handlers 0.112 s.
Strict lint reports zero issues; Acta `01M4Y68YY6QKNVF97KGNP8BEW5`. The 15-file
source and embedded template are frozen at
`ff3a974203742015969c6f47c4e35dfd1809c55d2dd87a211ad532b2238461e7`. Focused race checks passed: app 216.665 s, TUI 1.032 s and Web handlers 1.218 s;
Acta `01M4Y6HYY5E4VEBPX4MXV2CRE7`. Required pre-commit `make verify` passed: zero lint issues, core 2855.905 s, app 198.637 s, TUI 0.073 s and Web 0.094 s; all remaining Web packages passed. Acta `01M4Y9980W7F4RHA53HB2PC1WF`. The owned validation service exited successfully with unchanged source. Evidence is retained under owned unit
`chr-native-slice11-t113-validation-ff3a9742.service`. The existing
[Slice 11 report](../../report/slices/native-integration-queue/slice-11-integration-surfaces.md)
records exact commands, fixture corrections and retained process evidence.

Implementation: [e37e275b49c8a16f1349c916be9f5045b4be811f](https://forge.adrianpk.com/adrianpk/chr/commit/e37e275b49c8a16f1349c916be9f5045b4be811f). Acta commit activity: `01M4Y9ACQZTX6W5SSREHCWNX35`.
Canonical PR, final ordered clean-HEAD checks and verified integration remain pending.
