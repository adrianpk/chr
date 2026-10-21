# Slice 3: Clear Guarded Confirmations and Controls

Status: delivered
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `feat/slice-3-telegram-action-presentation`
PR: [#427](https://forge.adrianpk.com/adrianpk/chr/pulls/427)

## Purpose

Make Telegram confirmations readable while preserving the original operation,
complete scope, authority, expiry and native receipt.

## Delivered Behavior

- Ticket preparation, exact work admission, controls and harness changes show
  human scope and consequences with one complete identified scope attachment.
  Work documents include every selected source and validation contract.
- A missing, oversized, mismatched or incomplete scope cannot expose a new
  action. Full document storage rolls back prompt admission and authority.
- Buttons describe the actual effect while preserving callback tokens and
  choice keys. Caption, file, labels and choices are frozen under one guard.
- Repeated confirmations show the original result. Admission stays distinct
  from execution/integration; pause acceptance stays distinct from quiescence.
- Questions lead with the actual question and bounded choices. Complete retained
  questions, options and candidate bindings remain in an identified JSON export.
- Record results and role selections use readable bounded text. Existing direct
  typed metadata writes retain their owner and guards; no new lifecycle
  transition or follow-up action was added.

## Implementation Notes

PR [#427](https://forge.adrianpk.com/adrianpk/chr/pulls/427) was verified merged
into `dev` through Rebase + Fast-forward on 2026-10-21 at
`a4d0a865e8bb32a3caac65395a28839b573ed369`. Canonical Forgejo, `origin/dev`
and Acta merge event `01M5FRH8SFP1F2QX0PDWA4AJ4B` agree.

T3.1 is `b7a1174a5d09`, T3.2 is `94975b70a20b` and T3.3 is `099b9a965b0b`;
full identities are in the tracker. PR #426 resolved the earlier dependency at
`ff5c7d3838d8b2dea472485358d0842c69059e3f`. This candidate uses the canonical
activation update at `83802941df1c46bc34ab53201d56811a0e353156`.

The receiver's existing transaction creates the action, freezes its presentation
and document, acknowledges the source and seals the prepared-operation guard
with the presentation hash. Existing delivery attempts, plain fallback, original
message/file acknowledgement and reply association remain under the outbox.

The current workflow owner reads at most one complete 128-member selection
under a ten-second budget, checks each original source/revision/contract image,
and bounds combined source/contract bytes to 1 MiB before forming the scope
export. A first-page-only selection is refused. The existing 2 MiB document,
64 MiB database, 900-unit caption and 1,800-unit text limits remain unchanged.

A valid original callback can resolve an uncertain upload, subject to the
same operation checks. Concurrent answers retain one original answer identity.
Already confirmed owner receipts survive transport retry and result rendering.

## Contracts Added or Changed

Migration 27 adds `guard_version` and an optional presentation hash to receiver
actions. Version 1 covers the original operation and immutable presentation;
version 0 retains its original guard algorithm. Unknown versions are refused
before transmission or application. Stored manifests also retain the exact
original choices and their human labels.

Document-backed action application checks complete retained bytes, destination,
configuration, project, namespace and root binding before creating an answer.
Existing source, policy, target, expiry and owner checks still run at application.
Response-only message fields use the shared renderer without changing operation
authority, arguments, consumed choice or receipt identity.

## Files of Interest

- `internal/core/telegram_action_presentation.go`, `telegram_action_claim.go` and `telegram_action_finish.go`
- `internal/app/telegram_confirmation_document.go` and `service_workflow_selection.go`
- `internal/app/telegram_action_presentation.go` and the existing preparation/work/control/harness dispatchers
- `cmd/chr/assets/migration/sqlite/20261021000027-add_telegram_action_presentation.sql`
- `internal/core/telegram_confirmation_recovery_test.go`, `telegram_confirmation_migration_test.go` and `internal/app/telegram_confirmation_recovery_test.go`

## Validation

Looper evidence `RUKJK7R5AERILO54DESZZNK6UW` records these four checks
passed in order on the exact clean, pushed and merged candidate
`a4d0a865e8bb32a3caac65395a28839b573ed369`:

- `make verify`
- `go test ./internal/core ./internal/app -run '^TestTelegramPresentation(Action|Control|Confirmation)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramPresentation(Action|Control|Confirmation)' -count=1 -timeout=3m`
- `git diff --check`

Strict lint reported zero issues; short tests, eight CI script regressions and
build passed. Focused normal core/app tests took 1.426/1.178 seconds; focused
race took 32.121/12.513 seconds within the unchanged three-minute package budget.
The canonical PR contract passed after the PR-reference follow-up was pushed.

Affected legacy action, control, harness, preparation, work-admission and database
status regressions also passed on development candidate
`8ab3b25bfe3fc4bcd85b676db2137506dfb2ff4d` under Looper evidence
`TX3WMIKAGXJZLJOVJS4EN6ZSUI`. That earlier evidence is separate from the exact
merged-head sequence above. Local HTTP/SQLite fixtures do not establish real
Telegram client acceptance or scheduled exhaustive validation.

## Risks and Follow-ups

Selections whose complete source/contract view exceeds the retained bounds
require a smaller selection or local inspection before a new confirmation.
Progress/notifications remain Slice 4 work; real mobile/desktop acceptance
remains Slice 5 work under its existing activation authority.
