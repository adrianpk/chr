# Slice 4: Runtime Plan Contract

Status: reviewing
Slicing: Agent Runtime Boundary
Plan: `.prj/plans/slicing-agent-runtime-boundary-plan.md`
Tracker: `.prj/trackers/slicing-agent-runtime-boundary-tracker.md`
Branch: `feat/slice-4-runtime-plan-contract`
PR: `#32`

## Purpose

Provide one deterministic runtime boundary planner for future ticket execution
and validation code.

## Delivered Behavior

`PlanRuntime` computes a side-effect-free runtime plan from a project directory,
runtime config, and optional run ID. The plan includes absolute project and
workspace paths, Dockerfile path, image, container workdir, project mount,
artifact directories, result path, log path, and validation command.

## Implementation Notes

The planner does not create files, inspect Docker, build images, or execute
commands. It only normalizes config and computes the boundary metadata that
later execution and validation slices can consume.

## Contracts Added or Changed

New core types and function:

- `RuntimePlanOptions`
- `RuntimePlan`
- `RuntimeMount`
- `PlanRuntime`

The Docker mount contract is a writable project-root mount targeting
`runtime.workdir`. Relative container workdirs are normalized to absolute
container paths.

Runtime result paths:

- `<artifact_dir>/result.json`
- `<artifact_dir>/runtime.log`
- `<artifact_dir>/<run_id>/result.json` when a run ID is supplied
- `<artifact_dir>/<run_id>/runtime.log` when a run ID is supplied

## Files of Interest

- `internal/chc/runtime_plan.go`
- `internal/chc/runtime_plan_test.go`
- `docs/reference/project-workspace.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

The planner currently defines the first project-root mount only. Additional
mounts, environment variables, and Docker command rendering belong to later
execution-oriented slices if they become necessary.
