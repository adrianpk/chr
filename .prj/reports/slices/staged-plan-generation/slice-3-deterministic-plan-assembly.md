# Slice 3: Deterministic Plan Assembly

Status: reviewing
Slicing: Staged Plan Generation
Plan: `.prj/plans/slicing-staged-plan-generation-plan.md`
Tracker: `.prj/trackers/slicing-staged-plan-generation-tracker.md`
Branch: `feat/slice-3-deterministic-plan-assembly`
PR: pending

## Purpose

Connect staged plan generation into `GeneratePlan` and assemble the final plan
Markdown deterministically from a validated outline plus validated slice
details.

## Delivered Behavior

- `GeneratePlan` now generates an outline first, then one detail response per
  outline slice.
- The final plan is written once, after all staged responses validate.
- Slice detail failures prevent final plan writes.
- The generated plan body is assembled by service-owned deterministic section
  ordering.
- Staged request and response IDs are recorded in front matter and in a model
  call provenance section.

## Implementation Notes

The final generated plan artifact still uses the existing `GeneratedPlan`
writer and keeps status `planned`. Existing archive-before-replace and
same-scope generation guards continue to wrap the generation path.

The assembler lives in `internal/chc` and validates that every outline slice has
exactly one accepted detail response before returning Markdown. The service
sets source request and response IDs after successful staged generation, then
writes the canonical plan.

## Contracts Added or Changed

- `AssembleGeneratedPlanBody`
- `GeneratePlan` staged orchestration: outline, slice details, deterministic
  assembly, final write.
- Generated plan provenance now records staged request and response IDs.

## Files of Interest

- `internal/chc/plan_assembly.go`
- `internal/chcapp/service_plan.go`
- `internal/chcapp/service_plan_test.go`
- `internal/chcapp/service_quiz_test.go`

## Validation

- `gofmt -w internal/chc/plan_assembly.go internal/chcapp/service_plan.go internal/chcapp/service_plan_test.go internal/chcapp/service_quiz_test.go` - pass
- `go test ./internal/chc ./internal/chcapp` - pass
- `go test ./...` - pass
- `git diff --check` - pass
- `make verify` - fail: `lint-strict` reports pre-existing lint debt in existing test, trace, service, and TUI files.

## Risks and Follow-ups

- The complete repository gate remains blocked by existing lint-strict debt.
- The generated Markdown is deterministic but intentionally conservative; later
  tuning can adjust section wording if product review asks for it.
