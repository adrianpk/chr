# Slice 5: Agent Harness Invocation

Status: delivered
Slicing: Single Ticket Execution
Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
Branch: `feat/slice-5-agent-harness-invocation`
PR: pending

## Purpose

Run one bounded harness attempt for a prepared ticket execution and persist the
execution lifecycle state.

## Delivered Behavior

`chc ticket start <ticket-key>` now invokes the rendered harness in non-dry-run
mode after preflight creates the pending execution record.

The execution runner:

- creates the execution run artifact directory;
- marks the execution `running` before process invocation;
- runs the rendered Docker build command;
- runs the rendered Docker run command;
- captures command stdout and stderr into the runtime log artifact;
- marks the execution `succeeded` when both commands pass;
- marks the execution `failed` and records a failure reason when either command
  fails.

`--dry-run` remains preflight-only and does not persist, create artifacts, or
invoke the harness.

## Implementation Notes

The runner accepts a `HarnessExecutor` interface so tests can use a fake executor
without invoking Docker. The production executor delegates to `os/exec`.

Runner timestamps are injectable for deterministic tests.

## Contracts Added or Changed

New core types:

- `HarnessExecutor`
- `OSHarnessExecutor`
- `ExecutionRunner`
- `RunExecutionOptions`

New function:

- `NewExecutionRunner`

`ticket start` now prints the final execution status after a real harness
attempt.

## Files of Interest

- `internal/chc/execution_runner.go`
- `internal/chc/execution_runner_test.go`
- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chc ./cmd/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

The runner captures process output but does not parse structured result data.
Result envelope handling is planned for Slice 6.
