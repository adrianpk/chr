# Slice 2: Conversation Routing

Status: delivered

Delivery set: telegram-conversational-workflow

Plan: [Telegram Conversation Delivery](../../../plan/telegram-conversational-workflow.md)

Tracker: [Telegram Conversation Tracker](../../../tracker/telegram-conversational-workflow.md)

Branch: `feat/slice-2-telegram-conversations`

PR: [#391](https://forge.adrianpk.com/adrianpk/chr/pulls/391)

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

The live command registry exposes project selection, prepare/develop capture,
sessions/focus, explicit replies, bounded dump views and done/cancel. Help uses
the same descriptors as parsing. Unknown slash commands return help without
service/model interpretation. Valid replies outrank focus; ambiguous plain text
asks for a session. Reads use the existing central application service.

## Implementation Notes

PR #391 merged through Rebase + Fast-forward into dev on 2026-10-16 at
`f8f33d844c083997ffb0e479e9c22068e5c30e03`. The controller receipt records all
four ordered checks below passed on that clean HEAD. Acta merge event:
`01M526ENWR8E56B4R4T339HTBT`.

T2.1 implementation: `8e7faf9bf62c`; T2.2: `9946757b0528`. T2.3: `36f31feaa0c0c0d5b52e9d80071322a8574744b5`, adding the
live command adapter and HTTP journey. The tracker records task references.
Migration `20261016000017` extends the existing user database with peers,
conversations, parts and reply associations. Allocation enforces eight active
sessions per peer and 128 per bot after acquiring the SQLite write reservation.
No transaction spans network IO. Registry paths and their resolved physical roots both contribute to the
frozen binding digest. A changed root requires an explicit new session.

## Contracts Added or Changed

`TelegramConversationInput` carries deterministic local operations and current
registry/configuration evidence. `BindTelegramConversationQuestion` provides a
revision-guarded binding for the shared authoring owner; it does not invoke a
provider or answer the canonical question. Capture completion leaves subsequent
authoring to its recorded owner.

## Files of Interest

- `internal/app/telegram_conversation.go`: deterministic command descriptors,
  parsing and live dispatch through current central registry reads.
- `internal/app/telegram_conversation_test.go`: HTTP/SQLite journey, decoder
  bounds, duplicate replay and root-change diagnostics.

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

T2.3 focused normal and race checks passed:

- `go test ./internal/core ./internal/app -run '^TestTelegramConversation' -count=1 -timeout=3m`
- `go test -race ./internal/core ./internal/app -run '^TestTelegramConversation' -count=1 -timeout=3m`

Strict focused lint reported zero issues after formatting. The fixtures cover
TG-06/07/08 and the session/topic portion of TG-25: all project and conversation
precedences, exact names and ambiguity, two projects, group users and topics,
independent questions, root/alias changes, revocation, all byte/part bounds,
concurrent replay through separate SQLite handles, stale/late/pruned replies
and rollback when association retention or response capacity fails. The live
HTTP journey checks that no project artifacts or executions are created.

Final readiness evidence was recorded through `looper validate` after the PR
reference commit, binding the integrated clean HEAD to this successful ordered list:

1. `make verify`
2. `go test ./internal/core ./internal/app -run '^TestTelegramConversation' -count=1 -timeout=3m`
3. `go test -race ./internal/core ./internal/app -run '^TestTelegramConversation' -count=1 -timeout=3m`
4. `git diff --check`

All fixtures use temporary real SQLite stores, local roots and owned HTTP
servers. The checks do not establish live-bot or scheduled nightly certification.

## Risks and Follow-ups

Capture completion retains text for the approved authoring slice. Question
bindings are exercised without calling a provider; canonical question answering
and artifact generation remain owned by the authoring implementation. An
uncertain transport send can repeat a notice while source replay cannot repeat
capture. Live bot/model proof and exhaustive certification remain pending their
recorded validation stages.
