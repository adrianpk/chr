# Slice 6: Guarded Controls and Capacity

Status: reviewing

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-6-telegram-controls`

PR: `#395`

## Purpose

Expose exact-target native controls, bound decisions and shared capacity through
responsive Telegram conversations and the common workflow owner.

## Delivered Behavior

T6.1 (`a3e5e95be3b45aaa66cc0c1bc0744b73adac6770`) adds status and guarded pause, cancel, resume, continue and reconcile for
exact native invocations or bound conversations. Mutations display the mapped
operator and native revision, then use the shared `ControlWorkExtent` service
after explicit confirmation. Natural-language controls use that same typed path.

A stale confirmation displays current native state. New operator cancellation
survives an older pause prompt, and both stop intents survive asynchronous real
worker termination and retained checkpoints. Receiver status and draft
cancellation retain their existing commands.

T6.2 (`76b55b28ed770b13fc40da53ff39e42dbe822ae9`) adds explicit and natural-language approval, bounded grants and retained
question answers. Every preview freezes the exact invocation, review/candidate,
integration entry, budget and question revisions required by its native owner.
The displayed option number resolves to a retained option before confirmation;
the configured actor and native question authority are never model arguments.

Approval, grant and answer receipts share the existing native file journal with
their effect. A retry returns the original immutable request receipt before
checking newer mutable revisions and observes current status separately. This
recovers interrupted native transactions and SQLite acknowledgement loss without
repeating a grant, approving a successor or clearing a later operator pause.
Control conversations retain native identity and preserve original work-admission
authority; a control-only conversation grants no automatic execution.

T6.3 progress (`aaca31360292b67e72509557fa4b650021593aaf`) separates durable polling, application calls and outbound delivery.
One receiver shares 16 application slots between its work host and inbox calls:
12 ordinary slots, two decision slots and two status/stop/reconciliation slots.
Confirmation callbacks, explicit answers and replies derive priority from the
same retained peer-owned action. Continuations and authoring confirmations keep
ordinary slots because they can execute a worker or call a model.

The 1,024 pending-input ceiling includes 32 slots reserved for controls. Ordinary
overflow commits explicit rejected metadata and a capacity diagnostic with the
complete polling batch and cursor. A failed diagnostic write rolls back the
whole batch. Restart restores classification for existing pending input.
Control responses precede older ordinary output while preserving chunk order.

Interpretation and generation share two OS-owned model slots across store
instances, with 60-second deadlines and bounded source-lock stripes. Waiting for
capacity spends no retained retry attempt. Ordinary messages preserve source
order within a peer; explicit controls can bypass a blocked model. Interpreted
target resolution checks the original peer-context digest in its SQLite
transaction, preventing a changed focus or session from silently retargeting
an older model response.

T6.3 lifecycle progress (`b2f03c1159ea56ab4e482a2519489b1eec42e736`) extends native request receipts to pause, cancel,
resume, continue and reconcile. The existing stop/resume, task-checkpoint and
closing owners journal each accepted effect with its immutable caller receipt.
Publication-pointer repair, stop intent and the receipt now share one transaction,
including an invocation whose submission projection was interrupted.

A continuation receipt confirms acceptance of one bounded call. Its acceptance
checkpoint and receipt commit before execution leaves native exclusion. Replaying
that request returns its original acceptance and current status without starting
another checkpoint, replacing a later pause or resetting attempts. A partial
journal cannot be reported as a confirmed control. An accepted receipt whose
acknowledgement meets an expected domain barrier stays pending for receipt replay;
it does not terminate receiver intake. Existing callers without a request binding retain their prior operation APIs.
Read-only authority and receiver-status connections now use the same bounded
SQLite busy timeout as writers, allowing a short inbox commit to finish before
a concurrent read fails.

T6.3 receipt recovery (`dcf386f2423a669707d1609a92eeee9a784453a5`) reads a completed native control before considering a
fresh effect. Expired or superseded prompts retain their original answer binding;
current mapped actor, exact peer, allowed registered root and namespace remain
mandatory. A missing receipt cannot reserve a binding or apply a control.

Explicit answers, callbacks and invalidated replies recover the same original
receipt. A later answer can deliver that receipt after an earlier diagnostic
acknowledged the first source, without allocating another effect identity.
Acknowledgement preserves replacement conversation revisions and newer native
observations; it does not overwrite their phase or admission reference.

