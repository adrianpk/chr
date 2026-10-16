# Slice 4: Conversational Authoring

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-4-telegram-authoring`

PR: pending

## Purpose

Deliver bounded conversational intent and configured-provider preparation through
the existing shared authoring and artifact owners.

## Delivered Behavior

T4.1 routes ordinary messages through a typed resolver whose available operations
come from the same registries as deterministic controls. Leading slash commands
bypass interpretation. Conversation and record dispatch use the existing typed
handlers, current authorization and shared mutation receipts. Unknown operations,
malformed decisions, missing/null fields, duplicate fields, unknown arguments and
ambiguous targets produce correction prompts without guessed effects.

The resolver receives bounded allowed-project names/IDs and peer-local session
references, excluding transport identities and hidden transcripts. It retains
its proposed operation, configured operator and frozen target before dispatch.
Root, namespace, conversation revision and current authorization are rechecked.
Reopening cannot select a second operation or redirect the frozen namespace.

## Implementation Notes

The shared structured provider uses the existing Codex agent runner or OpenAI
Responses boundary. `CHR_CONVERSATION_GENERATION_BACKEND=openai` selects OpenAI;
otherwise Codex is used. Existing provider model, command and credential settings
retain their meaning. One provider instance permits two in-flight calls, each
with a 60-second deadline, at most 128 KiB of input and 64 KiB of accepted JSON
output. HTTP envelopes are read within 1 MiB. No SQLite transaction spans a
provider call.

Migration 19 adds durable interpretation attempt, retry, sanitized failure and
decision fields. An OS-backed bot-scoped interpretation lock cannot be stolen by
a wall-clock jump. Each attempt is reserved before invocation. Five failures
exhaust that source's budget; restart cannot reset it. Capacity, timeout and
cancellation have separate sanitized failure classifications. Retained decisions
count toward the existing logical transport capacity limit.

The private retained decision format is `chr.telegram.intent.v1`. Record payloads
are decoded to the same strict concrete requests used by slash commands; model
output cannot select an actor, command, executable or shared owner. Existing
transport tests use a deterministic capture provider, while new intent tests
exercise the actual configured HTTP boundary.

## Contracts Added or Changed

- `StructuredModelRequest`, `StructuredModelProvider` and
  `ConfiguredStructuredModel`: bounded configured-provider invocation.
- `ConversationIntentContext` and `ConversationIntent`: a typed proposal with
  registry-owned operation IDs and independently validated arguments.
- `ResolveTelegramInterpretation`: durable decisions and retained retry budgets.
- `TelegramIntentContext` and `ResolveTelegramConversationTarget`: exact-peer
  routing snapshots through the existing conversation owner.
- Strict operation JSON decoding rejects duplicate and unknown fields, excessive
  nesting, trailing input and non-object payloads.

## Files of Interest

- `internal/app/service_conversation_intent.go`: adapter-neutral resolution.
- `internal/app/telegram_intent*.go`: registry mapping, frozen guards and dispatch.
- `internal/core/structured_model*.go`: configured model boundary and capacity.
- `internal/core/telegram_interpretation.go`: retained model attempts and results.
- `internal/core/telegram_authoring*_test.go` and
  `internal/app/telegram_authoring_intent_test.go`: real SQLite, HTTP, replay and
  authority regressions.

## Validation

Successful T4.1 task checks:

- `go test ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^Test(TelegramAuthoring|TelegramConversation|TelegramOperation|TelegramReceiver|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app`
- `scripts/check-go-func-signatures.sh`
- `scripts/check-package-readmes.sh`
- `git diff --check`

Fixtures prove slash/NL start, capture, completion and record parity; correction
for malformed or unsupported proposals; unchanged conversations after provider
failure; deterministic status during failure; revoked authorization and changed
registry roots during model calls; frozen namespace and decision after reopening;
five retained attempts; cross-owner interpretation exclusion; provider capacity
and cancellation; and no native work admission.

Final clean-HEAD controller validation remains pending completion of all four
Slice 4 tasks and committed canonical PR references. These task fixtures are not
live-provider, live-bot or exhaustive scheduled evidence.

## Risks and Follow-ups

T4.2 through T4.4 remain: real shared authoring continuations and question/turn
bindings, retained explain/prepare/develop authority, and canonical artifact
provenance, guarded refinement and crash replay. The report remains drafting.
