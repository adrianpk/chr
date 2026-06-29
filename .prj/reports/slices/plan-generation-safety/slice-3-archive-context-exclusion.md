# Slice 3: Archive Context Exclusion

Status: delivered
Slicing: Plan Generation Safety
Plan: `.prj/plans/slicing-plan-generation-safety-plan.md`
Tracker: `.prj/trackers/slicing-plan-generation-safety-tracker.md`
Branch: `feat/slice-3-archive-context-exclusion`
PR: `#200`

## Purpose

Keep archived generated plans and specs recoverable on disk while excluding
them from active workspace context, TUI document listings, and generation
provenance.

## Delivered Behavior

- `.chr/plans/archive/**` is ignored by the workspace graph builder.
- `.chr/specs/archive/**` is ignored by the workspace graph builder.
- Archived generated documents are not included in workspace index lookups.
- Plan and spec generation context assembled from the workspace index excludes
  archived generated documents.
- Generated spec and plan provenance excludes archived generated document
  paths.
- The TUI document loader hides archived generated plan and spec documents.

## Implementation Notes

- Archive exclusion is implemented at the workspace graph layer so application
  services and downstream indexes inherit the same active-context boundary.
- The TUI document loader has a matching filter because it also scans document
  roots directly from disk.
- Archive browsing and recovery UI remain out of scope for this slice.

## Contracts Added or Changed

- Added active-context exclusion for generated archive directories:
  `.chr/plans/archive/**` and `.chr/specs/archive/**`.
- Added TUI document-list exclusion for the same generated archive directories.

## Files of Interest

- `internal/chr/workspace_graph.go`
- `internal/chr/workspace_index_test.go`
- `internal/chrapp/service_plan_test.go`
- `internal/chrapp/service_spec_test.go`
- `internal/chrtui/documents.go`
- `internal/chrtui/model_test.go`

## Validation

- `go test ./internal/chr`: pass
- `go test ./internal/chrapp`: pass
- `go test ./internal/chrtui`: pass
- `go test ./...`: pass
- `make verify`: fail

`make verify` fails in `lint-strict` on existing lint debt across the plan,
spec, and TUI areas. The current failure set includes pre-existing complexity,
function length, unused function, errcheck, formatting, and whitespace rules.

## Commits

- `0f97588 docs(plan): select slice 3 archive context exclusion`
- `c8605aa fix(index): exclude archived generated documents`
- `f554174 test(plan): verify archived documents stay out of generation`
- `8d658e9 fix(tui): hide archived generated documents`

## Risks and Follow-ups

- Archived generated documents remain on disk and recoverable manually.
- Service-level generation locking remains planned for Slice 4.
