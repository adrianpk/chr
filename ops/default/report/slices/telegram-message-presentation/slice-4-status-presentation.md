# Slice 4: Progress, Notifications and Failure Consistency

Status: reviewing
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `feat/slice-4-telegram-status-presentation`
PR: [#428](https://forge.adrianpk.com/adrianpk/chr/pulls/428)

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

T4.2 is implemented:

- Workflow notifications retain a concise human/project summary and one complete
  JSON event export. Manual tests identify their test purpose and project without
  displaying the workspace directory.
- The notification event and its view/export are committed atomically under the
  existing event repository. Both snapshot tables share the 64 MiB core database
  quota. Current sender, destination, namespace and root bindings are rechecked
  before each transmission.
- Real positive message/file acknowledgements establish delivery. One definite
  format rejection can select plain formatting for the same bytes; uncertain
  responses retain their event, snapshot and consumed allowance.

T4.3 adds fault evidence for both owners:

- Lost, missing and server-failed notification acknowledgements retain the
  original bytes and consumed attempt after reopening; they never claim delivery
  or silently dispatch another event. Access rejection is a definite failure.
- Future/corrupt manifests, missing images and changed destinations, token,
  namespace or canonical root fail before transmission. A changed sender between
  format rejection and fallback prevents the second upload.
- Receiver and notification admissions race for the last 2 MiB of the shared
  database quota. Only one commits; the other leaves no partial owner/image.
  Cleanup removes at most twenty aged terminal images and retains unresolved,
  uncertain and recent images plus all original hashes/transport receipts.
- A physical pre-28 notification database retains original events and NULL or
  version-zero payload meaning. Legacy dispatch reads persisted payload and
  refuses a stale caller's attempt to resend an already delivered event. Partial
  new mappings cannot become legacy: missing manifests or version-zero manifests
  with retained new binding metadata are refused before IO.
- An unavailable receiver read identifies unchanged state and local diagnostics.
  Notification rejection leaves a successful native ticket-loop result and its
  sole execution intact. Oversized export preparation records a failed audit
  without sending a shortened substitute.

## Implementation Notes

Activated from canonical `dev` at
`a4d0a865e8bb32a3caac65395a28839b573ed369` after verified PR #427 integration.
T4.1 is `708afec21584e546624848391a287a59fb5da5d4`.
T4.2 is `04fddeb15cf6184f4c263a55bdb222fcd74a6b1e`.
T4.3 is `01d9be1341e5c207c1748a22ced4a5aac8eafe1d`.

Migration 28 adds nullable notification presentation and document mappings,
transport evidence and bounded safe retention. Historical NULL/version-zero
records retain the old notification sender and payload. The shared uploader
supports the existing string destination without relaxing receiver token checks.

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
- `internal/core/notification_telegram_*.go` and migration 28
- `internal/core/telegram_notification_delivery.go` and `internal/app/telegram_notification_*.go`
- Notification HTTP, migration, quota, retention and failure fixtures
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

Successful T4.2 development checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramPresentation(Notification|Document|Migration|Confirmation)|NotificationEventRepository|NotifyTelegramTest|TicketLoop(DispatchesConfiguredFinishedNotification|RecordsFailedNotification|RecordsBlockedNotifications)|TicketApplyResultRecordsBlockedNotification|MaintenanceDBStatusReports)' -count=1 -timeout=3m` — core/app passed in 2.746/1.480 seconds.
- `go test -race ./internal/core ./internal/app -run '^TestTelegramPresentationNotification' -count=1 -timeout=3m` — core/app passed in 4.984/3.083 seconds.

Successful T4.3 development checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramPresentation(Progress|Notification|Failure)|TelegramWorkDeliveryFailure|TelegramHelp(HTTP|Unauthorized)|TelegramRecovery(Failures|Doctor|PollEpoch)|TelegramPresentationDelivery(HTTP|Revocation)|TelegramOperation(UnicodeParts|PartOrder))' -count=1 -timeout=3m` — core/app passed in 1.586/2.909 seconds.
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramPresentation(Progress|Notification|Failure)|TelegramWorkDeliveryFailure|TelegramHelp(HTTP|Unauthorized)|TelegramRecovery(Failures|Doctor|PollEpoch)|TelegramPresentationDelivery(HTTP|Revocation)|TelegramOperation(UnicodeParts|PartOrder))' -count=1 -timeout=3m` — core/app passed in 40.639/49.615 seconds.

The retained-row consistency regressions reproduced a legacy fall-through before
correction: missing/partial manifests reached the legacy sender. The reader now
requires all new mapping metadata to be NULL for a genuine legacy row. The
focused post-correction command passed (core/app: 0.581/0.461 seconds):

- `go test ./internal/core ./internal/app -run '^TestTelegramPresentation(NotificationLegacy|NotificationMigration|FailureNotification(Mapping|BeforeIO))' -count=1 -timeout=3m`

These local HTTP/SQLite/native fixtures exercise both delivery owners, including
real lost acknowledgements, revoked input/destinations and local-only outage
classification. They retain the three-minute package budget.

Final clean-HEAD evidence is recorded through Looper after the PR-reference
follow-up commit. The canonical PR lists successful commands from that exact
candidate in this order:

- `make verify`
- `go test ./internal/core ./internal/app -run '^TestTelegramPresentation(Progress|Notification|Failure)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramPresentation(Progress|Notification|Failure)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(TelegramWorkDeliveryFailure|TelegramHelp(HTTP|Unauthorized)|TelegramRecovery(Failures|Doctor|PollEpoch)|TelegramPresentationDelivery(HTTP|Revocation)|TelegramOperation(UnicodeParts|PartOrder))' -count=1 -timeout=3m`
- `git diff --check`

Controller evidence is tied to the clean PR head and this ordered command list;
the development results above do not replace it. Real client acceptance is
recorded separately in Slice 5.

## Risks and Follow-ups

Real mobile/desktop and user-operated journey acceptance remains Slice 5 work.
