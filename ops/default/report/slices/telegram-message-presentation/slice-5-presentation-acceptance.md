# Slice 5: Complete Coverage and Real Client Acceptance

Status: drafting
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `test/slice-5-telegram-presentation-acceptance`
PR: pending

## Purpose

Reconcile the complete producer inventory, prove the composed presentation
journey and record actual client acceptance alongside updated product guidance.

## Delivered Behavior

Slice 4 integration is verified and its report/tracker entries are closed.
T5.1 is implemented:

- All 48 initial families now record actual source owners, rare/dynamic branches
  and 90 exact regression roots. Producer and sender inspection found no
  additional semantic family. Two imprecise root references were corrected.
- Real SQLite capture tests cover near-limit accounting, exact accepted bytes,
  malformed UTF-8, oversized details and unchanged finished capture. HTTP tests
  cover rare help, syntax, empty-record and unavailable-owner outcomes without
  allocating work or altering a conversation.
- The composed local journey opens complete canonical artifact uploads, resolves
  a natural-language work preview through the configured HTTP model boundary,
  checks the complete confirmation and original keyboard, admits by exact reply,
  reopens SQLite and reaches a canonical native receipt on a real temporary Git
  target. Replaying confirmation creates no additional publication.

T5.2 real mobile/desktop inspection and user-operated acceptance remain pending.
T5.3 product guidance is prepared: the existing User Guide, Telegram transport
and notification references describe the actual help groups, attachment bounds,
complete confirmation scope, retry behavior and native evidence. Unreleased
notes describe one user outcome. The cycle report keeps client acceptance and
exact-candidate scheduled certification open.

## Implementation Notes

T5.1 is `ec0ea0f4c959188cee4477bc45888406f739a3a1`.

Activated from canonical `dev` at
`ae8d8d1b79ba6efddd75fdd287f235b6556babab` after PR #428 integration.
The retained live QA receiver and its native invocations remain under their
existing owner. This slice does not restart that receiver or fabricate input.

## Contracts Added or Changed

The accepted presentation bounds and operation guards remain unchanged. Fixture
certification remains separate from actual mobile/desktop observations and a
user-operated journey. The retained receiver is still on its earlier fixed
source `5bde31118db47c4fe52bff416555049a66d22911`; its activation requires a
separate decision under the approved plan.

## Files of Interest

- The approved plan, tracker and 48-family source review
- `internal/app/telegram*_test.go` and `internal/core/telegram*_test.go`
- Existing `docs/how-to/user-guide/index.md`
- `docs/reference/telegram-receiver.md` and `notifications.md`
- `CHANGELOG.md` and the delivery cycle report

## Validation

Successful T5.1 development checks:

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core ./internal/app` — zero issues.
- `go test ./internal/core ./internal/app -run '^TestTelegram(Conversation|Help|Authoring(Bridge|Intent|Prepared|Work|Explain|InstructionReference)|Presentation(Coverage|Document|Conversation|Journey))' -count=1 -timeout=3m` — core/app passed in 6.391/16.415 seconds.
- `go test ./internal/core ./internal/app -run '^Test(Telegram(Operation(LiveRecords|Record|Action|UnicodeParts|PartOrder)|Work(Extents|Capacity|OtherOwner|StalePreview|Projection|DeliveryFailure|HumanResult)|Control(Lifecycle|Decision|StaleReceipt|Reserve|RequestReserve)|HarnessSelection(Visibility|Unready|Confirm|ModelAuthority)|Presentation(Action|Control|Confirmation|Text|Progress|Notification|Failure|Format|Delivery)|ReceiverCallback|Recovery(Failures|PollEpoch))|TicketLoop(RecordsBlockedNotifications|DispatchesConfiguredFinishedNotification)|TicketApplyResultRecordsBlockedNotification|NotifyTelegramTest)' -count=1 -timeout=3m` — core/app passed in 30.516/30.256 seconds.
- `go test -race ./internal/core ./internal/app -run '^TestTelegramPresentation(Coverage|Journey)' -count=1 -timeout=3m` — core/app passed in 12.677/23.321 seconds.
- `git diff --check`

Both bounded normal selections include all 90 exact matrix roots. Source-owner
paths and local documentation links resolve. Initial fixture expectations were
corrected to use the existing syntax/empty-owner messages, the unique prepared
conversation and fresh status after queued progress. Native receipt and work
count assertions remain intact. All passes retain the three-minute package
budget. Real client observations and final PR-reference clean-HEAD validation
remain pending.

## Risks and Follow-ups

Real mobile/desktop client observations and the user-operated journey require
participation and a fixed authorized candidate receiver. The existing scheduled
nightly must certify the exact integrated final candidate; keep
`awaiting_nightly` until that evidence exists.