T6.3 completion (`27165b596e763e5117177a83592044540855a93f`) adds late reply handling and capacity rejection. An unbound reply
waits while its exact peer's sender can still commit a definite acknowledgement;
unknown and exhausted sends produce diagnostics without reaching a model.
The delivery transaction refreshes pending reply classes. A subsequent activation
checks the immutable operation class again before entering a worker or model.
Known decisions precede unresolved reply waiters within the bounded scan.

All 12 ordinary adapter slots can remain occupied while exactly two urgent and
two decision calls process status and diagnostics. A real HTTP fixture holds a
control prompt's acknowledgement, receives its reply and confirms native
cancellation only after the exact binding commits. Separate fixtures prove that
neither continuation nor authoring enters a reserved activation after a late
binding. Stale prompts include current native state only after current mapped
actor, peer, root and namespace checks succeed.

Request reservation retains 64 metadata slots for the 32 control prompt/answer
pairs, within the existing 1,024 pending and 8,192 retained request ceilings.
Ordinary reservations stop before consuming those slots. Exact control answers
inherit the allowance from their retained peer-owned parent. Existing request
bindings still replay at either ceiling. Rejected reservations leave no accepted
request or effect and return a durable capacity diagnostic; status still works.

T6.4 (`8346ab32efee666563fddc19ee335b4d86915e6d`) certifies one native pool across two Telegram conversations,
daemon continuation and manual task execution. The default admits one source
worker. At the configured maximum, one Telegram worker, one daemon worker and
six manual workers overlap behind actual source-process barriers. A ninth
conversation, manual claim and daemon activation wait without spending a new
attempt or continuation; duplicate activation preserves the current owner.
Releasing one slot admits the waiting conversation, every accepted task validates
once, and all slots return to the pool. Linked-root/namespace subprocesses also
prove distinct reservations at capacities one, two and eight.

Existing native publication fixtures verify target exclusion before competing
provider IO, accepted and lost acknowledgments, unresolved intent retention and
fresh combined validation/review after canonical target movement.

## Implementation Notes

Native lifecycle entry points recheck volatile caller authority while holding
the existing workflow owner lock. Telegram source requests retain the exact
project root, namespace, conversation revision, invocation and mapped operator.
Explicit standalone controls allocate a source-stable confirmation conversation;
status only observes state. No transport execution scheduler is introduced.

Control request identities reserve their exact payload under native exclusion.
They use the existing bounded operation-binding and receipt directories, with a
separate control schema and checksum. Native review, budget, question, lifecycle and recovery owners include the
receipt in their existing commit transaction. Volatile caller
authority is checked again before that commit. Legacy callers retain their
existing API; a final decision authority callback requires a request binding.

Claimed-worker stop polls read their canonical owner's exact invocation under
the existing repository exclusion and journal/read guard, checking owner root,
project, namespace and the attempt credential. They no longer enumerate every
invocation in every source worktree every fifty milliseconds. Identity-only
closing and publication observers retain public scope resolution. Polling,
worker and host-validation deadlines and configured limits are unchanged.

## Contracts Added or Changed

- Shared lifecycle requests accept a final authority callback; it is never serialized.
- Exact peer-owned conversation reads do not change focus or revision.
- Control previews freeze the invocation revision before confirmation.
- `ControlWorkExtent` accepts a caller-neutral request binding for all eight
  controls and returns the original receipt alongside separately observed status.
- A bound continuation replays its acceptance without re-entering execution.
- Native publication projection and stop intent commit together.
- Receipt-only reads recover native journals without constructing execution ports.
- Historical control acknowledgements preserve replacement conversation state.
- Early replies cannot become unbound authoring/model input or run work in reserved slots.
- Request-capacity rejection rolls back reservation before owner application.
- Inbox processing class and control-reserve membership are retained with input.
- Model context snapshots guard target resolution after external interpretation.
- Receiver application slots and model slots have separate bounded ownership.
- Decision controls seal all displayed guards and attribute the native history to
  the configured operator. Receipt replay cannot exchange their payload.
- Control-result conversation projections retain the original admission reference.

## Files of Interest

