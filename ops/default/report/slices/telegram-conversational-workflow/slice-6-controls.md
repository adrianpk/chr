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

## Implementation Notes

Native lifecycle entry points recheck volatile caller authority while holding
the existing workflow owner lock. Telegram source requests retain the exact
project root, namespace, conversation revision, invocation and mapped operator.
Explicit standalone controls allocate a source-stable confirmation conversation;
status only observes state. No transport execution scheduler is introduced.

Decision request identities reserve their exact payload under native exclusion.
They use the existing bounded operation-binding and receipt directories, with a
separate control schema and checksum. Native review, budget and question owners
include the receipt in their existing commit transaction. Volatile caller
authority is checked again before that commit. Legacy callers retain their
existing API; a final decision authority callback requires a request binding.

## Contracts Added or Changed

- Shared lifecycle requests accept a final authority callback; it is never serialized.
- Exact peer-owned conversation reads do not change focus or revision.
- Control previews freeze the invocation revision before confirmation.
- `ControlWorkExtent` accepts a caller-neutral request binding for review, grants
  and answers and returns its original receipt alongside current status.
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
- `internal/core/telegram_control_receipt_test.go`: partial-journal fault recovery.

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

Required slice validation will be recorded through the controller for the final
clean PR HEAD: `make verify`, focused `TestTelegramControl` normal and race
commands, then `git diff --check`. Live and scheduled exhaustive checks remain
separate.

## Risks and Follow-ups

T6.3 and T6.4 remain pending completion. Priority under model load, reserved
control capacity, expiry/replacement parity, interrupted lifecycle-control result
recovery and shared capacity certification remain outside the completed task
evidence. The full Slice 6 PR and exact-HEAD controller checks remain pending.
