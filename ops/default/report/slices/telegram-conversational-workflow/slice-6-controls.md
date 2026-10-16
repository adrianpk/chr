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

## Implementation Notes

Native lifecycle entry points recheck volatile caller authority while holding
the existing workflow owner lock. Telegram source requests retain the exact
project root, namespace, conversation revision, invocation and mapped operator.
Explicit standalone controls allocate a source-stable confirmation conversation;
status only observes state. No transport execution scheduler is introduced.

T6.2 will complete native request receipts and exact review/grant/decision
bindings before the slice is ready for a PR.

## Contracts Added or Changed

- Shared lifecycle requests accept a final authority callback; it is never serialized.
- Exact peer-owned conversation reads do not change focus or revision.
- Control previews freeze the invocation revision before confirmation.

## Files of Interest

- `internal/app/service_workflow_control.go`: existing common control owner.
- `internal/app/telegram_control_*.go`: typed control parsing, previews and effects.
- `internal/core/workflow_control_authority.go`: guarded native lifecycle entry points.
- `internal/app/telegram_control_*test.go`: real HTTP/SQLite, authority and worker checkpoints.

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

Required slice validation will be recorded through the controller for the final
clean PR HEAD: `make verify`, focused `TestTelegramControl` normal and race
commands, then `git diff --check`. Live and scheduled exhaustive checks remain
separate.

## Risks and Follow-ups

T6.2 through T6.4 remain pending completion. Native lost-acknowledgement
receipts, guarded review/grant/decision bindings, prioritized controls under model
load and shared capacity certification are not established by T6.1 evidence.
