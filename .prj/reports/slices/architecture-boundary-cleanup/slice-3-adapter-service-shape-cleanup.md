# Slice 3: Adapter Service Shape Cleanup

Status: delivered
Slicing: Architecture Boundary Cleanup
Plan: `.prj/plans/slicing-architecture-boundary-cleanup-plan.md`
Tracker: `.prj/trackers/slicing-architecture-boundary-cleanup-tracker.md`
Branch: `feat/slice-3-adapter-service-shape-cleanup`
PR: #211

## Purpose

Move residual CLI workflow logic behind service methods and reduce adapter
surface noise without changing CLI or TUI behavior.

## Delivered Behavior

- Metadata CLI commands now delegate metadata patching to service operations.
- Project metadata and case metadata patch requests preserve flag semantics,
  including unset fields and empty explicitly supplied values.
- Case metadata creation fallback and slug validation now live with the metadata
  service boundary.
- Shared CLI scope helpers use neutral workspace naming.
- TUI adapters are split into domain-focused files for artifacts, dump/quiz,
  loaders, spec/plan/ticket, and work actions.

## Implementation Notes

- `PatchProjectMetadata` and `PatchCaseMetadata` centralize metadata load,
  patch, fallback, and save behavior.
- CLI metadata code now only parses command flags, validates positional input,
  calls service operations, and prints results.
- Adapter splitting was behavior-preserving; constructor and executor names did
  not change.

## Contracts Added or Changed

- Added `ProjectMetadataPatchRequest`.
- Added `CaseMetadataPatchRequest`.
- Replaced `quizScopeFlags` and `quizScopeFromCommand` with
  `workspaceScopeFlags` and `workspaceScopeFromCommand`.
- Replaced `internal/chrapp/tui_adapters.go` with domain-specific adapter
  files.

## Files of Interest

- `internal/chrapp/service_metadata.go`
- `internal/chrapp/metadata_cli.go`
- `internal/chrapp/quiz_cli.go`
- `internal/chrapp/spec_plan_cli.go`
- `internal/chrapp/tui_adapters_artifact.go`
- `internal/chrapp/tui_adapters_dump_quiz.go`
- `internal/chrapp/tui_adapters_loaders.go`
- `internal/chrapp/tui_adapters_spec_plan_ticket.go`
- `internal/chrapp/tui_adapters_work.go`

## Validation

- `go test ./internal/chrapp`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

None recorded.
