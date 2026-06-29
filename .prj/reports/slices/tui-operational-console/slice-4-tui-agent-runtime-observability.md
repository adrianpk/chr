# Slice 4: Agent And Runtime Observability

Status: delivered
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-4-tui-agent-runtime-observability`
PR: #142

## Purpose

Make agent, loop, execution, and runtime readiness state visible in the TUI
before adding any control actions.

This slice keeps the Agent and Runtime views read-only.

## Delivered Behavior

- Loads daemon state through the existing daemon state repository.
- Treats missing daemon state as a stopped daemon for display.
- Loads project execution records and shows the latest execution.
- Loads project loop run records and shows the latest loop.
- Loads runtime config and validation checks through the existing runtime
  planning and validation services.
- Renders Agent view with daemon, latest loop, latest execution, failure
  reason, and evidence paths.
- Renders Runtime view with image, Dockerfile, artifact directory, validation
  command, and readiness checks.

## Implementation Notes

`internal/chrapp` remains the TUI startup data-loading boundary. It opens the
existing user-database-backed repositories and closes them immediately after
building the read-only TUI view data.

`internal/chrtui` owns the presentation model. It selects the latest execution
and loop by `UpdatedAt` and renders project-relative runtime paths when
possible.

No daemon, runtime, loop, ticket, or execution mutation is introduced by this
slice.

## Contracts Added or Changed

- `chrtui.Options` now accepts `ObservabilityView`.
- The TUI Agent view now renders daemon, loop, and execution state.
- The TUI Runtime view now renders runtime plan and validation readiness.

## Files of Interest

- `internal/chrapp/app.go`
- `internal/chrtui/observability.go`
- `internal/chrtui/model.go`
- `internal/chrtui/model_test.go`

## Validation

```sh
go test ./internal/chrtui ./internal/chrapp
make verify
```

All commands pass.

## Risks and Follow-ups

- Latest loop and execution selection is based on `UpdatedAt`; later slices can
  add richer history navigation.
- Agent and Runtime views remain plain text until later TUI layout work adds a
  richer viewport.
