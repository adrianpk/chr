# Slice 6: CLI Convergence Pass

Status: reviewing
Slicing: Application Operation Service Consolidation
Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`
Tracker: `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`
Branch: `feat/slice-6-cli-convergence-pass`
PR: pending

## Purpose

Route low-risk overlapping CLI ticket commands through the shared application
operation service.

This proves the service is a common operation layer rather than only a TUI
extraction.

## Delivered Behavior

- Added a generic service ticket creation operation.
- Added a generic service ticket comment operation.
- Added a generic service ticket status operation.
- Kept the existing TUI default-ticket and selected-ticket methods by
  delegating them to the generic service operations.
- Routed `chc ticket new` through the service.
- Routed `chc ticket comment` through the service.
- Routed `chc ticket status` through the service.
- Preserved current CLI output contracts.

## Implementation Notes

The migration intentionally targets the smallest obvious overlap between CLI
and TUI operation paths: ticket creation, comments, and status changes.

The CLI remains responsible for parsing command flags and printing its existing
human output. The service owns the underlying ticket repository lookup, write,
and reload behavior.

## Contracts Added or Changed

- Added `CreateTicketRequest`.
- Added `CreateTicketResult`.
- Added `service.CreateTicket`.
- Added `TicketCommentRequest`.
- Added `service.AddTicketComment`.
- Added `service.MoveTicketStatus`.
- Changed `service.CreateDefaultTicket`, `service.AddSelectedTicketComment`,
  and `service.MoveSelectedTicketStatus` to delegate to generic operations.

## Files of Interest

- `internal/chcapp/service_artifact_ticket.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/service_test.go`

## Validation

- `go test ./internal/chcapp -run 'TestTicket(New|Comment|Status)'` passed on
  2026-05-23.
- `go test ./internal/chcapp ./internal/chctui` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Broader CLI convergence remains intentionally out of scope for this slice to
  avoid a wide command rewrite.
