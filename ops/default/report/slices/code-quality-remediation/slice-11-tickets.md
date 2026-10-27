# Slice 11: Ticket, readiness and execution boundaries

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-11-quality-tickets`
PR: pending

## Purpose

Give ticket models/indexing/rules and execution/readiness their actual owners,
finish assigned signatures and typed states, and reconcile the six-parameter phase.

## Delivered Behavior

Ticket and indexed-ticket models, exact rendering, parsing and field/status
validation now have a direct ticket owner. Native status mutation retains its
actual guarded file-write and transition admission. Shared integer decoding lives
with the existing metadata codec. Implementation remains in progress.

## Implementation Notes

The approved T11.1–T11.6 task list owns scope. Moves and naming remain separate
commits with direct consumers and preserved durable/wire contracts.

## Contracts Added or Changed

Sixty-five moved declaration/test preimages reconcile after namespace qualification.
Fields, field order, text/JSON bytes, timestamps, empty values, legacy aliases and
errors are unchanged. Five existing debt identities/type qualifications moved
without adding an allowance; the global debt count remains 207.

## Files of Interest

- `internal/core/ticket/` and native ticket/indexing code.
- `internal/core/work_readiness_service.go` and execution/readiness consumers.
- The approved baseline inventory and core owner ledger.

## Validation

Ticket model/parser/validation move checks passed:

- `go test ./internal/core/ticket/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(Ticket|ValidateIndexedTicket|QualityRecordWire|Board|Leaf)' -count=1 -timeout=3m`
- `make check-go-style`: 9308 declarations, ceiling seven, 207 exact debts and one external exception.

Required final candidate validation remains `make verify`, focused real SQLite/Git
regressions and `git diff --check` through the controller.

The separate naming change exposes ticket.Ticket, Indexed, Renderer, Parser,
ParseOptions, ValidateIndexed, Status* and exact board/status predicates. Native
metadata callers use the shared review-reference policy and renderer directly.
The metadata child exposes OptionalInteger. All 65 source preimages and the
same focused tests/signature-policy checks pass after naming.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/ticket/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues after naming.

Indexed metadata/time prerequisites passed:

- `go test ./internal/core/ticket/... ./internal/core ./internal/app ./scripts/goquality -run '^Test(Ticket|ValidateIndexedTicket|QualityRecordWire|QualityStorage|Board|Leaf)' -count=1 -timeout=3m`
- `make check-go-style`: 9307 declarations, ceiling seven, 205 exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/ticket/... ./internal/core ./internal/app ./scripts/goquality`: zero issues.

Metadata timestamp/round parsing returns one private outcome plus error, retaining
parse order and zero failure results. The existing indexed SQL row assigns its
own time fields instead of passing six positional inputs. timePtr now exists
only in test fixtures. Index and key-sequence constructors consume the existing
database.DataDirectory contract; their IO order, migrations and errors are unchanged.
Two assigned signatures are fixed; the remaining slice work is not closed.

Native index composition prerequisite checks passed:

- `go test ./internal/core ./internal/app ./scripts/goquality -run '^Test(WorkflowIndex|NativeQueueIndex|NativeIntegrationQueue|TicketIndex|QualityRecordWire|QualityStorage|Board|Leaf)' -count=1 -timeout=3m`
- `make check-go-style`: 9308 declarations, ceiling seven, 205 exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/app ./scripts/goquality`: zero issues.

WorkflowTicketIndex composes the existing ticket storage with native workflow
projection capabilities on the same SQLite handle. Native projection methods
retain their actual owner and unchanged SQL, locks, source hashes, bounds and
recovery behavior. Existing database initialization, migration, cleanup and
transactional cache invalidation remain intact. Eight additional constructor/
receiver preimages reconcile, for 73 recorded declarations/tests in total.

Ticket storage/discovery move checks passed:

- `go test ./internal/core/ticket/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(Ticket|ValidateIndexedTicket|QualityRecordWire|QualityStorage|Board|Leaf|WorkflowIndex|NativeQueueIndex|NativeIntegrationQueue)' -count=1 -timeout=3m`
- `make check-go-style`: 9309 declarations, ceiling seven, 205 exact debts and one external exception.

