# Slice 1: Operation Service Boundary

Status: delivered
Slicing: Application Operation Service Consolidation
Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`
Tracker: `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`
Branch: `feat/slice-1-operation-service-boundary`
PR: #156

## Purpose

Introduce the central application operation service boundary with a low-risk
first migration.

This slice establishes the shape future slices will use to move adapter-facing
workflows out of TUI-specific executor structs.

## Delivered Behavior

- Added a central unexported `service` type in `internal/chrapp`.
- Centralized shared operation dependencies on the service.
- Added an adapter-neutral workspace refresh result.
- Moved workspace refresh graph building behind the service boundary.
- Kept the existing TUI workspace refresh executor as a thin adapter wrapper.
- Preserved current TUI refresh behavior and result mapping.

## Implementation Notes

The service is intentionally unexported for this slice. It only needs to be
used inside `internal/chrapp` while the existing TUI adapter wiring remains in
place.

The first migrated path is workspace refresh because it is a focused operation
with limited side effects. Later slices can move metadata, artifact, ticket,
and work-control operations without redefining the boundary.

## Contracts Added or Changed

- Added `service` as the central operation implementation type.
- Added `WorkspaceRefreshResult` in `internal/chrapp`.
- Added `service.RefreshWorkspace`.
- Kept `chrtui.WorkspaceRefreshExecutor` unchanged.

## Files of Interest

- `internal/chrapp/service.go`
- `internal/chrapp/service_workspace.go`
- `internal/chrapp/service_test.go`
- `internal/chrapp/app.go`

## Validation

- `go test ./internal/chrapp ./internal/chrtui` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- The service is still only used for workspace refresh; later slices need to
  migrate metadata, artifact, ticket, and work-control operations.
