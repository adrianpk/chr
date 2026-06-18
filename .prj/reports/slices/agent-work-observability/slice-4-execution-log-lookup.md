# Slice 4: Execution Log Lookup

Status: reviewing
Slicing: Agent Work Observability
Plan: `ops/plans/2026-06-18-agent-work-observability-plan.md`
Tracker: `ops/trackers/2026-06-18-agent-work-observability-tracker.md`
Branch: `feat/slice-4-execution-log-lookup`
PR: pending

## Purpose

Expose a stable way to locate detailed runtime evidence for one execution
without embedding a raw log viewer in the TUI.

## Delivered Behavior

The CLI now supports:

```text
chc work logs <execution-id>
```

The command resolves the current project, loads the execution record, and prints
the execution status plus runtime artifact paths. Output includes relative paths
and existence flags for the run directory, runtime log, validation log, result
file, execution report, and event stream.

The running work overlay now renders a direct lookup hint when recent events
include an execution id:

```text
Lookup: chc work logs <execution-id>
```

This keeps the normal Dash event stream concise while giving users a stable
debug path for deeper logs.

## Implementation Notes

`work logs` reads existing `ExecutionRecord` data and does not infer container
ids. Missing log files do not fail the lookup; they are reported with
`*_exists: false`.

The command rejects unknown executions and executions that do not belong to the
resolved project.

## Contracts Added or Changed

- `chc work logs <execution-id>`
- Work-running TUI details can include `Lookup: chc work logs <execution-id>`.

## Files of Interest

- `internal/app/work.go`
- `internal/app/cli_test.go`
- `internal/tui/work_actions.go`
- `internal/tui/model_test.go`

## Validation

- `go test ./internal/app` (pass)
- `go test ./internal/tui` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The command locates logs but does not tail or render log contents.
- Event readback and recovery summaries remain a later slice.
