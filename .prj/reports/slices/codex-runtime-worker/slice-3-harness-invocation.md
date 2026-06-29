# Slice 3: Harness Worker Invocation

Status: delivered
Slicing: Codex Runtime Worker
Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
Branch: `feat/slice-3-harness-worker-invocation`
PR: `#51`

## Purpose

Run the Codex worker command through the existing Docker harness invocation.

## Delivered Behavior

Harness invocation now renders `docker run` so the runtime container executes
the Codex worker shell command instead of directly running the validation
command.

The run command passes:

- execution ID;
- ticket key;
- ticket path;
- worker context path;
- result artifact path;
- runtime log artifact path;
- validation artifact path.

## Implementation Notes

Runtime planning now carries `ValidationPath`, and execution records copy it
from the runtime plan. The Docker command construction remains side-effect
free.

Post-worker validation is still a later slice. This slice only routes the
container invocation to the Codex worker command.

## Contracts Added or Changed

Updated core types:

- `HarnessInvocationOptions`
- `RuntimePlan`
- `ExecutionArtifacts`

Updated behavior:

- `NewHarnessInvocation` now requires `TicketPath`.
- Docker run command renders the Codex worker shell script.

## Files of Interest

- `internal/chr/harness_invocation.go`
- `internal/chr/harness_invocation_test.go`
- `internal/chr/runtime_plan.go`
- `internal/chr/execution.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chr ./cmd/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

The invocation now expects the runtime image to provide the `codex` executable.
Later slices must run post-worker validation and improve worker-specific failure
classification.
