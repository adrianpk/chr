# Slice 5: Guided Telegram Journey

Status: drafting
Delivery set: guided-use-case-workflow
Plan: `ops/default/plan/guided-use-case-workflow.md`
Tracker: `ops/default/tracker/guided-use-case-workflow.md`
Branch: `feat/slice-5-guided-telegram`
PR: pending
Initial routing/capture commit: `77b7f637587a63026ce914368939e7f0ed7158e7`

## Purpose

Map Telegram onto the same retained guided use case delivered by the Web
checkpoint. The first T5.1 increment implements canonical routing, explicit
natural-language start, name and capture replies. T5.1 remains in progress;
buttons and the remaining journey are not yet complete.

Plan: [Approved slice map](../../../plan/guided-use-case-workflow.md).
Tracker: [Tasks and evidence](../../../tracker/guided-use-case-workflow.md).

## Delivered Behavior

- Explicit natural-language development starts create one canonical guided
  session through the shared service. Incidental or negated development text
  cannot grant progression authority through this deterministic entry.
- Telegram retains a peer-scoped association with the canonical session and
  freezes each complete shared-owner input before application. Name and capture
  replies use the displayed authoring revision and existing reply/focus guards.
- Repeated input does not duplicate capture. Replaced prompts, foreign peers and
  changed registered roots cannot replace the original binding or canonical input.
- The additive nullable `guided_run_id` column leaves existing manual rows and
  their authoring eligibility unchanged. Manual authoring/capture commands cannot
  adopt or mutate a guided association.

## Implementation Notes

The receiver stores transport routing and display references in its existing
conversation/request/outbox stores. Authoring state, development intent,
checkpoints and native work remain in the project owners. Conversation
`collecting` denotes transport input eligibility; it is not execution status.

The application resolves the current registered project and rechecks its frozen
root binding before entering shared guided operations. Retained requests bind
actor, namespace, canonical run, original source and payload hash. Acknowledging
input, replacing the prompt binding and queueing its reply share one receiver
transaction after the owner result.

## Contracts Added or Changed

- Optional `telegram_conversations.guided_run_id` identifies a canonical guided
  session without extending the existing manual intent/phase enums.
- Receiver-owned guided routing is separate from manual authoring eligibility.
- The same existing reply retention/pruning, peer allocation limits, prepared
  request size bounds and canonical authoring CAS remain authoritative.

## Files of Interest

- `internal/core/telegram_guided_binding.go` and `telegram_guided_finish.go`:
  stable association, peer/root checks and atomic result/reply retention.
- `internal/app/telegram_guided_service.go`, `telegram_guided_dispatch.go`,
  `telegram_guided_apply.go` and `telegram_guided_input.go`: shared-operation
  mapping and frozen current-input binding.
- `cmd/chr/assets/migration/sqlite/20261023000029-add_telegram_guided_binding.sql`:
  additive routing association with existing manual-row preservation.
- `internal/core/telegram_guided_test.go` and
  `internal/app/telegram_guided_test.go`: actual SQLite, HTTP and project/Git tests.

## Validation

Focused checks passed for the initial routing/capture increment:

- `go test ./internal/core ./internal/app -run '^TestGuidedUseCaseTelegram' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(GuidedUseCaseTelegram|TelegramConversation|TelegramAuthoring)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(GuidedUseCaseTelegram|TelegramConversation|TelegramAuthoring)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core/... ./internal/app/...` — zero issues.

Initial focused runs exposed the existing constrained intent enum and the
existing expired-reply diagnostic. The additive association preserves that enum;
the stale-reply regression checks the actual rejection and unchanged owner state.
Corrected focused checks pass.

The affected manual integration selection took core 5.236s and app 7.591s
normally, core 119.881s and app 100.355s with race instrumentation. Later task
checks should stay focused on the changed guided behavior unless another manual
owner changes.

The physical pre-migration database preserves its manual topic, capture,
revision and eligibility after upgrade. The receiver fixture uses real HTTP,
SQLite and temporary project/Git stores. It does not establish live Telegram
client acceptance or complete-journey delivery.

## Risks and Follow-ups

T5.1 still needs current-step buttons and complete clarification input routing.
T5.2 owns managed lifecycle wiring, notifications/coalescing, original documents,
controls, results and cross-interface continuation. T5.3 owns the complete
transport/fault journey and Telegram guidance/changelog. No actual user-operated
Telegram client observation has been authorized or performed for this slice.
