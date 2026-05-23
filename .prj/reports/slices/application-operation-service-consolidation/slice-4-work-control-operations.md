# Slice 4: Work Control Operations

Status: delivered
Slicing: Application Operation Service Consolidation
Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`
Tracker: `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`
Branch: `feat/slice-4-work-control-operations`
PR: #159

## Purpose

Move execution and daemon work controls into the shared application operation
service.

This keeps TUI confirmation and rendering in the TUI adapter while ticket
execution, ready-loop execution, daemon stop requests, and work-state reloads
run through reusable service operations.

## Delivered Behavior

- Added service-level selected-ticket execution.
- Added service-level one-ticket ready-loop execution.
- Added service-level daemon stop request.
- Centralized ready-loop ticket selection for migrated work operations.
- Centralized ticket reload after work actions.
- Kept TUI work executor as a thin adapter wrapper.
- Kept TUI observability view assembly in the TUI adapter boundary.

## Implementation Notes

Work operation result types live in `internal/chcapp` and do not depend on
`internal/chctui`.

The service continues to use the existing execution preflight, harness
invocation, loop execution, daemon state repository, and ticket index helpers.
The TUI wrapper maps the service result into `chctui.WorkActionResult` and
reloads the TUI observability view.

## Contracts Added or Changed

- Added `WorkOperationResult`.
- Added `service.StartSelectedTicket`.
- Added `service.StartReadyTicketLoop`.
- Added `service.RequestDaemonStop`.
- Added service-level ready-loop ticket selection and work-state reload helpers.
- Changed TUI work executor to delegate to the operation service.

## Files of Interest

- `internal/chcapp/service_work.go`
- `internal/chcapp/app.go`
- `internal/chcapp/service_test.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Full selected-ticket execution and ready-loop execution remain covered by
  existing integration paths; this slice adds focused direct service coverage
  for daemon stop only to avoid invoking external runtime harnesses in unit
  tests.
