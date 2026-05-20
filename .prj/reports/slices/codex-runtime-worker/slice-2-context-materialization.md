# Slice 2: Worker Context Materialization

Status: delivered
Slicing: Codex Runtime Worker
Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
Branch: `feat/slice-2-worker-context-materialization`
PR: `#50`

## Purpose

Write the ticket execution context consumed by Codex inside the runtime
container.

## Delivered Behavior

The runtime plan now reserves `context.md` in the execution run artifact
directory.

The core package also exposes a worker context materializer that writes a
deterministic Markdown artifact containing:

- execution ID;
- ticket key;
- ticket path;
- result artifact path;
- runtime log artifact path;
- validation artifact path;
- structured JSON result contract.

## Implementation Notes

The materializer creates parent directories before writing the context. It does
not execute Codex, run Docker, run validation, or write result artifacts.

## Contracts Added or Changed

New core types:

- `CodexWorkerContextMaterializer`
- `MaterializeCodexWorkerContextOptions`

New functions:

- `NewCodexWorkerContextMaterializer`
- `RenderCodexWorkerContext`

Updated core type:

- `RuntimePlan`

## Files of Interest

- `internal/chc/runtime_plan.go`
- `internal/chc/codex_worker_context.go`
- `internal/chc/codex_worker_context_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

The context currently includes the minimal result contract. Later slices will
wire this artifact into the Docker harness and may add richer ticket context if
the worker needs it.
