# Slice 6: Execution Result Capture

Status: delivered
Slicing: Single Ticket Execution
Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
Branch: `feat/slice-6-execution-result-capture`
PR: `#39`

## Purpose

Capture the structured result artifact produced by a harness run and use it to
finalize execution status.

## Delivered Behavior

The core package now defines an execution result envelope:

- `execution_id`
- `ticket_key`
- `status`
- `summary`

After harness commands complete, the execution runner reads the result artifact
from `record.Artifacts.ResultPath`. A valid `succeeded` result marks the
execution `succeeded`.

The runner marks the execution `failed` and records the diagnostic when:

- the result artifact is missing;
- the result JSON is malformed;
- the result status is invalid;
- the result execution ID does not match the execution record;
- the result ticket key does not match the execution record;
- the result status is `failed`.

## Implementation Notes

The result loader rejects unknown JSON fields and validates the required
identity/status fields. The runner keeps process output capture in
`runtime.log`; structured result interpretation is isolated to the result
artifact.

## Contracts Added or Changed

New core types:

- `ExecutionResult`
- `ExecutionResultStatus`

New functions:

- `LoadExecutionResult`
- `ValidateExecutionResult`

Harnesses must write a result envelope to `CHC_RESULT_PATH`.

## Files of Interest

- `internal/chc/execution_result.go`
- `internal/chc/execution_result_test.go`
- `internal/chc/execution_runner.go`
- `internal/chc/execution_runner_test.go`
- `cmd/chc/main_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chc ./cmd/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice validates and records result capture failures as generic failed
executions. Slice 7 is planned to classify failure types more explicitly.
