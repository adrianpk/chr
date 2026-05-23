# Slice 3: Artifact And Ticket Operations

Status: delivered
Slicing: Application Operation Service Consolidation
Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`
Tracker: `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`
Branch: `feat/slice-3-artifact-ticket-operations`
PR: #158

## Purpose

Move artifact creation and selected-ticket mutation behavior into the shared
application operation service.

This keeps TUI confirmation and rendering in the TUI adapter while the actual
ticket and authoring workflows run through reusable application operations.

## Delivered Behavior

- Added service-level default ticket creation.
- Added service-level TUI authoring session creation.
- Added service-level selected-ticket comment addition.
- Added service-level selected-ticket status update.
- Centralized ticket lookup and ticket reload paths used by these operations.
- Kept TUI artifact and ticket executors as thin adapter wrappers.
- Kept TUI document reload in the TUI adapter because `DocumentView` is a UI
  model.

## Implementation Notes

Artifact and ticket operation result types live in `internal/chcapp` and do not
depend on `internal/chctui`.

The service still uses existing domain services and repositories for ticket
creation, comments, status updates, authoring sessions, and ticket index
refresh.

Unsupported TUI actions remain rejected by the adapter wrapper because those
commands are currently TUI command-palette affordances without implemented
operation workflows.

## Contracts Added or Changed

- Added `ArtifactOperationResult`.
- Added `service.CreateDefaultTicket`.
- Added `service.CreateTUIAuthoringSession`.
- Added `TicketOperationResult`.
- Added `service.AddSelectedTicketComment`.
- Added `service.MoveSelectedTicketStatus`.
- Added shared service ticket lookup and list helpers for migrated operations.

## Files of Interest

- `internal/chcapp/service_artifact_ticket.go`
- `internal/chcapp/app.go`
- `internal/chcapp/service_test.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Work-control operation migration remains for the next slice.
- Unsupported checklist and draft/context quiz commands still need dedicated
  implementation slices before they can become service operations.
