# Slice 7: Execution Failure Handling

Status: delivered
Slicing: Single Ticket Execution
Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
Branch: `feat/slice-7-execution-failure-handling`
PR: `#40`

## Purpose

Classify execution failures consistently and print concise diagnostic output
from `chr ticket start`.

## Delivered Behavior

Execution failures are classified with stable failure kinds:

- `preflight`
- `runtime`
- `harness`
- `validation`
- `result_capture`

The execution runner records classified failure reasons for harness,
validation, and result-capture failures.

`chr ticket start` now prints failure diagnostics when available:

- failure kind;
- failure reason;
- log, result, and validation artifact paths.

Runtime validation failures are rendered as blocked pre-execution records in the
CLI output, without creating a durable execution record.

## Implementation Notes

Failure classification is stored in `failure_reason` using a stable prefix such
as `harness:` or `result_capture:`. This avoids a schema migration while keeping
the classification queryable and visible in CLI output.

Diagnostic paths are planned artifact paths. They may point to files that do not
exist when the failure happens before an artifact is written.

## Contracts Added or Changed

New core types:

- `ExecutionFailureKind`
- `ExecutionFailure`

New functions:

- `NewExecutionFailure`
- `ClassifyExecutionFailure`
- `ExecutionFailureKindFromReason`

`ticket start` failure output adds:

- `failure_kind`
- `failure_reason`
- `diagnostic_log_path`
- `diagnostic_result_path`
- `diagnostic_validation_path`

## Files of Interest

- `internal/chr/execution_failure.go`
- `internal/chr/execution_failure_test.go`
- `internal/chr/execution_runner.go`
- `internal/chr/execution_runner_test.go`
- `cmd/chr/main.go`
- `cmd/chr/main_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chr ./cmd/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

Failure kind is encoded in `failure_reason` rather than a dedicated database
column. A later reporting/query slice can add a column if direct filtering by
failure kind becomes necessary.
