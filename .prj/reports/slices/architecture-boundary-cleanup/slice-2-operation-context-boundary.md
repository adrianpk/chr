# Slice 2: Operation Context Boundary

Status: reviewing
Slicing: Architecture Boundary Cleanup
Plan: `.prj/plans/slicing-architecture-boundary-cleanup-plan.md`
Tracker: `.prj/trackers/slicing-architecture-boundary-cleanup-tracker.md`
Branch: `feat/slice-2-operation-context-boundary`
PR: #210

## Purpose

Move model-backed and generation-related service operations to
operation-scoped context so CLI, TUI, and tests pass cancellation and timeout
parents explicitly.

## Delivered Behavior

- Spec generation receives the caller context at the service boundary.
- Plan generation, regeneration, and tracking materialization receive the
  caller context at the service boundary.
- Generation runtime timeout wrapping now derives from the operation context
  passed to the service method.
- CLI generation commands pass their command context directly into service
  operations.
- TUI generation adapters pass the active TUI context into service operations.
- Obsolete `contextcheck` suppressions were removed from generation command
  handlers.

## Implementation Notes

- The service still owns timeout duration and generation guard behavior.
- The caller now owns the parent context used by those guarded generation
  operations.
- Existing behavior remains unchanged apart from the corrected context
  boundary.

## Contracts Added or Changed

- `GenerateSpec(ctx, request)` replaces `GenerateSpec(request)`.
- `GeneratePlan(ctx, request)` replaces `GeneratePlan(request)`.
- `RegeneratePlan(ctx, request)` replaces `RegeneratePlan(request)`.
- `MaterializePlanTracking(ctx, request)` replaces
  `MaterializePlanTracking(request)`.
- `withGenerationRuntime(ctx, kind, scope, run)` now derives its timeout from
  the supplied parent context.

## Files of Interest

- `internal/chcapp/generation_runtime.go`
- `internal/chcapp/service_spec.go`
- `internal/chcapp/service_plan.go`
- `internal/chcapp/spec_plan_cli.go`
- `internal/chcapp/tui_adapters.go`

## Validation

- `go test ./internal/chcapp`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

None recorded.