- `internal/app/service_workflow_control.go`: existing common control owner.
- `internal/app/telegram_control_*.go`: typed control parsing, previews and effects.
- `internal/core/workflow_control_authority.go`: guarded native lifecycle entry points.
- `internal/app/telegram_control_*test.go`: real HTTP/SQLite, authority, worker checkpoints and decision receipt recovery.
- `internal/core/workflow_decision_control*.go`: shared guarded decision entry point.
- `internal/core/workflow_control_receipt.go`: native binding and atomic receipt files.
- `internal/core/workflow_control_read.go`: receipt-only native recovery.
- `internal/app/service_workflow_control_read.go`: caller-neutral receipt observation.
- `internal/app/telegram_control_stale_receipt_test.go`: accepted and unaccepted expiry/replacement parity.
- `internal/app/telegram_control_recovery_guard_test.go`: revoked authority and acknowledgement races.
- `internal/core/telegram_control_receipt_test.go`: partial-journal decision recovery.
- `internal/core/workflow_control_request.go`: common immutable control binding.
- `internal/core/workflow_lifecycle_control*.go`: lifecycle acceptance and atomic receipts.
- `internal/core/telegram_control_lifecycle_test.go`: task/closing/publication crash boundaries.
- `internal/app/telegram_control_lifecycle_ack_test.go`: five-action acknowledgement recovery and live retry.
- `internal/core/project_registry_read.go`: bounded read-only database contention.
- `internal/core/telegram_control_registry_test.go`: exclusive-writer/read-authority handoff.
- `internal/app/telegram_application_host.go`: bounded asynchronous inbox calls.
- `internal/app/telegram_activation_slots.go`: common receiver call accounting.
- `internal/core/telegram_processing_*.go`: priority, source order and restart classification.
- `internal/core/telegram_intake_reserve.go`: atomic capacity rejections.
- `internal/core/telegram_model_slots.go`: shared provider exclusion and limits.
- `internal/core/telegram_control_*test.go`: reserve, storage failure, model slots and action priority.
- `internal/core/telegram_reply_resolution.go`: exact reply wait and exhausted-send boundary.
- `internal/core/telegram_request_allowance.go`: bounded control request reservations.
- `internal/app/telegram_application_lane.go`: activation class checks before effects.
- `internal/app/telegram_control_activation_test.go`: all ordinary slots occupied with bounded control processing.
- `internal/app/telegram_control_reply_test.go`: actual HTTP binding race and worker/model lane barriers.
- `internal/app/telegram_control_request_capacity_test.go`: durable overflow diagnostics across operation types.
- `internal/app/telegram_control_priority_test.go`: real HTTP receiver under model load.

- `internal/core/workflow_stop_watch.go`: guarded claim-scoped stop observation.
- `internal/core/telegram_control_stop_owner_test.go`: stop, credential and identity-only polling boundaries.
- `internal/core/workflow_capacity_process_test.go`: distinct cross-process slots through eight workers.
- `internal/app/telegram_control_shared_capacity_test.go`: actual Telegram/daemon/manual source overlap and shared capacity reuse.

- `internal/app/telegram_schema_fixture_test.go`: production-migrated immutable schema image with isolated per-fixture databases.

- `internal/core/workflow_claim_read.go`: exact-owner credential lookup with existing scope resolution for aliases and other roots.
- `internal/core/telegram_control_claim_scope_test.go`: colliding invocation IDs cannot cross repository ownership.

## Validation

Activation documentation passed `git diff --check`.

T6.1 checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramControl|WorkflowStop$|WorkflowStopBarrier$|WorkflowRecoveryStop$)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramControl|WorkflowStop$|WorkflowStopBarrier$|WorkflowRecoveryStop$)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The broader command
`go test ./internal/core ./internal/app -run '^Test(TelegramControl|TelegramWork|TelegramConversationCommands|WorkflowControl|WorkflowTerminalControl|ServiceWorkflowContinuation|DaemonWorkflowRevision)' -count=1 -timeout=3m`
failed at the unchanged `TestTelegramWorkResponsive` assertion, which consumed a
queued native progress notice instead of receiver status. The isolated command
`go test ./internal/app -run '^TestTelegramWorkResponsive$' -count=1 -timeout=3m`
passed. No unrelated production change was made for that scheduling observation.

T6.2 checks passed:

