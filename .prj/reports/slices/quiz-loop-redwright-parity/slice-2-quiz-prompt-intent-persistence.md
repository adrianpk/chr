# Slice 2: Quiz Prompt Intent Persistence

Status: delivered
Slicing: Quiz Loop RedWright Parity
Plan: `.prj/plans/slicing-quiz-loop-redwright-parity-plan.md`
Tracker: `.prj/trackers/slicing-quiz-loop-redwright-parity-tracker.md`
Branch: `feat/slice-2-quiz-prompt-intent-persistence`
PR: #194

## Purpose

Persist and rehydrate the quiz model prompt intent returned with each active
question.

The previous flow preserved display and routing fields such as topic and depth,
but reconstructed `current_question.prompt_intent` from those fields. That lost
semantic metadata returned by the model, including follow-up depth and routing
state.

## Delivered Behavior

- `QuizEntry` now stores the formal prompt intent associated with the answered
  or skipped question.
- Raw quiz capture writes a `Prompt intent` section for each turn when prompt
  intent is available.
- Raw capture parsing reloads that prompt intent into `QuizEntry`.
- Service model request construction reuses persisted prompt intent from active
  questions and raw entries.
- Existing fallback reconstruction remains for old raw captures and non-model
  questions.

## Implementation Notes

The model response already stores formal prompt intent JSON in
`QuizQuestion.Rationale`. This slice carries that JSON into `QuizEntry` during
answer or skip recording, then parses it back into
`QuizModelCurrentQuestionPayload.PromptIntent` for the next model request.

The fallback path remains intentionally narrow and compatibility-oriented. It is
used only when the stored prompt intent is empty or invalid.

## Contracts Added or Changed

- `QuizEntry.PromptIntent` is the persisted source for previous-question model
  intent.
- Raw quiz capture may include:

```text
Prompt intent:

<compact JSON object>
```

- `current_question.prompt_intent` must prefer stored model intent over derived
  depth/status metadata.

## Files of Interest

- `internal/chr/quiz_runtime.go`
  - adds prompt intent to quiz entries
- `internal/chr/quiz_artifact_repository.go`
  - writes prompt intent into raw capture
- `internal/chr/quiz_curation.go`
  - parses prompt intent from raw capture
- `internal/chrapp/service_quiz.go`
  - rehydrates prompt intent into next model requests
- `internal/chrapp/service_quiz_test.go`
  - covers active-question and raw-entry rehydration

## Validation

- `go test ./internal/chr ./internal/chrapp`
- `go test ./internal/chrapp`
- `git diff --check`
- `make verify`

## Risks and Follow-ups

- Slice 3 still needs to use the persisted prompt intent for routing decisions.
- Old raw captures without prompt intent continue through compatibility
  fallback metadata.
