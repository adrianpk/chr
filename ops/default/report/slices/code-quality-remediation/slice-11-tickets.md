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
only in the test fixture. Index and key-sequence constructors consume the existing
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

## Risks and Follow-ups

T11.1–T11.6 remain open; no Slice 11 PR is prepared yet.