- `go test ./internal/core ./internal/app -run '^TestTelegramControl' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramControl' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(ServiceIntegration|ServiceNativeReview|WorkflowReviewApproval|NativeBudgetGrant|NativeDecisionAnswer|WorkflowIntegrationAnswer)' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^TestTelegramControlDecisions$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

These focused checks cover native review/grant/question effects, each stale
displayed fence, final authority loss, changed candidate/base rejection,
disconnected-provider receipt replay, natural-language grants, malformed/spoofed
inputs, duplicate actions and native/SQLite acknowledgement recovery.
The existing service and CLI integration controls retain their prior behavior.

T6.3 progress checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramControl|TelegramReceiver|TelegramAuthoring(Interpretation|Model)|TelegramConversationJourney$|TelegramWorkResponsive$)' -count=1 -timeout=3m`
- `go test ./internal/core -run '^TestTelegramControlModelCapacity$' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramControl|TelegramReceiver|TelegramAuthoring(Interpretation|Model)|TelegramConversationJourney$|TelegramWorkResponsive$)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The combined normal check completed in core 8.665 seconds and app 38.276
seconds; race mode completed in core 111.073 seconds and app 163.656 seconds.
The separate model-capacity check verifies the final fixture cleanup change.
The fixtures exercise
32 reserved inputs and the unchanged 1,024 ceiling, atomic rejection failure,
callback/text/reply priority including worker-capable exclusions, pending-input
reclassification, two concurrent model purposes, preserved retry budgets,
context changes, and real HTTP status/draft cancellation while a model is blocked.
The existing worker-load regression now waits for its status response within its
original two-second deadline while consuming unrelated progress notices.

T6.3 lifecycle progress checks passed:

- `go test ./internal/core ./internal/app -run '^TestTelegramControlLifecycle(Journal|Fence|Ack)$' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^Test(TelegramControl(Lifecycle|Natural|Journal|Decisions|Fences|Guards)|NativeStop|WorkflowClosingRecovery$|WorkflowRecoveryBudget$|WorkflowContinuation|WorkflowTerminalControl$|ServiceWorkflowContinuation|ServiceWorkflowDependency)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramControl(Lifecycle|Natural|Journal|Decisions)|NativeStopPending$|WorkflowClosingRecovery$|WorkflowRecoveryBudget$|WorkflowTerminalControl$|ServiceWorkflowContinuation|ServiceWorkflowDependency)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The combined normal selection completed in core 59.326 seconds and app 20.419
seconds; the focused race selection completed in core 60.569 seconds and app
74.499 seconds. Deterministic fixtures interrupt native journal replacement,
recover task and closing checkpoints, repair a missing submission pointer with
stop intent, reject late authority loss and changed actors, and recover all five
HTTP/SQLite lifecycle acknowledgements after later native revisions. They also
preserve legacy continuation, prerequisite resume and native publication-stop
behavior. The initial acknowledgement fixture timed out because reopened SQLite
owners were scoped to individual subtests; the passing fixture retains one
owner lifecycle across the complete five-action journey.

The live receipt-retry fixture exposed an unconfigured read-only SQLite busy
timeout (`scan project: database is locked`). The reader factory now applies the
existing bounded connection policy; an exclusive-writer fixture proves the
read waits and succeeds after release. The production receiver also retries an
accepted control acknowledgement without repeating its native effect and remains
available for a subsequent status command.

