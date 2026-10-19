---
id: TKT-20261019103833
title: "Telegram harness selection"
status: reviewing
kind: task
severity: medium
priority: normal
scope: api
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:33Z
ready_at: 2026-10-19T10:38:33Z
started_at: 2026-10-19T17:37:13Z
branch: feat/slice-7-telegram-harness-selection
reviewed_at: 2026-10-19T18:01:45Z
commits: [472d576a34f70a94611626d3d2a0c652f42b842f, 92a45627362a4dcb7c7c1b5a27f655f5eca21033]
---

# Telegram harness selection

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 7 of 8 — Telegram harness selection.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `feat/slice-7-telegram-harness-selection`.

Expected PR title: `feat(slice-7): change project harness selections through Telegram`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-7-telegram-harness-selection.md`.

Activated after canonical PR #417 was verified merged into dev. T7.1 is
the first recorded task; the canonical report accompanies its implementation.

Prerequisite: Slice 6 integrated into dev; implementation authorization for this set.

## Required Outcome

An authorized mapped user can inspect and confirm role/profile changes for the selected project through Telegram.

Owners: Existing Telegram deterministic commands, callback/confirmation binding, help and central application operations.

Acceptance: HP-03, HP-19, HP-20, HP-21.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T7.1 | Add a deterministic /harness inspection/selection entrypoint and concise help with readable role/profile/model names and active-versus-new-work semantics; use the selected registered project. |
| T7.2 | Bind confirmation to the mapped actor, exact project and role/profile choices, observed revision and request identity; reject stale, foreign, replayed-with-changed-arguments or unauthorized requests through the existing shared mutation owner. |
| T7.3 | Verify transport redelivery/lost responses and cross-interface visibility while active generation and workers keep their bindings; no model-only instruction may authorize changing configuration or private setup. |

## Validation

Use deterministic Telegram updates/callbacks through the existing receiver and shared service. Cover stale confirmation, changed selected project, actor mismatch and replay without sending messages to the live bot.

Focused regression (planned): `go test ./internal/core ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`.

Before slice review/integration, run `make verify` and `git diff --check`.
These commands are planned checks, not recorded successful executions.
Retain commits, PR, validation and blocker evidence here and in the slice report;
the tracker remains an order map.

## Implementation Progress

T7.1 is implemented: deterministic `/harness` inspection uses the selected
registered project, shared readiness and separate retained worker bindings.
The bounded reply keeps private setup local and distinguishes new-operation
defaults from admitted work. Help includes the strict named-role syntax.

Passed: `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramHelp)' -count=1 -timeout=3m`.

T7.2 is implemented: exact deterministic confirmations bind mapped actor,
project/root, choices, full revision and retained request ID. Current authority
is checked at publication through the shared configuration owner. Stale or
unauthorized actions cannot publish; uncertain owner failures retain their
request identity for recovery. Terminal confirmation contexts release capacity.

Passed: `go test ./internal/core ./internal/app -run '^(TestTelegramHarnessSelection|TestTelegramOperationAction|TestHarnessSelectionChange)' -count=1 -timeout=3m`.

T7.3 is implemented: receiver redelivery, lost owner acknowledgements and lost
Telegram send acknowledgements preserve one request/receipt. Existing Web/TUI
readers and an independent process see the published image; live claims and
nested/resumed generation keep their binding, while new admissions use the new
default. Model-only proposals and private setup fields grant no authority.

Passed: `go test ./internal/core ./internal/app -run '^TestTelegramHarnessSelection' -count=1 -timeout=3m`.

Final clean-HEAD validation is delegated to the controller after PR-reference
follow-up commits. Its required prefix is `make verify`, focused selection and
focused race; affected existing action/configuration regressions follow, then
`git diff --check`. The report lists the exact ordered commands.
