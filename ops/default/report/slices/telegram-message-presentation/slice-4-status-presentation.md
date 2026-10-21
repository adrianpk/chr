# Slice 4: Progress, Notifications and Failure Consistency

Status: drafting
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `feat/slice-4-telegram-status-presentation`
PR: pending

## Purpose

Present truthful progress, waits and failures through the receiver and existing
notification sender with shared bounded views and frozen details.

## Delivered Behavior

T4.1 is implemented:

- Progress leads with work/project context, actual counts and the current wait,
  stop, blocker or next responsible party. Integration requires the retained
  native receipt; a successful validation or open PR does not stand in for it.
- New progress notices freeze one bounded presentation under the original native
  revision. Recovery keeps the same work, notice identity and remaining attempts.
- Full queues explicitly reject the new request while retaining existing work
  and reserved control capacity. Lost delivery acknowledgement warns of a repeated
  notice without allocating a domain operation.
- Newly composed plain responses use one view within 1,800 UTF-16 units and a
  supported narrowing/local inspection route. Stored legacy multipart rows retain
  predecessor acknowledgement and blocking; formatted help keeps its existing
  bounded command navigation.

T4.2 notification snapshots/sending and T4.3 both-path fault coverage remain pending.

## Implementation Notes

Activated from canonical `dev` at
`a4d0a865e8bb32a3caac65395a28839b573ed369` after verified PR #427 integration.
T4.1 is `708afec21584e546624848391a287a59fb5da5d4`.

The existing notification event repository owns its dispatch and audit records;
no receiver, scheduler, trigger or workflow authority is added.

## Contracts Added or Changed

The adapter-neutral work status exposes an optional source-owned work title.
The existing progress projection can carry the shared message blocks; the outbox
retains their manifest/hash without changing native authority, coalescing or
send reservation ownership. Existing stored rows are not rerendered.

## Files of Interest

- `internal/app/telegram_work_progress.go`, `telegram_work_activation.go` and `telegram_control_dispatch.go`
- `internal/core/telegram_work_projection.go` and `telegram_response_chunks.go`
- `internal/app/telegram_progress_presentation_test.go` and `internal/core/telegram_progress_presentation_test.go`
- Existing native-work, control, harness, help, capacity and legacy multipart fixtures

## Validation

Successful T4.1 development checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramPresentationProgress|TelegramOperation(UnicodeParts|PartOrder|ActionLostAcknowledgement)|TelegramWork(Projection|Evidence|HumanResult|DeliveryFailure|Automatic)|TelegramControl(Reserve|StaleReceipt)|TelegramHarnessSelection(LostAck|SendRetry|Redelivery)|TelegramHelp|TelegramPresentationConversationHelp)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core ./internal/app` — zero issues.
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramPresentationProgress|TelegramWorkDeliveryFailure)' -count=1 -timeout=3m`
- `git diff --check`

Normal core/app checks took 0.833/17.477 seconds; the focused race selection took
3.085/6.423 seconds within the unchanged three-minute package budget. Real local
HTTP/SQLite/native-work fixtures cover original-result and acknowledgement
recovery, legacy message ordering and unchanged CLI/TUI/Web evidence access.
Long-title checks reserve the state and inspection route before truncation.

Required final clean-HEAD checks remain `make verify`, the configured normal
progress/notification/failure selection, its focused race selection and
`git diff --check`, recorded through the controller after PR references.
No final gate or actual Telegram client acceptance is claimed here.

## Risks and Follow-ups

Real mobile/desktop and user-operated journey acceptance remains Slice 5 work.
