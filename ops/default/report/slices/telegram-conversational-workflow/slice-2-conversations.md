# Slice 2: Conversation Routing

Status: drafting

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-2-telegram-conversations`

PR: pending

## Purpose

Retain independent project-bound sessions and bounded capture without changing
canonical project artifacts or admitting work.

## Delivered Behavior

Peer selection is keyed by bot, chat, topic and actual sender. Explicit project
IDs or unique exact names take precedence, followed by a bound session, peer
selection, authorized default and sole authorized enabled project. Ambiguity and
explicit/reply mismatches return a diagnostic. Existing sessions keep their
project, resolved root identity, namespace and intent when selection changes.
Current registry and allowlists are rechecked before applying queued input.

Capture parts retain source IDs, digests and byte ranges. One transaction
changes capture, acknowledges its inbox source and creates its outbox response.
Replay cannot append again; response-capacity failure rolls the whole change
back. Limits are 16 KiB per part, 64 KiB including separators and 32 parts.
Completion retains text in the preparing phase; cancellation discards unstarted
capture. Neither operation claims artifact generation or execution.

Definite send acknowledgement and exact peer/message/revision reply association
commit together. Replaced questions, changed revisions, terminal captures and
pruned associations cannot route answers. Canonical question bindings retain
session, question, revision and digest independently per conversation. At most
64 reply associations remain active for each conversation.

## Implementation Notes

T2.1 implementation: `8e7faf9bf62c` (full task references are in the tracker).
Migration `20261016000017` extends the existing user database with peers,
conversations, parts and reply associations. Allocation enforces eight active
sessions per peer and 128 per bot after acquiring the SQLite write reservation.
No transaction spans network IO. Registry roots are resolved physically before
freezing their identity. A changed root requires an explicit new session.

## Contracts Added or Changed

`TelegramConversationInput` carries deterministic local operations and current
registry/configuration evidence. `BindTelegramConversationQuestion` provides a
revision-guarded binding for the shared authoring owner; it does not invoke a
provider or answer the canonical question. Capture completion leaves subsequent
authoring to its recorded owner.

## Files of Interest

- `internal/core/telegram_conversation*.go`: peer/project routing, retained
  capture, question guards and acknowledged reply associations.
- `cmd/chr/assets/migration/sqlite/20261016000017-add_telegram_conversations.sql`:
  additive retained state.
- `internal/core/telegram_receiver_outbox.go`: transaction-shared responses and
  definite delivery acknowledgement.

## Validation

T2.1 focused routing and migration checks passed:

- `go test ./internal/core ./internal/app -run '^Test(TelegramConversation|MaintenanceDBStatus)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app` — zero issues.

T2.2: `go test ./internal/core -run '^TestTelegramConversation' -count=1 -timeout=3m` passed, including byte/part limits, replay/reopening, atomic capacity rollback, peer/topic isolation, independent questions, stale revisions and 64-reply retention. Strict focused lint again reported zero issues.

Remaining task and exact candidate evidence will be recorded before PR readiness.
All current fixtures use temporary real SQLite stores and local project roots.

## Risks and Follow-ups

Capture completion retains text for the approved authoring slice. Question
bindings are exercised without calling a provider; canonical question answering
and artifact generation remain owned by the authoring implementation. An
uncertain transport send can repeat a notice while source replay cannot repeat
capture. Live bot/model proof and exhaustive certification remain pending their
recorded validation stages.
