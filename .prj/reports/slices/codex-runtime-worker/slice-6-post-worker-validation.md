# Slice 6: Post-Worker Validation

Status: in review
Slicing: Codex Runtime Worker
Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
Branch: `feat/slice-6-post-worker-validation`
PR: `#54`

## Purpose

Run the configured validation command after the Codex worker command and keep
validation output separate from worker runtime logs.

## Delivered Behavior

Harness invocation now carries a dedicated validation command in addition to
the Docker build and worker run commands.

The execution runner:

- opens the runtime log artifact for build and worker output;
- opens the validation artifact for validation output;
- runs build, worker, and validation commands in order;
- classifies validation command failures with the `validation` failure kind;
- captures the result envelope only after the validation command passes.

Execution success now means the worker command completed, validation passed, and
the result envelope was captured successfully.

## Implementation Notes

The validation command is rendered as a Docker run with the same runtime mounts
and workdir as the worker command. It receives artifact environment variables
and executes the runtime plan validation command.

Validation command failures are reclassified at the runner boundary because the
low-level OS executor wraps process failures as harness failures.

## Contracts Added or Changed

Changed core type:

- `HarnessInvocation`

Changed behavior:

- harness invocation includes a `Validation` command;
- execution runner writes validation stdout and stderr to the validation
  artifact path;
- validation command failure is stored with a `validation:` failure reason.

## Files of Interest

- `internal/chc/harness_invocation.go`
- `internal/chc/harness_invocation_test.go`
- `internal/chc/execution_runner.go`
- `internal/chc/execution_runner_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chc -run 'Test(NewHarnessInvocation|ExecutionRunner)'`:
  pass
- `make verify`: pass

## Risks and Follow-ups

The worker and validation commands currently run in separate containers with
the same workspace mounts. Later orchestration should ensure this remains
compatible with any runtime image state that must persist only through mounted
workspace paths.
