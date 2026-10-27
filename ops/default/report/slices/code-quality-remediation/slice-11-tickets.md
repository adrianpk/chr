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

## Risks and Follow-ups

T11.1–T11.6 remain open; no Slice 11 PR is prepared yet.
