---
id: TKT-20261019103832
title: "TUI harness selection"
status: in_progress
kind: task
severity: medium
priority: normal
scope: ui
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:32Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T17:19:15Z
branch: feat/slice-6-tui-harness-selection
commits: []
---

# TUI harness selection

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 6 of 8 — TUI harness selection.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `feat/slice-6-tui-harness-selection`.

Expected PR title: `feat(slice-6): change project harness selections from the TUI`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-6-tui-harness-selection.md`.

Activated after canonical PR #416 was verified merged into dev. T6.1 is
the first recorded task; the canonical report accompanies its implementation.

Prerequisite: Slice 5 integrated into dev; implementation authorization for this set.

## Required Outcome

Local users can inspect and change the same per-project role defaults from the TUI without restarting the TUI or service.

Owners: Existing TUI project controls and thin application adapters.

Acceptance: HP-19, HP-20, HP-21.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T6.1 | Add the project-context role/profile picker with readable current defaults, readiness and separately retained active bindings; reuse the central operation service and local actor mapping. |
| T6.2 | Handle keyboard selection, clear, cancel and stale/conflicting saves without losing the visible choice; verify that Web/CLI changes become visible on the next refresh/admission and no TUI-owned configuration store exists. |

## Validation

Exercise keyboard/event-driven update behavior through the real TUI adapter with temporary workspace configuration, cancellation, stale revision and independent process changes; retain existing compact layout behavior.

Focused regression (planned): `go test ./internal/tui ./internal/app -run '^TestTUIHarnessSelection' -count=1 -timeout=3m`.

Before slice review/integration, run `make verify` and `git diff --check`.
These commands are planned checks, not recorded successful executions.
Retain commits, PR, validation and blocker evidence here and in the slice report;
the tracker remains an order map.
