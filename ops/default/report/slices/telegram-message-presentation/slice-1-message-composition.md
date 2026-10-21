# Slice 1: Readable Text and Conversation Messages

Status: delivered
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `feat/slice-1-telegram-message-composition`
PR: `#424`

## Purpose

Make existing Telegram text and conversation messages readable while preserving
request ownership, target selection, authority and operation receipts.

## Delivered Behavior

- Messages use only the necessary introduction, content and next-step blocks.
  Literal Markdown and code use explicit Telegram entities; dynamic delimiters
  remain visible data. Offsets use the final visible UTF-16 text.
- Capture start acknowledges the retained request and makes additional details
  optional. Project names, topics, human states and exact selection commands
  replace routine identifiers, revisions and byte counters. Near-limit notices
  retain the actual capture bounds.
- Capture inspection shows literal retained text. Focus, stale prompts, capture
  errors and cancellation explain their actual effect without retargeting work.
- Confirmed dump generation says the request and details were saved. Canonical
  clarification questions lead with the actual question and preserve reply routing.
- Basic help defers advanced controls to `/help all` and explicit command groups.
  Every native command remains discoverable in bounded views. Unsupported
  commands give a concise `/help` route; empty project access requires local setup.

## Implementation Notes

T1.1 is `ffcbc9b4dabc`; T1.2 is `c5665e602b5f`; T1.3 is `eb7096813576`.
The tracker records their full immutable commit identities.

The shared renderer and immutable text snapshots remain in `internal/core`.
The existing receiver outbox freezes content before HTTP, preserves source replay
and retains definite-rejection-only fallback within its existing attempt allowance.
The additive migration introduces nullable presentation data and its fallback
selection; existing plain and legacy-Markdown rows retain their exact meaning.
Snapshot bytes count toward the existing intake quota. Unknown presentation
versions are refused before transmission.

Visible composed text stays below 1,800 UTF-16 units, including shortening markers.
The renderer reserves 64 units for the existing uncertain-delivery notice. It
bounds source blocks before creating entities and reserves space for the next
instruction. Session lists keep complete target rows and identify omitted entries.

The tracker records outcomes and source/test ownership for families 01-20. Legacy
routing, project ambiguity, capture limits, peer isolation, source replay and
canonical authoring tests supply the relevant guard evidence.

## Contracts Added or Changed

`TelegramMessageBlocks`, `TelegramPresentation` and `TelegramTextEntity` provide
optional composition and a versioned frozen text/format value. Ordinary text is
sent without a markup parser. Literal Markdown/block code uses a `pre` entity;
inline code uses `code`. Plain fallback preserves the exact visible content.

`QueuePresentationResponse` shares the existing transaction, outbox identity,
acknowledgement and quota owner. The receiver API sends explicit entities through
its existing bounded HTTP call. Existing legacy Markdown help remains supported.

## Files of Interest

- `internal/core/telegram_presentation.go` and `telegram_presentation_outbox.go`
- `internal/core/telegram_conversation_presentation.go`
- `internal/app/telegram_help.go` and `telegram_drain.go`
- `internal/app/telegram_authoring_apply.go`
- `internal/app/telegram_presentation_test.go`
- `cmd/chr/assets/migration/sqlite/20261021000025-add_telegram_text_presentation.sql`

## Validation

Successful development checks:

- `go test ./internal/core ./internal/app -run '^TestTelegramPresentation' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^(TestTelegramConversation|TestTelegramHelp|TestTelegramReceiverRead|TestTelegramAuthoringBridge|TestTelegramAuthoringIntent|TestTelegramControlAuthoringLane)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core ./internal/app` — zero remaining issues.
- `go test ./internal/app -run '^(TestMaintenanceDBStatusReports(PendingMigrationsWhenDatabaseIsMissing|AppliedMigrations)|TestTelegramControlModelLoad)$' -count=1 -timeout=3m`
- `git diff --check`

The initial `make verify` at `0c83951f1b2260676cf29d8961e2e7fbe89307eb`
failed on stale migration-count and cancellation/target-diagnostic expectations.
Controller attempt `4AOGSXL2ESCN672UHLH44HBTLO` retains the diagnostics. The
focused correction updates both migration listings for migration 25 and the
expected user-facing notices; all existing state, capture and single-model-call
assertions remain. The focused correction command above passed.

The authoritative final-head evidence belongs to `looper validate` after the
PR-reference follow-up commit. Its ordered commands are `make verify`, the
configured `TestTelegramPresentation(Text|Format|Conversation)` selection,
the affected legacy regression and correction selections above and
`git diff --check` last.
These are local development/HTTP/SQLite checks, not exhaustive scheduled or
real Telegram client evidence.

## Risks and Follow-ups

Mobile/desktop layout and the user-operated journey remain Slice 5 acceptance
work. Later slices own document delivery, guarded action presentation, progress
and notification producers. Existing authoring and workflow owners retain their
current state and authority; this slice does not substitute presentation receipts
for operation or integration evidence.

## Integration

PR [#424](https://forge.adrianpk.com/adrianpk/chr/pulls/424) was merged by the
controller through Rebase + Fast-forward into dev at
`a2bb0974eac21b1fe33cac78e21f4c1845871e96`. Canonical Forgejo, dev ancestry and
Acta merge event `01M5F5AATQPJW7DV1AB99E4Q2A` agree. Controller evidence
`UYX4BPRVESSKGMZMYMQQG72NMA` records all five ordered checks passed on that
exact clean candidate, including `make verify` and whitespace validation.
