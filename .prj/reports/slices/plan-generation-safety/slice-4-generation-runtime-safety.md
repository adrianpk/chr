# Slice 4: Generation Runtime Safety

Status: delivered
Slicing: Plan Generation Safety
Plan: `.prj/plans/slicing-plan-generation-safety-plan.md`
Tracker: `.prj/trackers/slicing-plan-generation-safety-tracker.md`
Branch: `feat/slice-4-generation-runtime-safety`
PR: `#201`

## Purpose

Harden expensive spec and plan generation so long-running model calls have an
appropriate timeout, duplicate same-scope requests are blocked, and failure
messages make clear that saved content remains visible.

## Delivered Behavior

- Spec and plan generation now run under a dedicated long-generation timeout.
- The OpenAI adapter uses a longer HTTP timeout for generated specs and plans
  than for next-question quiz generation.
- Service-level guards prevent concurrent spec or plan generation for the same
  project/scope.
- Guard blocks occur before a second same-scope model invocation is started.
- Plan regeneration uses the same plan-scope guard as plan generation.
- TUI spec and plan generation failures state that the last saved artifact is
  still visible when available.
- Spec generation now emits diagnostic lifecycle events matching the plan
  generation path.
- Guard, archive, model, write, success, and failure events are covered by the
  diagnostic logging path.

## Implementation Notes

- The runtime guard is keyed by project directory, generation kind, and scope.
- `GeneratePlan`, `RegeneratePlan`, and `GenerateSpec` use the shared runtime
  guard and timeout wrapper.
- `GeneratePlan` and `RegeneratePlan` share the same `plan` guard key so a
  regenerate cannot overlap with a create for the same scope.
- The next-question quiz path still uses the existing shorter/default model
  execution behavior.

## Contracts Added or Changed

- Added long-generation timeout handling for spec and plan generation.
- Added service-level generation guard behavior.
- Added OpenAI timeout selection by quiz model action.
- Updated TUI generation failure status copy for specs and plans.

## Files of Interest

- `internal/chr/quiz_openai_model.go`
- `internal/chr/quiz_model_test.go`
- `internal/chrapp/generation_runtime.go`
- `internal/chrapp/service_plan.go`
- `internal/chrapp/service_spec.go`
- `internal/chrapp/service_plan_test.go`
- `internal/chrapp/service_spec_test.go`
- `internal/chrtui/plan_actions.go`
- `internal/chrtui/spec_actions.go`
- `internal/chrtui/model_test.go`

## Validation

- `go test ./internal/chr`: pass
- `go test ./internal/chrapp`: pass
- `go test ./internal/chrtui`: pass
- `go test ./...`: pass
- `make verify`: fail

`make verify` fails in `lint-strict` on existing lint debt across the plan,
spec, and TUI areas. The current failure set includes complexity, function
length, unused functions, errcheck, formatting, and whitespace rules.

## Commits

- `1189b57 docs(plan): select slice 4 generation runtime safety`
- `6a58219 feat(generation): guard long-running spec and plan requests`
- `27dc933 fix(tui): clarify generation failure state`
- `505f092 test(generation): cover timeout and guard behavior`
- `07ca43b feat(generation): log spec generation lifecycle`
- `ba67767 fix(openai): extend spec and plan generation timeout`
- `6634da0 test(tui): fix generation failure assertion flow`

## Risks and Follow-ups

- The guard is process-local. A future background worker or daemonized job
  model may need a durable lock if multiple processes can generate the same
  scope concurrently.
- Strict lint remains blocked by pre-existing codebase debt.
