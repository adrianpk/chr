# Slice 3: Artifact Backend Defaults

Status: reviewing
Slicing: Codex Artifact Generation
Plan: `.prj/plans/slicing-codex-artifact-generation-plan.md`
Tracker: `.prj/trackers/slicing-codex-artifact-generation-tracker.md`
Branch: `feat/slice-3-artifact-backend-defaults`
PR: pending

## Purpose

Make Codex the default backend for long artifact generation while preserving an
explicit OpenAI override for direct-model paths.

## Delivered Behavior

- Added Codex-backed dump curation through the shared agent runner.
- Made dump, spec, plan, and plan tracking generation use Codex by default.
- Preserved per-artifact `openai` overrides for direct OpenAI-backed behavior.
- Kept quiz question generation outside the Codex default path.
- Added routing tests for default Codex behavior and OpenAI overrides.
- Documented artifact generation backend settings.

## Implementation Notes

The reconciled dump implementation uses the shared `AgentRunner` and records
deferred Codex response trace events after payload parsing. The service keeps
file ownership: Codex returns structured curation JSON, and Go still writes dump
files.

Backend selection now treats an empty backend environment variable as Codex.
Setting the relevant backend variable to `openai` selects the previous
direct-model behavior for that artifact type.

## Contracts Added or Changed

- `CHC_DUMP_GENERATION_BACKEND=openai` selects direct OpenAI dump curation.
- `CHC_SPEC_GENERATION_BACKEND=openai` selects direct OpenAI spec generation.
- `CHC_PLAN_GENERATION_BACKEND=openai` selects staged direct-model plan
  generation.
- `CHC_PLAN_TRACKING_GENERATION_BACKEND=openai` selects direct OpenAI plan
  tracking synthesis.
- Empty backend values use Codex for all four artifact kinds.

## Files of Interest

- `internal/chc/agent_run.go`
- `internal/chcapp/service.go`
- `internal/chcapp/service_dump_curation_codex.go`
- `internal/chcapp/service_dump_curation_test.go`
- `internal/chcapp/service_plan_test.go`
- `internal/chcapp/service_test.go`
- `docs/reference/artifact-generation-backends.md`
- `docs/index.md`

## Validation

- `gofmt -w internal/chc/agent_run.go internal/chcapp/service.go internal/chcapp/service_dump_curation_codex.go internal/chcapp/service_dump_curation_test.go internal/chcapp/service_plan_test.go internal/chcapp/service_test.go`
- `go test ./internal/chc ./internal/chcapp`
- `make verify`

## Risks and Follow-ups

- Slice 4 should run a clean bundle and trace audit against a reference
  workspace to verify the defaults end to end.
