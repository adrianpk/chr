# Slice 1: Quiz Provider Parity Fixtures

Status: delivered
Slicing: Quiz Loop RedWright Parity
Plan: `.prj/plans/slicing-quiz-loop-redwright-parity-plan.md`
Tracker: `.prj/trackers/slicing-quiz-loop-redwright-parity-tracker.md`
Branch: `feat/slice-1-quiz-provider-parity-fixtures`
PR: #193

## Purpose

Lock the quiz model provider boundary with deterministic fixtures before
changing runtime routing or persistence.

The slice proves the current one-question-at-a-time contract in tests so later
slices can change prompt-intent persistence and branch routing without guessing
which payload shape reaches the provider.

## Delivered Behavior

- Next-question OpenAI request bodies serialize focused quiz payloads without
  broad `context_sections`.
- Draft-spec OpenAI request bodies preserve `context_sections` for synthesis.
- Deep provider responses decode one active question for normal, skipped-entry,
  and completion-recommended cases.
- Completion recommendation maps to a non-required active question while keeping
  the recommendation in prompt intent metadata.

## Implementation Notes

The slice adds tests around `OpenAIQuizModelAdapter.openAIRequestBody` instead
of only checking the internal provider payload map. This verifies the actual
request envelope, JSON input string, response format name, storage flag, and
metadata action sent to the provider.

The response fixtures keep output decoding grounded in the same JSON contract
used by the OpenAI structured-output format.

## Contracts Added or Changed

- `quiz_next_question` provider requests include:
  - `project`
  - `quiz`
  - `recent_entries`
  - `current_question`
  - `output_contract`
- `quiz_next_question` provider requests do not include broad
  `context_sections`.
- `quiz_draft_spec` provider requests include broad `context_sections`.
- Completion-recommended question responses produce optional active questions.

## Files of Interest

- `internal/chr/quiz_model_test.go`
  - adds provider payload and response parity fixtures

## Validation

- `go test ./internal/chr`
- `git diff --check`
- `make verify`

## Risks and Follow-ups

- This slice does not change routing behavior; prompt-intent persistence is
  handled by Slice 2.
- This slice does not prove manual TUI question quality; lab validation is
  handled by Slice 4.
