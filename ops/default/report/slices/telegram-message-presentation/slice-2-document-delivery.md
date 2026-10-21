# Slice 2: Document Excerpts and Complete Attachments

Status: delivered
Delivery set: Telegram Message Presentation
Plan: [ops/default/plan/telegram-message-presentation.md](../../../plan/telegram-message-presentation.md)
Tracker: [ops/default/tracker/telegram-message-presentation.md](../../../tracker/telegram-message-presentation.md)
Branch: `feat/slice-2-telegram-document-delivery`
PR: [#425](https://forge.adrianpk.com/adrianpk/chr/pulls/425)

Integration verified: `a5c6a7e068a8e5ec547d13d73d39afe1f0477335` on 2026-10-21.
Exact-head controller receipt `NQQQ4DVYVPESUBCN7TURIYUTUC` records all five
ordered checks passed. Real Telegram client acceptance remains Slice 5 work.

## Purpose

Give Telegram users a short readable view and one complete relevant document,
while retaining the original operation, source version and delivery allowance.

## Delivered Behavior

- Dump, specification, refinement and plan results attach the primary canonical
  Markdown image rather than resending earlier artifacts. Captions use direct
  opening body lines, omit YAML front matter and show literal Markdown safely.
- Explanations attach an identified complete Markdown export with canonical
  sources. Record reads show human titles/states and attach the complete JSON
  export for that page; empty views need no attachment.
- Complete images are frozen before upload. Reopen, source-file edits, uncertain
  sends and plain formatting fallback preserve those bytes and destination.
- Missing/changed sources or full document storage produce explicit guidance.
  Successful domain results keep their original receipt and current conversation
  binding, including when the conversation advances before local completion.
- Upload acknowledgements retain the original Telegram message and file identity.
  Lost, malformed or incomplete acknowledgements retain uncertainty and the
  existing bounded attempt allowance.

## Implementation Notes

The task commits are `0c50fa368801` for T2.1, `bb6a5575d69d` for T2.2,
`c48039eb4af5` for T2.3 and `7e2b7d21e136` for T2.4. Full identities are
recorded in the tracker. The preceding closure commit verifies PR #424's exact
integration and marks Slice 1 delivered.

The existing receiver database owns immutable document rows, presentation hashes,
source/destination/project bindings and upload receipts. Document bytes are at
most 2 MiB, and admission checks the 64 MiB retained-byte quota under the same
SQLite write reservation as source acknowledgement and outbox creation.
A storage failure rolls back document admission; an already confirmed operation
keeps its receipt and queues one truthful text result instead.

The shared renderer limits document captions to 900 UTF-16 units, reserves the
existing retry notice, and calculates literal entities after shortening. Excerpts
use the first six nonempty body lines and at most 600 UTF-16 units including the
marker. Complete attachment bytes remain unchanged.

Uploads use the existing ten-second HTTP budget, inside its sixteen-second
reservation, with bounded multipart buffers and response decoding. Peer/config
and current project-root authority are checked before upload. The uploader never
reads a path or asks Telegram to fetch a private artifact URL.

Cleanup considers at most 20 images in the existing poll transaction. Only aged
acknowledged images without an unresolved delivery or live action lose their
payload after seven days. Source/content hashes remain available. Unknown
presentation versions and inconsistent retained hashes fail before transmission.

## Contracts Added or Changed

`TelegramDocument` carries complete authorized bytes or an identified export;
`TelegramDocumentResponse` pairs them with the short view. `QueueDocumentResponse`
freezes one owner/destination-bound image and its presentation atomically.
Migration 26 is additive: legacy plain/Markdown rows and Slice 1 version-one text
presentations keep their original meaning.

`SendDocument` and `DeliveryDocumentSent` extend the existing receiver API and
outbox owner. Explicit entities, original action keyboard values and plain
fallback use that owner; document delivery does not create a second dispatcher.
The response-only operation field cannot change operation authority or its guard.

Authoring images come from the existing consistent source owner and must match
the confirmed result's source hash. A later edit cannot replace admitted bytes.
Record pages and explanations are explicitly labeled exports, not claimed as a
complete original file of another format.

## Files of Interest

- `internal/core/telegram_document.go`, `telegram_document_outbox.go` and `telegram_document_result.go`
- `internal/core/telegram_document_http.go` and `telegram_receiver_outbox.go`
- `internal/app/telegram_document_delivery.go` and `telegram_document_results.go`
- `internal/app/telegram_authoring_apply.go`, `telegram_explanation_dispatch.go` and `telegram_record_dispatch.go`
- `cmd/chr/assets/migration/sqlite/20261021000026-add_telegram_document_snapshots.sql`
- `internal/core/telegram_document_recovery_test.go` and `telegram_document_migration_test.go`

## Validation

Successful development checks:

- `go test ./internal/core ./internal/app -run '^TestTelegramPresentation(Document|Delivery|Migration)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app -run '^(TestTelegramPresentation(Document|Delivery)|TestTelegramAuthoringBridge|TestTelegramAuthoringExplain|TestTelegramOperationLiveRecords)' -count=1 -timeout=3m`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./internal/core ./internal/app` — zero issues.
- `git diff --check`

The normal focused run, including the real 64 MiB quota fixture, reported
0.608 seconds for core and 0.477 seconds for app, excluding compilation.
The original three-minute package budget is unchanged.

Final PR-head evidence is recorded through `looper validate` after the
PR-reference follow-up. Required order: `make verify`, the configured focused
normal selection, the configured focused race selection, affected legacy
regressions, then `git diff --check`. This local HTTP/SQLite evidence does not
replace scheduled exhaustive validation or real Telegram client acceptance.

## Risks and Follow-ups

A lost upload acknowledgement can cause a repeated attachment within the same
bounded allowance; it cannot repeat a confirmed domain operation. Document-backed
confirmation guards remain Slice 3 work, notifications remain Slice 4 work, and
actual mobile/desktop attachment opening remains Slice 5 acceptance.
