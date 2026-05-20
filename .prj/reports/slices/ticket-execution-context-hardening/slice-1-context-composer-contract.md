# Slice 1: Context Composer Contract

Status: reviewing
Slicing: Ticket Execution Context Hardening
Plan: `.prj/plans/slicing-ticket-execution-context-hardening-plan.md`
Tracker: `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`
Branch: `feat/slice-1-context-composer-contract`
PR: pending

## Purpose

Define the side-effect free contract that plans ticket execution context before
any source-specific collectors or Docker materialization are added.

## Delivered Behavior

The repository now has a context composition model that can:

- accept ticket identity, execution identity, runtime paths, and source
  inventory entries;
- classify sources as invariant or fresh;
- classify source kinds such as agent identity, project context, protocol
  context, documents, filesystem tree, text search, gograph, validation, and
  previous attempts;
- derive run-scoped context artifact paths from `RuntimePlan`;
- report readiness as ready or blocked when required sources are missing.

## Implementation Notes

`PlanContextComposition` is intentionally side-effect free. It normalizes and
validates the planned inventory, derives artifact paths under the runtime run
artifact directory, and returns readiness metadata without reading or writing
files.

The runtime plan remains the single source of truth for project and artifact
paths. Later slices can add collectors and renderers around this contract
without changing Docker invocation semantics.

## Contracts Added or Changed

New public internal contracts:

- `ContextComposerOptions`
- `ContextCompositionPlan`
- `ContextArtifactPaths`
- `ContextSourceInventoryEntry`
- `ContextSourceClass`
- `ContextSourceKind`
- `ContextReadiness`
- `PlanContextComposition`

The main context artifact remains `RuntimePlan.WorkerContextPath`. Additional
partial context artifacts are planned under:

```text
<run-artifact-dir>/context/
```

## Files of Interest

- `internal/chc/context_composer.go`
- `internal/chc/context_composer_test.go`
- `.prj/trackers/slicing-ticket-execution-context-hardening-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestPlanContextComposition'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice defines the contract only. It does not yet collect `AGENTS.md`,
`.chc/`, project documents, file tree output, text search, gograph output, or
Ralph attempt evidence.
