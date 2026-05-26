# Slice 2: Service Boundaries

Status: reviewing
Slicing: CLI TUI Surface Parity
Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`
Tracker: `.prj/trackers/slicing-cli-tui-surface-parity-tracker.md`
Branch: `feat/slice-2-service-boundaries`
PR: pending

## Purpose

Align audited CLI and TUI operations around shared application service
boundaries without adding new user-facing commands.

## Delivered Behavior

- Corrected the parity matrix where Slice 1 had over-classified existing shared
  paths as divergent.
- Marked ticket creation, ticket comments, ticket status changes, single-ticket
  execution, and ready-loop execution as matched where code already shared the
  service or execution helpers.
- Extracted daemon stop state mutation into `requestDaemonStopState`, shared by
  CLI daemon stop and TUI work stop.
- Added `service.CreateAuthoringSession` and routed CLI/TUI authoring session
  creation through it.

## Implementation Notes

No new CLI commands or TUI actions were added.

The daemon stop change preserves CLI output and TUI message behavior while
sharing the state mutation. The authoring change preserves CLI configurability
and TUI defaults while sharing the session repository write.

## Contracts Added or Changed

- `requestDaemonStopState` is the shared daemon stop mutation boundary.
- `service.CreateAuthoringSession` is the shared authoring session creation
  boundary.
- The parity matrix now reports no high-priority service divergences in the
  audited Slice 2 surface.

## Files of Interest

- `internal/chcapp/daemon_stop.go`
- `internal/chcapp/daemon.go`
- `internal/chcapp/service_work.go`
- `internal/chcapp/service_artifact_ticket.go`
- `internal/chcapp/author.go`
- `docs/reference/cli-tui-action-matrix.md`

## Validation

Pending final validation for this slice.

## Risks and Follow-ups

- The remaining parity gaps are mostly missing CLI commands for TUI runtime
  document operations and missing TUI surfaces for CLI admin workflows.
- Slice 3 should add CLI entrypoints for the `missing_cli` operational actions
  rather than widening this refactor slice.
