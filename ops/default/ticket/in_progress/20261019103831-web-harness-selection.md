---
id: TKT-20261019103831
title: "Web harness selection"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:31Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T16:50:50Z
branch: feat/slice-5-web-harness-selection
commits: []
---

# Web harness selection

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 5 of 8 — Web harness selection.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `feat/slice-5-web-harness-selection`.

Expected PR title: `feat(slice-5): change project harness selections from the Web UI`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-5-web-harness-selection.md`.

Activated after canonical PR #415 was verified merged into dev.
T5.1 is the first recorded task. Introduce the canonical report with the first
implementation task commit; no PR is open yet.

Prerequisite: Slice 4 integrated into dev; implementation authorization for this set.

## Required Outcome

An authenticated user can inspect and change both role defaults in a compact project-context control, with clear effects on new work.

Owners: Existing Web projections/actions/templates and the shared application selection operation.

Acceptance: HP-19, HP-20, HP-21.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T5.1 | Add the compact project-context profile chooser using shared readable selection/readiness projections; show role defaults separately from active invocation bindings, support clearing a role and state that changes apply to new operations. |
| T5.2 | Submit exact revision-bound changes through existing authenticated mutation/origin guards; preserve input on rejection and verify login, stale updates, project scope and the absence of a service restart or credential/path editor. |

## Validation

Exercise real HTTP action/authentication boundaries and one focused browser journey on wide and narrow layouts, including a switch while an invocation remains active. Inspect the persisted result from an independent process.

Focused regression (planned): `go test ./internal/webui ./internal/app -run '^TestWebHarnessSelection' -count=1 -timeout=3m`.

Before slice review/integration, run `make verify` and `git diff --check`.
These commands are planned checks, not recorded successful executions.
Retain commits, PR, validation and blocker evidence here and in the slice report;
the tracker remains an order map.