Follow-up checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramReceiver|ProjectRegistry|TelegramControl(RegistryWait|LifecycleAck|ReceiptRetry)$)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramControl(RegistryWait|LifecycleAck|ReceiptRetry)$|TelegramReceiverReadJourney$|ProjectRegistry)' -count=1 -timeout=3m`
- `go test ./internal/core -run '^TestTelegramControlRegistryWait$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The follow-up normal checks completed in core 1.599 seconds and app 3.419 seconds;
race mode completed in core 6.631 seconds and app 20.054 seconds. The separate
registry check verifies the final fixture cleanup refactoring.

T6.3 stale-prompt receipt checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramControl(StaleReceipt|Receipt|Lifecycle|Decisions|Natural|Journal|Fences|Guards)|TelegramOperationAction)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramControl(StaleReceipt|Receipt|Lifecycle|Decisions|Natural|Journal|Fences|Guards)|TelegramOperationAction)' -count=1 -timeout=3m`
- `go test ./internal/app -run '^TestTelegramControlStaleReceipt$' -count=1 -timeout=3m`
- `go test -race ./internal/app -run '^TestTelegramControlStaleReceipt$' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The combined normal checks completed in core 33.258 seconds and app 45.452
seconds; the race checks completed in core 53.931 seconds and app 168.005
seconds. All 18 expiry/replacement/supersession cases remain named and cover
accepted and unaccepted text, callback and reply sources. A shared fixture per
stale condition reduces repeated admission setup; its normal follow-up completed
in app 2.063 seconds and its focused race follow-up in 19.720 seconds. Additional
fixtures prove current authority rejection, original-source acknowledgement followed by a new answer, replacement during
native-to-SQLite acknowledgement, and missing-receipt probes without reservations.
Existing action and lifecycle/decision journal regressions remain green.

The system-installed lint v2.12.2 crashed in static analysis. The repository's
pinned v2.14.0 completed successfully without changing checks or configuration.

T6.3 completion checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramControl(Reply|Lane|AuthoringLane|Request|StaleReceipt|ReceiptAuthority|ReceiptRetry|ModelLoad)|TelegramConversation|TelegramOperationAction|TelegramAuthoring(Bridge|Interpretation|Model))' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^TestTelegramControl(ActivationBound|Reply|Lane|AuthoringLane|Request|StaleReceipt|Reserve$|StorageFailure$|ModelCapacity$)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramControl(ActivationBound|Reply|Lane|AuthoringLane|Request|StaleReceipt|Reserve$|StorageFailure$|ModelCapacity$)' -count=1 -timeout=3m`
- `go test ./internal/core -run '^TestTelegramControlReply' -count=1 -timeout=3m`
- `go test -race ./internal/core -run '^TestTelegramControlReply' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

The broader normal regression completed in core 21.020 seconds and app 24.447
seconds. Final focused normal checks completed in core 2.479 seconds and app
5.291 seconds; the focused race run completed in core 67.497 seconds and app
45.096 seconds. The final reply priority/exhaustion follow-up completed in core
0.358 seconds normally and 6.858 seconds under race instrumentation. Strict
focused lint reports zero issues.

The initial broader race selection timed out in core at three minutes while
repeatedly scanning the 8,192-row retained-capacity fixture; app passed in
145.549 seconds. Fixtures now bulk-seed settled backlogs, keeping real intake
at the last ordinary slot, actual overflow rollback, all 32 pending control
pairs, and first/last retained-ceiling control pairs. The passing focused run
retains the unchanged three-minute deadline and all hard-ceiling assertions.
The reply fixtures also reject foreign peers, preserve sender attempt budgets,
and prevent an exhausted fifth send from deferring an unbound reply forever.

T6.4 focused checks:

```sh
go test ./internal/core ./internal/app -run '^Test(TelegramControlSharedCapacity|WorkflowCapacityOverlap|WorkflowParallelPublication)$' -count=1 -timeout=3m
go test ./internal/core ./internal/app -run '^Test(TelegramControl(StopOwner|SharedCapacity)|Workflow(Stop|StopBarrier|ManualStop|ClosingStop|SubmissionStop)|DaemonWorkflowActiveStop)$' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^Test(TelegramControl(StopOwner|SharedCapacity)|Workflow(CapacityOverlap|ParallelPublication|Stop|StopBarrier|ManualStop|ClosingStop|SubmissionStop)|DaemonWorkflowActiveStop)$' -count=1 -timeout=3m
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app
scripts/check-go-func-signatures.sh
scripts/check-package-readmes.sh
git diff --check
```

Passed. The initial normal capacity/publication selection completed in core
38.300 seconds and app 29.411 seconds. After the owner-scoped polling correction,
the normal stop/capacity selection completed in core 24.991 seconds and app
23.303 seconds; the final race selection passed in core 99.799 seconds and app
58.067 seconds. Strict focused lint reports zero issues.

Initial eight-worker race attempts hit the unchanged thirty-second host gate:
the combined capacity/publication race command passed core in 73.791 seconds
but failed app in 73.075 seconds; an app-only rerun failed in 72.677 seconds.
Staging worker completion in pairs did not remove the failure. The claimed-owner
stop read removes repeated repository-wide lookup from the fifty-millisecond
poll while preserving exclusion, journal guards and final validation. The
passing fixtures retain eight actual overlapping sources and all ownership,
overflow, reuse and host-validation assertions. Owned fixture processes stop
after their checks.

The first controller attempt at `a87bc24384a1b687154bf6df2b53539ecbc4d1a5`
passed strict lint, then `make verify` failed in two database-status tests that
still expected 21 migrations. Slice 6 introduces migration 22. Correction
`8476423bc7aa84d116273e7df120a7916b96284f` keeps exact counts and adds explicit
pending/applied assertions for `20261016000022 add_telegram_processing_lanes`.
The focused command passed in 0.056 seconds:

```sh
go test ./internal/app -run '^TestMaintenanceDBStatusReports(PendingMigrationsWhenDatabaseIsMissing|AppliedMigrations)$' -count=1 -timeout=3m
```

The failed controller evidence is retained; the changed PR HEAD requires fresh
evidence for the complete configured command list.

At `454f5c8761d149ba6ab28b84b96e4158e64c2c6d`, controller `make verify`
passed, followed by the full `TestTelegramControl` normal selection (core
28.353 seconds, app 71.902 seconds). The configured race selection passed core
in 135.658 seconds but reached the app package's unchanged three-minute deadline
during fresh schema migration in `TestTelegramControlAuthoringLane`.

Fixture correction `8b01210ac76dd5a20cbc14e4ec2e34472adc723a` builds one empty
schema image through the production migrator, closes its only connection to
checkpoint the WAL, and retains read-only bytes. Each app fixture creates its
own database file and opens its own actual store through the public constructor.
Receiver state and live connections are never shared. Actual empty-database
migration tests remain independent. No test assertion, command or timeout is
removed or relaxed.

The focused normal and race commands passed in 3.014 and 36.808 seconds,
respectively; strict focused lint and repository style checks also passed:

```sh
go test ./internal/app -run '^TestTelegram(ReceiverReadJourney|Control(AuthoringLane|RequestCapacity|StaleReceipt))$' -count=1 -timeout=3m
go test -race ./internal/app -run '^TestTelegram(ReceiverReadJourney|Control(AuthoringLane|RequestCapacity|StaleReceipt))$' -count=1 -timeout=3m
```

At `67b738083ffb41052f2b6f9da6333e56fb4c8d1b`, controller `make verify`
and the normal control selection passed (core 27.944 seconds, app 64.549 seconds).
Core race passed in 105.537 seconds; app reached the unchanged three-minute
limit near the end of the stale-receipt matrix. This failed evidence is retained.

T6.4 correction `a4faf4367bdaecf28577d4ed569da7d24952e725` uses a direct invocation read only
when a credential's owner root and project exactly match the current store.
Alias and foreign-root callers retain repository scope resolution. Existing
input, lease, token, generation and task-image checks remain unchanged. A
normal CPU profile identified repeated repository-wide lookup in claimed work;
the scope regression includes a foreign repository with a colliding invocation
ID and proves that neither scope check mutates ownership.

Capacity fixtures now exercise actual manual preparation, supervised execution
and exact candidate reconciliation. All eight sources still overlap before
capacity/duplicate assertions. FIFO permits replace busy source wait loops;
manual candidates reconcile after all source groups exit. One native snapshot
asserts every expected invocation, one validated attempt per invocation and no
additional invocation. Ticket-only admission fixtures retain an independent
target and one unselected grouped task; all five multi-extent fixtures retain
the complete plan. No assertion, command or deadline is relaxed.

The following normal/race selection passed in core 25.252/40.208 seconds and
app 31.169/110.928 seconds. Strict focused lint and repository style checks passed.

```sh
go test ./internal/core ./internal/app -run '^Test(TelegramControl(ClaimScope|SharedCapacity|StopOwner|Lifecycle|StaleReceipt)|TelegramWork(Activation|Capacity|OtherOwner)|WorkflowClaim)' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^Test(TelegramControl(ClaimScope|SharedCapacity|StopOwner|Lifecycle|StaleReceipt)|TelegramWork(Activation|Capacity|OtherOwner)|WorkflowClaim)' -count=1 -timeout=3m
```

Controller continuation at `1be7869a1ec8cf068da4195a109fd845638752cb` passed
`make verify` and the full normal control selection (core 25.456 seconds, app
67.980 seconds). Core race passed in 125.900 seconds. App reached the unchanged
three-minute package deadline while `TestTelegramControlSharedCapacity/8` was
running; no assertion failure preceded the deadline. Evidence ID:
`EM6ZSG2SETNYTQQNYDVW2AGADS`.

PR #395 remains open and is not ready for integration. Its final app race
selection still needs to complete within the configured bound, preserving all
cases and assertions. No owned capacity-fixture processes remained after the
failed run. Subsequent PR HEADs require a fresh complete controller receipt;
these partial results cannot establish final readiness.

Authoritative final slice validation is the controller receipt for the clean PR
HEAD, in this order:

```sh
make verify
go test ./internal/core ./internal/app -run '^TestTelegramControl' -count=1 -timeout=3m
go test -race ./internal/core ./internal/app -run '^TestTelegramControl' -count=1 -timeout=3m
git diff --check
```

Live and scheduled exhaustive checks remain separate.

## Risks and Follow-ups

All Slice 6 tasks have focused normal/race and strict-check evidence. Final PR
readiness is established by the exact clean-HEAD controller receipt. Live
activation, user operation and scheduled exhaustive certification remain in
the approved successor work.
