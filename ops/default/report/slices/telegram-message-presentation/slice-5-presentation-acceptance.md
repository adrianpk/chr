# Slice 5: Complete Coverage and Real Client Acceptance

Status: delivered
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `test/slice-5-telegram-presentation-acceptance`
PR: [#429](https://forge.adrianpk.com/adrianpk/chr/pulls/429)

## Purpose

Reconcile the complete producer inventory, prove the composed presentation
journey and record actual client acceptance alongside updated product guidance.

## Delivered Behavior

Slice 4 integration is verified and its report/tracker entries are closed.
T5.1 is delivered:

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

T5.2 is delivered with the client coverage limits recorded below. The first
real desktop request to show a specification
delivered an explanation export rather than the original file. The authorized
repair distinguishes document-read intent from explanation, reads the selected
canonical dump/spec/plan without generation, and freezes its full original
bytes for delivery. Missing documents and changed sources remain failures.
The desktop retry subsequently delivered all 937 bytes of the original
specification. The real user-started native journey subsequently completed with
a verified delivery receipt and the ticket marked done. Its operator-assisted
recovery, mobile document proof and remaining client coverage limits are
recorded below.
T5.3 product guidance is delivered: the existing User Guide, Telegram transport
and notification references describe the actual help groups, attachment bounds,
complete confirmation scope, retry behavior and native evidence. Unreleased
notes describe one user outcome. The cycle report records the accepted client
scope and its limits, with exact-candidate scheduled certification still open.

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
budget. Later real client observations are recorded below. The controller
records final PR-reference validation against its exact clean HEAD.

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
The canonical PR uses report-introduction commit
`8260c0d484521883db7c01f38e44991118e963e1`.
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

The clean native-provider candidate
`0a9f83d954454dbb28276c67d7ea5850b99a239e` passed `make verify`, the bounded
normal selection and focused race checks. Its activation preserved the original
worker proof and prior stopped invocations. The real host then admitted native
integration and exposed an authentication-context defect: its disposable bare
repository did not inherit the retained project's credential helper.

A real authenticated HTTP Git fixture first reproduced that preparation failure.
The correction carries only local credential settings into child environments
for both preparation and verification. A separate control regression exposed
the raw adapter in request-bound continuation; lifecycle and decision controls
now use the same native provider. Explicit revision-bound continuation can
recover an interrupted preparation before its validation input exists, retaining
the failure and all counters. Unbound calls, stopped runs, accepted validation
inputs and receipt replay cannot retry that work. The real-provider fixture
uses a standalone ticket with one worker and one submission allowed.

The bounded credentials/recovery selection passed in core/app in 9.672/5.167
seconds. The private read-only Git preflight also passed using the receiver's
own environment and the retained project's helper context, with no SSH agent.
The existing positive HTTPS rebase certification is selected for that isolated
host target. Candidate activation and confirmed integration remain pending at
this checkpoint.

Candidate `a577d03490fe57262657e72954240cb0c156a610` passed clean development,
focused normal and race checks and was activated with the existing certified
HTTPS publisher profile. A real CLI continuation then exposed a silent no-op:
the CLI omitted its observed revision, so the deliberately unbound call could
not authorize preparation retry. The CLI now reads that revision and forwards
it to the same guarded operation. Its regression reproduces the no-op before
the correction and requires completed preparation afterward.

The guarded real CLI recovery completed preparation without repeating its
original worker. The native host then reviewed and atomically published the
candidate; canonical Git and the original PR both reported the prepared result.
Receipt progression exposed one more missing adapter capability: the wrapper
did not forward the publisher's read-only inspection interface. It now exposes
that interface through the configured target transport. The real-provider test
covers both automatic observation and explicit native reconciliation through
canonical receipt finalization. Explicit submission reconciliation also uses
the native owner, allowing observation after the retained polling deadline
without another publication or a deadline/budget reset.

Clean candidate `8d5d13a0215dbbd7d5ecd9a6d138e9f8ff412c67` passed:

- `make verify` — zero lint issues; short Go tests, eight CI helper tests and build passed.
- `go test ./internal/core ./internal/app -run '^Test(NativeHostPublication|ServiceNativeProvider)$' -count=1 -timeout=3m` — core/app passed in 26.815/18.944 seconds.
- `go test -race ./internal/core ./internal/app -run '^Test(NativeHostPublication|ServiceNativeProvider)$' -count=1 -timeout=3m` — core/app passed in 30.673/24.211 seconds.
- `git diff --check` — passed.

That exact clean executable is active in the retained QA receiver. The genuine
user-started request reached `completed` / `ticket_integrated`: one validated
task, one confirmed ticket, zero remaining tickets and a hash-verified native
delivery receipt. The canonical PR and both remote refs identify the prepared
result. Its README preserves the entire original content and appends exactly the
requested paragraph; no other file changed. The terminal notification was
delivered through the existing receiver.

This was an operator-assisted recovery, not a fresh unattended acceptance run.
It reused the original worker, retained the failed preparation, retried only
preparation under the observed revision and reconciled the already-published
result after the observation deadline. There was one Git publication; original
limits and the deadline were unchanged. Prior paused/cancelled invocations were
verified byte-identical. Private client evidence retains the exact run, receipt,
remote observations, content hashes and activation records.

The next client repair removes nested ticket/source metadata from the short
caption by reading the original description from the immutable execution
contract, preserving the complete proposed ticket and operation in the attachment.
Preparation uses the title in the retained execution contract when no metadata
title exists, omits the internal project ID
from its heading and removes the storage path from the prepared-ticket result.
Progress confirms the retained receipt without announcing zero integrated units
for standalone tickets; validated and confirmed ticket counts remain distinct.
No prepared input, choice key, receipt or execution limit changes.

The existing progress-state and actual HTTP/SQLite preparation tests first
reproduced these defects. After correction,
`go test ./internal/core ./internal/app -run '^Test(TelegramPresentation(DocumentExcerpt|ProgressStates)|TelegramAuthoringPreparedDecision)$' -count=1 -timeout=3m`
passed in core/app in 0.005/0.767 seconds. Clean-candidate development and broader
affected regressions were then run against the final repair candidate.

The retained QA receiver also exhausted its polling transport retries after a
network failure. A fresh connection doctor passed. One supported local polling
retry cycle and explicit service activation restored ready health with no queued
updates or deliveries. Prior stopped invocations and the completed real journey
were verified byte-identical; no work or model allowance was reset.

Clean candidate `ce16bea12d4acf0d129d83de643d17333607a957` passed:

- `make verify` — zero lint issues; short Go tests, eight CI helper tests and build passed.
- `go test ./internal/core ./internal/app -run '^Test(TelegramPresentation(Document|Progress|Journey|ActionWork|Confirmation)|TelegramAuthoringPrepared|TelegramWork(Automatic|Evidence|HumanResult|DeliveryFailure))' -count=1 -timeout=3m` — core/app passed in 2.262/28.040 seconds.
- `git diff --check` — passed.

The unchanged prepared inputs, complete attachments, action replay, native
receipt and single-publication checks remain enforced. Earlier strict-lint and
old-wording assertion failures are retained with their actual candidates rather
than reused as passing evidence. No storage or concurrency owner changed.

The exact clean executable is pushed and active in isolated QA. Fresh polling,
ready health and empty queues were verified after activation. Updating the
executable required recreating the transient unit after stopping it unloaded
its definition. The polling retry cycle, publisher profile, prior invocation
hashes, completed journey and one-publication count remain unchanged. The
delivered-notice count also remained unchanged; activation generated no new
input or duplicate result. Private evidence retains validation, backups and
activation identity. Actual review of the repaired messages remains pending.

The next real status question exposed missing execution references in intent
context: preparation focus was selected despite having no native binding.
The bounded context now identifies observed invocation bindings and the most
recent acknowledged work message. An explicit replied work message takes
precedence. Both references are peer-bound and filtered through currently
allowed projects; they grant no control authority.

Read-only status resolution accepts acknowledged progress, terminal and bound
status messages after capture closes. Status responses retain that reference
atomically with source acknowledgement. Native work replies bypass capture
dispatch before ordinary interpretation. Capture and mutating-control reply
guards remain unchanged. An unbound status proposal gets an inspection question
rather than a control diagnostic.

The composed HTTP/model/SQLite/Git journey first reproduced the missing context,
then the finished-reply capture failure. It now covers a plain status question,
a terminal reply and a reply to a status response without changing canonical
work or repeating publication. SQLite tests also reject foreign peers,
unacknowledged messages, scope conflicts and changed roots, and preserve an
explicit reply when a newer work message exists. The focused development result
is core/app 0.226/13.790 seconds. Clean-candidate checks and real-client retry
remain separate acceptance obligations.

Clean candidate `29fc714b0687e149397ec4f6dea69fb638e5fcce` passed:

- `make verify` — zero lint issues; short Go tests, eight CI helper tests and build passed.
- `go test ./internal/core ./internal/app -run '^Test(TelegramWorkStatus(Reply|Guards)|TelegramPresentationJourney|TelegramControl(Conversation|Natural|Lifecycle|ReplyArrival|AuthoringLane|Authority|Stale|LaneFence|Reference|ContextGuard)|TelegramAuthoringIntent(RootChange|Invalid))$' -count=1 -timeout=3m` — core/app passed in 0.385/15.223 seconds.
- `go test -race ./internal/core ./internal/app -run '^TestTelegram(WorkStatus(Reply|Guards)|ControlContextGuard|PresentationJourney)$' -count=1 -timeout=3m` — core/app passed in 8.114/21.787 seconds.
- `git diff --check` — passed.

Earlier control checks still expected the pre-presentation paused/cancelled text;
their wording was corrected while retaining native revision, stop-intent and
stale-action assertions. Failed candidate evidence remains separate. The clean
executable is pushed and active in QA with fresh polling and empty queues.
An atomic executable replacement and restart preserved the transient unit's
environment, transport cycle, publisher profile and prior invocation hashes.
The failed real client interpretation remains unchanged; it was not replayed
or replaced. A read of retained state confirmed that the latest acknowledged
work reference identifies the completed journey.

The subsequent real Telegram retry passed: a new natural-language status query
identified invocation `00005-workflow-invocation` and returned its finished
result, one validated task, one confirmed ticket, zero remaining tickets and
the integrated pull request. The user supplied the actual client response. It
confirmed integration from retained native receipts and requested no further
action. This verifies plain status targeting and the terminal ticket counters
on the active QA candidate; it does not close the remaining client matrix or
establish a fresh unattended work execution.

## Client Acceptance Refinement

On 2026-10-22 the real natural-language request first arrived without a reply
binding and correctly asked which conversation owned the specification. A
genuine follow-up identified the README change and delivered the original
spec.md with a bounded direct excerpt. The maintainer supplied the actual
response and identified mobile Telegram as its source.

The retained delivered payload and the maintainer's downloaded copy both contain
all 937 original bytes and SHA-256
`4d9a7a241c23dd7fe346656652c9aee3e80c0004cbae338143c9618a892d6447`.
Independent file inspection confirmed exact source equality, including front
matter and provenance. The initial classification as a desktop download was
incorrect: where the file was stored did not identify the Telegram client.
Acta correction `01M5J7YD51QVDHVTCFWPM6ZVCT` records the mobile origin.

The maintainer rejected duplicating that same account's document check on
desktop. The specification, plan and cycle matrix now record the refined
acceptance scope. Original-document delivery is accepted; client/build versions,
separate desktop rendering and mobile reader layout remain unobserved limits.
Neither the mobile result nor fixture passes certify those unobserved layouts.
The prior native journey remains operator-assisted, with its original worker,
receipt, single publication and execution limits unchanged.

Looper receipt `UR5P2V7ZBTEZUVWSD6TVHNJ6NY` passed all five frozen commands on
clean pushed candidate `0ea4463a02c0cfa144bee0d26ecc5e5106713538`. This acceptance
record changes documentation only. The PR-reference commit requires its own
exact-HEAD validation before the existing controller integrates PR #429.

Final local validation passed the frozen four-command prefix above, then these
bounded acceptance-repair regressions and `git diff --check` last:

- `go test ./internal/core ./internal/app -run '^Test(NativeHost(Publication|Denials|PreparationRetry)|NativePublisherProfiles|Workflow(IntegrationCredentials|PreparationRetryAuthority)|ServiceNativeProvider|WorkContinuePreparation|TelegramWorkStatus(Reply|Guards))$' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(NativeHost(Publication|Denials|PreparationRetry)|NativePublisherProfiles|Workflow(IntegrationCredentials|PreparationRetryAuthority)|ServiceNativeProvider|WorkContinuePreparation|TelegramWorkStatus(Reply|Guards))$' -count=1 -timeout=3m`

These roots cover configured provider wiring, native publication and receipt
observation, profile denials, credential propagation, revision-bound recovery
and acknowledged status context. The frozen presentation selections retain
document-read and composed-journey coverage. No exhaustive gate is added.

## Integration

Canonical PR #429 merged into `dev` at
`af999d9c73e4ee9d302c9296c89e29bcb1e1d10a`. Looper receipt
`4JWXLVN42SZHAF2L3LILCY4UMJ` passed all seven ordered commands on that exact
candidate. Acta merge event `01M5J9BCR8BVWE0ZHVDH09V329` confirms the
controller-owned Rebase + Fast-forward integration. Slice 5 of 5 is delivered.

The subsequent gate-preparation changes are documentation only. The controller
owns the full `make verify` run on their clean pushed commit; the merged slice's
receipt is not reassigned to that later candidate. Exact scheduled exhaustive
certification remains `awaiting_nightly`.

## Risks and Follow-ups

Readiness was rechecked on 2026-10-22 without changing the receiver. The running
executable reports clean source `29fc714b0687e149397ec4f6dea69fb638e5fcce` and
SHA-256 `3cc62f6f09f29f0eee70bb59f368251a668081b456cd7e4af3eef99fbfdeb7fe`.
Status reports ready health, fresh polling, zero pending updates and zero pending
deliveries. Later commits through `13a908914ed69f82e9ea29d6ad79373af061f988`
change only acceptance records. Canonical PR #429 is integrated into `dev` at
the exact candidate recorded above.

The [cycle matrix](../../telegram-message-presentation-cycle.md) distinguishes
the observed mobile document, operator-assisted native journey and plain status
retry from unobserved client layouts and versions. The standalone-ticket
caption/result repairs remain active in QA and the real status retry verified
terminal ticket counters. These observations establish the recorded acceptance
scope, not a fresh unattended run or complete cross-client UI certification.
The existing scheduled nightly must still certify the exact integrated final
candidate; keep `awaiting_nightly` until that evidence exists.
