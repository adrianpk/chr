# Slice 6: Guarded Controls and Capacity

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-6-telegram-controls`

PR: pending

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

T6.3 lifecycle progress extends native request receipts to pause, cancel,
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

## Contracts Added or Changed

- Shared lifecycle requests accept a final authority callback; it is never serialized.
- Exact peer-owned conversation reads do not change focus or revision.
- Control previews freeze the invocation revision before confirmation.
- `ControlWorkExtent` accepts a caller-neutral request binding for all eight
  controls and returns the original receipt alongside separately observed status.
- A bound continuation replays its acceptance without re-entering execution.
- Native publication projection and stop intent commit together.
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
- `internal/app/telegram_control_priority_test.go`: real HTTP receiver under model load.

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

Required slice validation will be recorded through the controller for the final
clean PR HEAD: `make verify`, focused `TestTelegramControl` normal and race
commands, then `git diff --check`. Live and scheduled exhaustive checks remain
separate.

## Risks and Follow-ups

T6.3 remains in progress: receipt-only recovery through expired or replaced
Telegram prompts, final text/callback expiry parity, late reply-binding races
and request-capacity diagnostics still require completion. Native lifecycle
receipts and their focused crash/acknowledgement proof are implemented; the
adapter must still distinguish historical receipt recovery from fresh effects
when a prompt has become stale. T6.4 shared serial/parallel capacity and
cross-adapter certification remain pending. Focused model-load and intake-reserve
evidence does not complete either task. The Slice 6 PR and exact-HEAD controller
checks remain pending.
