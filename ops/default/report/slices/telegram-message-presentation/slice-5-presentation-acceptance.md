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

T5.2 is in progress. The first real desktop request to show a specification
delivered an explanation export rather than the original file. The authorized
repair distinguishes document-read intent from explanation, reads the selected
canonical dump/spec/plan without generation, and freezes its full original
bytes for delivery. Missing documents and changed sources remain failures.
The desktop retry subsequently delivered all 937 bytes of the original
specification. Mobile/desktop coverage and user-operated native completion
remain pending.
T5.3 product guidance is prepared: the existing User Guide, Telegram transport
and notification references describe the actual help groups, attachment bounds,
complete confirmation scope, retry behavior and native evidence. Unreleased
notes describe one user outcome. The cycle report keeps client acceptance and
exact-candidate scheduled certification open.

## Implementation Notes

T5.1 is `ec0ea0f4c959188cee4477bc45888406f739a3a1`.

Activated from canonical `dev` at
`ae8d8d1b79ba6efddd75fdd287f235b6556babab` after PR #428 integration.
The validated presentation candidate was activated in the retained isolated QA
receiver for user-operated acceptance. Its existing data, unconsumed confirmation
and native invocation ownership were preserved. No inbound input was fabricated.

## Contracts Added or Changed

The accepted presentation bounds and operation guards remain unchanged. Fixture
certification remains separate from actual mobile/desktop observations and a
user-operated journey. Natural-language document reads use the existing
explanation operation with a validated dump/spec/plan selector and an exact
observed conversation. The shared service binds the original path and hash;
the read cannot substitute model context excerpts or generated prose. The
existing explanation route and retained requests remain compatible.

Real work admission exposed another acceptance defect: the application supplied
the ordinary remote adapter without the prepared-native capability used by the
test providers. The validated task and canonical PR remained waiting while
bookkeeping revisions repeatedly generated identical progress notices. Commit
`9c5cea5de904d0cd8f84d847e746765912a70737` suppresses unchanged notices across
receiver restart while preserving the observed native revision and real changes.

The native integration repair binds the default service's preparation, review,
controls and continuation to the exact retained invocation. A private host
publisher profile supplies the certified repository/target, transport and
credential binding. Current source ownership, API identity, repository policy
and actual receive-pack capability are checked before publication. Host profile
changes invalidate the existing grant. No branch prefix, source-only merge API
or worker-reported validation can authorize publication. The same configured
publisher performs the existing atomic exchange and retains its original PR.
The setup and native-integration reference document the host profile format.

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

Looper receipt `5I4WQ32LVGGE7LXENUNTHC3G7T` passed the complete configured
ordered command list on clean, pushed candidate
`658f39ae34747bd85c6ff2b8b777306da87e0f73`:

- `make verify`
- `go test ./internal/core ./internal/app -run '^TestTelegramPresentation(Coverage|Journey)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^TestTelegramPresentation' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramPresentation(Document|Delivery|Migration|Action|Control|Confirmation|Progress|Notification|Failure)' -count=1 -timeout=3m`
- `git diff --check`

The staged executable records that exact clean source and SHA-256
`b5c32a29a11593f3e2f1d088c4a89beb39707ce501b6153b592706d6c2466c04`.
The canonical PR payload is prepared with report-introduction commit
`8260c0d484521883db7c01f38e44991118e963e1`; it has not been submitted.
This receipt belongs to the named candidate and cannot certify later report,
acceptance or PR-reference commits.

The desktop capture and durable delivery snapshot confirmed a 546-byte generated
explanation where the user requested a 937-byte canonical specification. That
is a failed document-selection observation, not successful acceptance. The
repair regression first failed at the document-read boundary. Focused checks
cover complete original dump/spec/plan bytes beyond the 8 KiB model excerpt,
unchanged authoring state and generation count, missing/changed source and root
guards, and original bytes after SQLite reopen and an independent file edit.
Final repair-candidate checks and the actual client retry are recorded separately
from the earlier candidate's receipt.

The native provider regression first failed with
`production provider cannot prepare native publication`. After the repair:

- `go test ./internal/core ./internal/app -run '^Test(NativeHost(Publication|Denials)|NativePublisherProfiles|NativeCapability|NativePreparation(Methods|Admission)|ServiceNative(Provider|Review))$' -count=1 -timeout=3m` passed, core/app in 40.549/13.969 seconds.
- `go test -race ./internal/core -run '^TestNativeHost(Publication|Denials)$' -count=1 -timeout=3m` passed in 19.049 seconds.
- `go test -race ./internal/core ./internal/app -run '^Test(NativeHost|NativePublisherProfiles|ServiceNativeProvider)$' -count=1 -timeout=3m` passed the profile and service-wiring roots; the host-publication roots are covered by the separate command above.
- A private read-only preflight passed against the retained real invocation and
  canonical PR, using existing QA credentials and repository/method certification.

An earlier broad prefix selection exceeded the three-minute package budget
while enumerating unrelated native review-capability cases. That incomplete
run is retained as failed evidence; the bounded exact-root selection above is
the focused regression proof. Clean-candidate development checks and activation
of this provider repair remain separate from these local results.

## Risks and Follow-ups

Real mobile/desktop client observations and the user-operated journey require
participation and the fixed validated repair candidate. The original pending
confirmation and native invocation ownership remain preserved. T5.2 cannot close
from local fixture evidence. The existing scheduled
nightly must certify the exact integrated final candidate; keep
`awaiting_nightly` until that evidence exists.
