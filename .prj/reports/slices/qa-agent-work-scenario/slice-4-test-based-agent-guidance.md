# Slice 4: Test Based Agent Guidance

Status: reviewing
Slicing: QA Agent Work Scenario
Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`
Tracker: `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`
Branch: `feat/slice-4-test-based-agent-guidance`
PR: pending

## Purpose

Make the test-based development preference explicit for CHC agents and for
generated tracking/ticket work units.

## Delivered Behavior

- CHC agent conventions now prefer reproducing bug fixes with a failing test
  when feasible.
- CHC agent conventions keep table-driven tests as the default for input/output
  behavior.
- Advanced testing approaches are documented as use-case dependent rather than
  required ceremony.
- Plan tracking generation now instructs generated tickets to include
  test-based validation expectations for bug fixes.
- Prompt tests assert the tracking prompt contains the test-based guidance.
- The blog QA reference includes `AGENTS.md` and `.agents/CONVENTIONS.md` with
  equivalent testing guidance.

## Implementation Notes

The prompt guidance is added to plan tracking materialization because that is
where generated tickets are created. Specs and plans may describe behavior, but
tickets are the execution units that later agents will read directly.

## Contracts Added or Changed

- `.agents/CONVENTIONS.md` now includes a test-based bug-fix preference.
- `quizModelPlanTrackingInstructions` now biases generated ticket bodies toward
  failing-test reproduction and table-driven validation when applicable.
- `qa/ref/001/blog/AGENTS.md` delegates to blog-local conventions.

## Files of Interest

- `.agents/CONVENTIONS.md`
- `internal/chc/quiz_openai_model.go`
- `internal/chc/quiz_codex_model_test.go`
- `qa/ref/001/blog/AGENTS.md`
- `qa/ref/001/blog/.agents/CONVENTIONS.md`
- `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`

## Validation

- `go test ./internal/chc`
- `make qa-reset-blog`
- `make qa-test-blog` expected failure: unknown post returns `500`, test
  expects `404`
- `make verify`
- `git diff --check`

## Risks and Follow-ups

- The preference is advisory; execution agents are not mechanically forced to
  create or run tests.
- Slice 5 still needs final runbook and smoke-validation documentation.