Ticket rows/tags, discovery, rebuilding, reconciliation, status checking and key
reservation now have the ticket owner. Native writes share the same row/tag
operation in the caller's transaction; the native owner retains its workflow
metadata update and transaction. Existing snapshot invalidation, SQL/bind order,
NULL timestamps, migrations, cancellation and parse blockers remain unchanged.
Three storage regressions construct the index through its existing public database
constructor rather than exposing its private SQL implementation. Another 89
preimages reconcile, bringing the move ledger to 162. The separate naming change exposes Index, Discoverer, Rebuilder, Reconciler,
IndexStatusChecker and KeySequenceRepository with direct operation names.
The same focused tests and signature policy pass after naming. Strict lint passes
with zero issues; all 162 declaration preimages reconcile after qualification
and naming.

Single-ticket reconciliation uses a private named parse result with an optional
blocker, preserving successful upserts and the original parse-error mapping.
The retained-row regression proves a blocked refresh changes no indexed values.
This closes another assigned result-shape entry: 204 exact debts remain.

- `go test ./internal/core/ticket/... ./internal/core ./internal/app -run '^Test(TicketReconciler|QualityStorage|QualityRecordWire)' -count=1 -timeout=3m`: passed.
- `make check-go-style`: 9309 declarations, ceiling seven, 204 exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/ticket/... ./internal/core ./internal/app`: zero issues.

The filesystem writer receives its required key allocator explicitly. Application
creation still reserves the durable key before constructing the writer, and the
writer validates cancellation/title/synopsis before calling its allocator. Native
key observation retains the complete native envelope/task decode, Board path and
namespace bounds; it shares the ticket filename/text reader rather than moving
workflow policy into ticket storage. The reader now returns a named number/presence
observation plus error. The obsolete fixed-number option and allocation forwarding
method were removed after their live caller moved to the injected allocator.

- `go test ./internal/core/ticket/... ./internal/core ./internal/app -run '^Test(Ticket|ServiceWorkflowTicket|QualityRecordWire|Board)' -count=1 -timeout=3m`: passed, including native envelope/namespace key recovery.
- `go test ./internal/core -run '^TestTicketRepositoryRejectsMissingSynopsis$' -count=1 -timeout=3m`: passed; invalid synopsis never reaches allocation.
- `make check-go-style`: 9307 declarations, ceiling seven, 203 exact debts and one external exception.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/ticket/... ./internal/core ./internal/app`: zero issues after the allocator/number-reader change.

The filesystem repository and shared key readers now have the ticket owner.
Sixteen more declaration preimages reconcile, for 178 total. The mechanical move
passes the existing ticket creation/native key recovery, Board/application/wire
regressions and signature policy. The separate naming exposes Repository.Create,
CreateOptions/CreateResult, NewRepository, WithKeyPrefix/WithRenderer/
WithRepositoryClock and ReadNumber/NumberObservation/NumberFromText.
The same focused selection and strict lint pass after naming.

- `go test ./internal/core/ticket/... ./internal/core ./internal/app ./scripts/goquality -run '^Test(Ticket|ServiceWorkflowTicket|QualityRecordWire|Board|Leaf)' -count=1 -timeout=3m`: passed.
- `make check-go-style`: 9307 declarations, ceiling seven, 203 exact debts and one external exception.

Twenty-four existing ticket component tests now live beside the ticket owner and
use its public APIs with real SQLite, shipped migrations and temporary files.
The coordinator retains native result-writer/key-observation integration tests and
their private fixtures. No production exports were introduced for tests. Forty-eight
test/fixture declarations reconcile unchanged after signature formatting, for
226 total preimages. timePtr remains exclusively in test fixtures.

- `go test ./internal/core/ticket/... ./internal/core ./internal/app ./scripts/goquality -run '^Test(Ticket|ServiceWorkflowTicket|QualityRecordWire|QualityStorage|Board|Leaf|WorkflowIndex|NativeQueueIndex|NativeIntegrationQueue)' -count=1 -timeout=3m`: passed after test placement.
- `make check-go-style`: 9307 declarations, ceiling seven, 203 exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/ticket/... ./internal/core ./internal/app ./scripts/goquality`: zero issues.

## Risks and Follow-ups

Remaining work includes remaining ticket parser/validation test ownership, execution/readiness
ownership and D03, other assigned signatures, typed state/preparation contracts,
the six-parameter reconciliation and final exact-head controller validation.
No Slice 11 PR has been opened.
