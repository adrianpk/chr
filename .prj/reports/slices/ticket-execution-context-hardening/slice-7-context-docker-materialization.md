# Slice 7: Docker-Compatible Context Materialization

Status: reviewing
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-7-context-docker-materialization`
PR: #84

## Purpose

Write the composed execution context package into the runtime artifact layout
that the existing Docker harness already mounts into the worker container.

## Delivered Behavior

The repository now has a context package materializer that can:

- write the main `context.md` at `RuntimePlan.WorkerContextPath`;
- write sibling context artifacts under `run_artifact_dir/context/`;
- write `agent.md`, `project.md`, `protocols.md`, `documents.md`,
  `code-context.md`, `previous-attempt.md`, and `source-inventory.md`;
- render Docker-visible `/workspace/...` paths in the main context;
- keep `CHC_WORKER_CONTEXT_PATH` as the main Codex input;
- render included/skipped source visibility for dry-run and review.

## Implementation Notes

This slice adds materialization as a standalone package writer. It does not yet
wire all context collectors into ticket execution. That integration remains
controlled by the final readiness and handoff slice.

The main context references both container-visible paths and host paths so the
same artifact can be used by Codex inside Docker and by humans reviewing local
execution artifacts.

## Contracts Added or Changed

New internal contracts:

- `ContextPackageMaterializer`
- `MaterializeContextPackageOptions`
- `NewContextPackageMaterializer`
- `RenderMainContextPackage`
- `RenderContextPackageDryRun`

## Files of Interest

- `internal/chc/context_materializer.go`
- `internal/chc/context_materializer_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestContextPackageMaterializer|TestRenderMainContextPackage|TestRenderContextPackageDryRun'`: pass
- `make verify`: pass

## Risks and Follow-ups

The materializer writes supplied context strings and does not decide readiness
or collector ordering. Slice 8 is expected to add the readiness gate and handoff
evidence links before agent execution.
