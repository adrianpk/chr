# Slice 2: Workspace And Metadata Operations

Status: delivered
Slicing: Application Operation Service Consolidation
Plan: `.prj/plans/slicing-application-operation-service-consolidation-plan.md`
Tracker: `.prj/trackers/slicing-application-operation-service-consolidation-tracker.md`
Branch: `feat/slice-2-workspace-metadata-operations`
PR: #157

## Purpose

Move workspace refresh, Project save, and Case save behavior into the shared
application operation service.

This reduces TUI-specific operation ownership while preserving the existing TUI
executor interfaces as adapter wrappers.

## Delivered Behavior

- Kept workspace refresh behind the shared service boundary.
- Added service-level Project metadata save operation.
- Added service-level Case metadata save operation.
- Kept TUI Project and Case executors as thin mapping wrappers.
- Preserved existing TUI Project and Case save behavior.
- Added direct service tests for workspace refresh, Project save, and Case save.

## Implementation Notes

The operation service continues to be unexported inside `internal/chcapp`.
The new metadata operation result types live next to their service methods and
do not import TUI types.

The TUI-facing executors still satisfy `internal/chctui` interfaces, but they no
longer own the metadata save workflows.

## Contracts Added or Changed

- Added `ProjectMetadataSaveResult`.
- Added `service.SaveProjectMetadata`.
- Added `CaseMetadataSaveResult`.
- Added `service.SaveCaseMetadata`.
- Changed TUI metadata executors to delegate to the operation service.

## Files of Interest

- `internal/chcapp/service_metadata.go`
- `internal/chcapp/service_test.go`
- `internal/chcapp/app.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Artifact, ticket, and work-control operation migration remains for later
  slices.
