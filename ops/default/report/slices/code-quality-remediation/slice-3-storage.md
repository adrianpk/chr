# Slice 3: Stable records and positional SQL

Status: delivered
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-3-quality-storage`
PR: `#440`

## Purpose

Group durable and public records and make wide SQLite row order explicit while
preserving retained bytes, identities, defaults and transaction ownership.

## Delivered Behavior

WorkflowInvocation, WorkflowMergerInput, IndexedTicket and WorkflowWorkStatus
use cohesive value groups. All 164 original fields retain their names, leaf
tags, omission and ordering. Frozen zero, empty and populated JSON fixtures
compare exact bytes, including the invocation file's owner/request envelope.
Updated constructors retain their original values and promoted field access.

Fourteen wide SQL boundaries across eight tables now use typed row mappings
and one static ordered column definition per table. Values and scan targets
stay together. Execution and notification updates use the same row values with
the ID bound last. Ticket upsert and read share their existing row holder;
conversation scans preserve their COALESCE defaults. Inbox and both document
admissions retain their specific ownership shapes and conditional writes.

Real SQLite tests compare every stored column with pre-refactor writer images,
reopen a populated old database and exercise NULL/empty/default distinctions.
Distinct peer integers and independent source/content hashes expose same-type
swaps. Injected failures after multi-record writes start prove rollback retains
the captured data and durable owner images.

## Implementation Notes

Contiguous value embeddings preserve the original wire order without promoted
serialization methods. The outer invocation file keeps its owner and request
metadata; existing unknown-field retention and canonical hashing paths remain
at their original owners.

SQL identifiers stay static and values stay bound. Only the identical document
quota predicate is shared. The receiver owner_kind remains its original static
literal. SET/WHERE filters, write reservations, quota checks and transaction
commit/rollback owners remain unchanged. No schema migration is added.

The legacy SQLite fixture was captured before changing SQL mappings. Its
provenance records the producer, digest, populated table counts and completed
synthetic project parents; all eight original mapped images were preserved.
Foreign-key checks pass. Validation copies fixtures rather than regenerating
them. The root file/declaration ledger includes the seven new row files.

## Contracts Added or Changed

Internal Go field construction now names cohesive groups, and wide persistence
operations use table-owned row values. Durable JSON, hash inputs, state
spellings, enum/time/JSON decoding, SQL nullability and public status encoding
retain their existing contracts. The signature policy remains at ceiling seven
with 517 exact debt entries and one external-interface exception.

## Files of Interest

- `internal/core/workflow_invocation.go`, `workflow_merger.go`,
  `indexed_ticket.go` and `internal/app/service_workflow_status.go`: value groups.
- `internal/core/execution_row.go`, `remote_command_row.go`,
  `notification_event_row.go`, `indexed_ticket_row.go`: paired row mappings.
- `internal/core/telegram_conversation_row.go`, `telegram_update_row.go`,
  `telegram_document_row.go`: retained transport projections and admissions.
- `internal/core/quality_storage_*_test.go` and `testdata/quality-storage/`:
  old database, stored-column comparisons, defaults and rollback faults.
- `internal/core/quality_record_wire_test.go`,
  `internal/app/quality_record_wire_test.go` and adjacent `testdata/quality-records/`:
  frozen serialization fixtures and original field manifests.
- [Original-to-grouped fields](../../../plan/code-quality-remediation-record-fields.csv).
- [Every wide SQL boundary](../../../plan/code-quality-remediation-sql-rows.csv).
- [Compatibility model](../../../spec/code-quality-remediation-model.md).

## Validation

- `go test ./internal/core ./internal/app -run '^TestQualityRecordWire$' -count=1 -timeout=3m`
  — passed before and after grouping, with exact frozen bytes.
- `go test ./internal/core ./internal/app ./internal/tui -run 'Test(QualityRecordWire|WorkflowMerger|WorkflowInvocation|TicketIndex|WorkflowWorkStatus)' -count=1 -timeout=3m`
  — passed for affected callers and record boundaries.
- `go test ./internal/core -run '^TestQualityStorage' -count=1 -timeout=3m`
  — passed for legacy reopening, complete column images, defaults, distinct
  identifiers/hashes and injected rollback.
- `go test ./internal/core -run '^Test(QualityStorageBindings|ExecutionRepository)' -count=1 -timeout=3m`
  — passed after execution updates moved to shared row values.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/...`
  — passed with zero issues.
- `go run ./scripts/goquality` — passed; 517 debt entries, one exception.

Final clean-head validation runs through the controller after the PR-reference
commit: make verify, the PR's explicit focused record/storage command without
short mode, then git diff --check. The controller retains the exact HEAD,
ordered commands, outcomes and timings. That record and PR Validation own final
candidate results.

## Risks and Follow-ups

None recorded.

Integration: PR #440 was verified merged into dev at `fec20d73396156cf8f389fb11e68de8ca0cadcf1`. Final candidate checks passed in controller order:

- `make verify` — passed.
- `go test ./internal/core ./internal/app -run '^Test(QualityRecordWire|QualityStorage|WorkflowMerger|WorkflowInvocation|TicketIndex|ExecutionRepository|RemoteCommandRecordRepository|NotificationEventRepository|TelegramPresentation|TelegramReceiverBatch|TelegramReceiverMigration|TelegramConversation)' -count=1 -timeout=3m` — passed.
- `git diff --check` — passed.
