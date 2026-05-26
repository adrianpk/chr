# Slice 1: Dependency Context Cleanup

Status: delivered
Slicing: Architecture Boundary Cleanup
Plan: `.prj/plans/slicing-architecture-boundary-cleanup-plan.md`
Tracker: `.prj/trackers/slicing-architecture-boundary-cleanup-tracker.md`
Branch: `feat/slice-1-dependency-context-cleanup`
PR: #209

## Purpose

Remove hidden application startup dependency state around embedded assets while
preserving the existing CLI, TUI, service, ticket, daemon, notification, and
maintenance behavior.

## Delivered Behavior

- Application startup now creates explicit `appDependencies` from the embedded
  asset filesystem passed to `Run`.
- Root CLI command construction stores those dependencies on the command tree
  and passes them directly into TUI startup paths.
- Project context resolution receives the asset filesystem explicitly instead
  of reading package-level mutable state.
- Services receive the asset filesystem at construction time and reuse it for
  repository-backed operations.
- Ticket index setup is shared through a helper to keep the command handler
  within strict lint limits.

## Implementation Notes

- `internal/chcapp` no longer declares a package-level `assetsFS`.
- Commands that need repository or migration-backed state resolve application
  dependencies from the root command metadata.
- Non-command service and notification paths carry the asset filesystem through
  explicit function parameters.
- TUI reload paths use the service-owned asset filesystem when rebuilding
  observability state.

## Contracts Added or Changed

- `newRootCommand` now requires `appDependencies`.
- `newCommandProjectContextResolver` now requires an asset filesystem.
- `newService` now requires an asset filesystem.
- Ticket execution helper functions now receive an asset filesystem when they
  create execution, loop, notification, or index repositories.

## Files of Interest

- `internal/chcapp/app.go`
- `internal/chcapp/project_context.go`
- `internal/chcapp/service.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/daemon.go`
- `internal/chcapp/maintenance.go`
- `internal/chcapp/notify.go`
- `internal/chcapp/notification_emit.go`

## Validation

- `go test ./internal/chcapp`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

None recorded.
