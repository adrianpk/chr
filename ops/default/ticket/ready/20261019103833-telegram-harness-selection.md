---
id: TKT-20261019103833
title: "Telegram harness selection"
status: ready
kind: task
severity: medium
priority: normal
scope: api
tags: harnesses, pi, codex
source: chat
reported_at: 2026-10-19T10:38:33Z
ready_at: 2026-10-19T10:38:33Z
commits: []
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

These are planned identities; no implementation branch, PR or report exists
merely because this ticket is ready.

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
