# Slice 3: Clear Guarded Confirmations and Controls

Status: reviewing
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `feat/slice-3-telegram-action-presentation`
PR: pending

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

Successful development checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramPresentation(Action|Control|Confirmation)|TelegramAuthoringPreparedIntent|TelegramHarnessSelection(Confirm|SendRetry))' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core ./internal/app` — zero issues.
- `go test ./internal/core ./internal/app -run '^Test(TelegramOperationAction|TelegramControl|TelegramHarness|TelegramAuthoringPrepared|TelegramWork(Extents|LostAck|StalePreview|Admission|Namespace|CommandBounds)|MaintenanceDBStatusReports)' -count=1 -timeout=3m`
- `git diff --check`

The normal selection took 1.226 seconds for core and 2.033 seconds for app,
excluding compilation. It includes the actual 64 MiB quota fixture, independent
SQLite connections, old physical-schema upgrade, complete scope upload and
original-result replay. The three-minute package budget is unchanged.

The broader development legacy selection first exposed obsolete chat-audit
assertions and a lost-ack injector limited to sendMessage. Those fixtures now
check retained owner identity and document-upload acknowledgement loss. The
complete affected legacy selection then passed in 27.933 seconds for core and
43.048 seconds for app. Looper evidence `TX3WMIKAGXJZLJOVJS4EN6ZSUI`
records that command followed by `git diff --check` on the exact clean candidate
`8ab3b25bfe3fc4bcd85b676db2137506dfb2ff4d`. This is focused development
evidence, separate from the required final command sequence.

Final exact clean-HEAD evidence must be recorded through `looper validate`
after canonical publication and the PR-reference commit: `make verify`, the
configured normal selection, its race selection, then `git diff --check`.
No final gate or real Telegram client acceptance is
claimed here.

## Risks and Follow-ups

History reconciliation retains the earlier published closure/blocker commit
as an ancestor of the current candidate. Conflict resolution keeps the current
Slice 2 delivery evidence, active Slice 3 implementation and resolved admission
prerequisite. Runtime files are unchanged. Publication can use an ordinary
fast-forward push; final PR-head validation remains required.

Selections whose complete source/contract view exceeds the retained bounds
require a smaller selection or local inspection before a new confirmation.
Progress/notifications remain Slice 4 work; real mobile/desktop acceptance
remains Slice 5 work under its existing activation authority.
