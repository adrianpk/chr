# Slice 11: Live Work Dash

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-11-work-dash`
PR: pending

## Purpose

Deliver the approved [plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md) Slice 11. Slice 10
is verified integrated through PR #407 at
`ce1a19c0c10a50c27807b976fa353c81a0f6912e`.

## Delivered Behavior

The shared observation reads native or supported legacy queue readiness,
separate invocation/attempt/ticket contexts, recent activity and retained
results. Host validation remains distinct from confirmed ticket integration.
Dash now renders the queue at left, independent current contexts at upper right,
and activity/results below, stacking at narrow widths. Compact readiness,
exclusions and actual counters retain their source meanings. Retained controls
use their existing operations and confirmations with a separate result pane.
Bounded refresh and stale recovery remain in T11.3.

## Implementation Notes

T11.1 and T11.2 are complete. Native reads use the existing inspection lock and shared
selection/prerequisite rules without an unrelated legacy Board dependency.
Legacy reads use authoritative files and an existing SQLite database in read-only
mode. Event matching rejects other projects, executions and tickets. Missing
phase events do not erase known results or make quiet validation terminal.
T11.3 follows in the recorded order.

## Contracts Added or Changed

`ObserveWork` is the central adapter-neutral read. `InspectWorkflowWork` reads
coherent native records without recovery. Legacy reads never initialize a
database. Queue pages have 128 items and continuation; results and event tails
have explicit source/output bounds documented in the owning reference.

## Files of Interest

- `internal/app/service_work_observation*.go`: shared authority-aware projection.
- `internal/core/workflow_work_observation.go`: coherent native observation.
- `internal/core/legacy_work_observation.go`: read-only legacy records and queue.
- `internal/core/agent_event_recent.go`: bounded structured event tails.
- [Work observation reference](../../../../../docs/reference/work-observation.md): sources, identity, bounds and state.

## Validation

T11.1 checks passed:

- `go test -race ./internal/core ./internal/app -run 'WorkflowWorkObservation|WorkflowWorkPending|LegacyWorkObservation|RecentWorkEvents|WorkObservation' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/app`
- `git diff --check`

Real fixtures cover two persisted native claims in separate contexts, concurrent
native inspection, blocked journal recovery, corrupt unrelated legacy data,
absent/read-only SQLite state, retained completed results without events,
quiet validation, context filtering and stable event deduplication. Tests close
owned claims and database handles; no worker or live runner is launched.

T11.2 checks also passed:

- `go test -race ./internal/core ./internal/app ./internal/webui -run 'WorkflowWorkObservation|WorkObservation|WebDash|DashControls|HandlerProjectBoardDashRoutes|HandlerBoardAndDashFragments|HandlerRendersGenerationAndWorkActionForms' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/app`
- `git diff --check`

Real HTTP tests observe native claims and retained legacy results. The semantic
view revision excludes observation time. Native pending queues exclude active or
validated records and avoid counting slice-owned tasks again in the loose
selection. Active process observations expose only the executable name, preserving
its own attempt binding and omitting arguments and process identities.

## Risks and Follow-ups

None recorded.
