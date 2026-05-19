# Slice 4: Harness Invocation Contract

Status: delivered
Slicing: Single Ticket Execution
Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
Branch: `feat/slice-4-harness-invocation-contract`
PR: `#37`

## Purpose

Define the side-effect free contract for turning a runtime plan into Docker
build and run command arguments for one ticket execution.

## Delivered Behavior

The core package now exposes a harness invocation model that renders Docker
build and run commands from a `RuntimePlan`, execution ID, and ticket key.

Runtime artifact paths can now be scoped under both ticket key and execution ID:

`<artifact_dir>/<ticket_key>/<execution_id>/`

`chc ticket start` passes the selected ticket key into runtime validation so the
pending execution record stores ticket-scoped artifact paths.

## Implementation Notes

Harness invocation construction validates required identity fields and returns
command data only. It does not execute Docker, create directories, or write
artifacts.

The run command carries execution context through environment variables:

- `CHC_EXECUTION_ID`
- `CHC_TICKET_KEY`
- `CHC_RESULT_PATH`
- `CHC_LOG_PATH`

## Contracts Added or Changed

New core types:

- `HarnessInvocationOptions`
- `HarnessInvocation`
- `HarnessCommand`

New function:

- `NewHarnessInvocation`

Runtime plan options now accept `TicketKey`. When `TicketKey` and `RunID` are
both set, run artifacts are placed below `<artifact_dir>/<ticket_key>/<run_id>/`.

## Files of Interest

- `internal/chc/harness_invocation.go`
- `internal/chc/harness_invocation_test.go`
- `internal/chc/runtime_plan.go`
- `internal/chc/runtime_plan_test.go`
- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chc ./cmd/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice renders command arguments only. Slice 5 is planned to invoke the
harness and capture process output.
