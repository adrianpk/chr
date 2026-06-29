# Slice 7: Worker Failure Classification and Docs

Status: in review
Slicing: Codex Runtime Worker
Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
Branch: `feat/slice-7-worker-failure-docs`
PR: `#55`

## Purpose

Make Codex worker failures distinguishable from Docker build failures, result
capture failures, and post-worker validation failures.

## Delivered Behavior

Execution failure classification now includes a dedicated `worker` failure kind.

The execution runner classifies failures by command phase:

- Docker build command failures remain `harness`;
- Codex worker command failures are `worker`;
- post-worker validation command failures remain `validation`;
- result envelope read and validation failures remain `result_capture` or
  `validation` according to the existing result contract.

`ticket start` now prints `failure_kind: worker` for worker command failures and
continues to print diagnostic artifact paths for runtime log, result, and
validation artifacts.

## Implementation Notes

The low-level OS executor still wraps process failures as harness failures
because it only receives a command. The runner reclassifies errors at the phase
boundary where build, worker, and validation commands are known.

## Contracts Added or Changed

New failure kind:

- `worker`

Changed behavior:

- worker command failures produce `worker:` failure reasons;
- `ExecutionFailureKindFromReason` recognizes `worker`;
- ticket start output surfaces worker failures through `failure_kind`.

## Files of Interest

- `internal/chr/execution_failure.go`
- `internal/chr/execution_runner.go`
- `internal/chr/execution_runner_test.go`
- `cmd/chr/main_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chr ./cmd/chr -run 'TestExecution(Failure|Runner)|TestTicketStartPrints(Worker|Harness)Failure'`:
  pass
- `make verify`: pass

## Risks and Follow-ups

Worker diagnostics are currently surfaced through existing artifact paths.
Later orchestration can add richer worker-specific excerpts once execution log
parsing exists.
