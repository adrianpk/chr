# Slice 4: Result Envelope Changed Files

Status: delivered
Slicing: Codex Runtime Worker
Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
Branch: `feat/slice-4-result-changed-files`
PR: `#52`

## Purpose

Carry changed-file summaries in the structured execution result envelope.

## Delivered Behavior

Execution result JSON now accepts an optional `changed_files` list.

Each changed file includes:

- workspace-relative path;
- optional summary.

Changed file paths are validated as workspace-relative paths. Absolute paths
and paths that escape the workspace are rejected.

`NewResultApplication` maps execution-result changed files into the existing
result application changed-file contract.

## Implementation Notes

Strict JSON decoding remains enabled. Unknown envelope fields are still
rejected.

The existing explicit `ResultApplicationOptions.ChangedFiles` path remains
available. Result envelope changed files are appended to that list.

## Contracts Added or Changed

New core type:

- `ExecutionResultChangedFile`

Updated core types:

- `ExecutionResult`
- `ResultApplication`

## Files of Interest

- `internal/chr/execution_result.go`
- `internal/chr/execution_result_test.go`
- `internal/chr/result_application.go`
- `internal/chr/result_application_test.go`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chr ./cmd/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice carries worker-reported changed files. The next slice discovers
actual workspace changes independently after execution.
