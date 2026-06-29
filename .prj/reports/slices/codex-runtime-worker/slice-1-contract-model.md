# Slice 1: Worker Contract Model

Status: delivered
Slicing: Codex Runtime Worker
Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
Branch: `feat/codex-worker-slice-1-contract-model`
PR: `#49`

## Purpose

Define the side-effect free Codex worker command and context contract.

## Delivered Behavior

The core package now exposes a Codex worker command model that renders the
inside-container Codex invocation.

The command carries:

- execution ID;
- ticket key;
- ticket path;
- worker context path;
- result artifact path;
- runtime log artifact path;
- validation artifact path.

The command renders `codex exec --skip-git-repo-check --full-auto` with shell
input redirected from `CHR_WORKER_CONTEXT_PATH`.

## Implementation Notes

This slice only defines command construction. It does not execute Codex, write
context files, change Docker harness invocation, or create result artifacts.

## Contracts Added or Changed

New core types:

- `CodexWorkerCommandOptions`
- `CodexWorkerCommand`
- `CodexWorkerEnv`

New functions:

- `NewCodexWorkerCommand`
- `RenderCodexWorkerShellScript`

## Files of Interest

- `internal/chr/codex_worker.go`
- `internal/chr/codex_worker_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

The concrete command assumes the `codex` executable is available inside the
runtime container. Later slices must document container prerequisites and wire
the worker command into the Docker harness.
