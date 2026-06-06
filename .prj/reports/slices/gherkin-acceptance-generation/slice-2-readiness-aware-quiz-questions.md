# Slice 2: Readiness-Aware Quiz Questions

Status: reviewing
Slicing: Gherkin Acceptance Generation
Plan: `ops/plans/2026-06-06-gherkin-acceptance-generation-plan.md`
Tracker: `ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md`
Branch: `feat/slice-2-readiness-aware-quiz-questions`
PR: pending

## Purpose

Connect the acceptance readiness contract to quiz generation so early questions
can gather Gherkin-like behavior inputs when the source context supports it,
while keeping normal dump-oriented discovery for unsuitable cases.

## Delivered Behavior

Quiz model requests now include an optional `acceptance_readiness` payload and
context section derived from project, quiz, and supplemental context.

The quiz prompt now treats readiness as guidance. `needs_more_info` biases the
next questions toward missing actor, trigger, initial state, observable outcome,
success path, failure path, or boundary details without exposing Gherkin terms
to the user. `not_applicable` keeps the quiz on normal dump/spec discovery.

Curated quiz notes now include an `Acceptance Readiness` section when prompt
intent provenance contains readiness decisions, missing slots, or inferred
slots.

## Implementation Notes

Readiness inference is intentionally conservative and deterministic in this
slice. It inspects existing context for acceptance-test signals and records
inferred elements as defaults, not confirmed requirements.

The provider payload still keeps full supplemental context out of next-question
generation. Only the compact readiness payload is sent with the focused quiz
request.

## Contracts Added or Changed

- Added optional `acceptance_readiness` to `QuizModelPayload`.
- Added an `acceptance_readiness` quiz context section when readiness is present.
- Added readiness guidance to deep and fast quiz question instructions.
- Added `acceptance_readiness` and `acceptance_prompt_intent` to quiz routing
  state when applicable.
- Added `Acceptance Readiness` to curated quiz notes.

## Files of Interest

- `internal/chc/quiz_model_payload.go`
- `internal/chc/quiz_openai_model.go`
- `internal/chc/quiz_curation.go`
- `internal/chc/quiz_artifact_repository.go`
- `internal/chcapp/service_quiz.go`

## Validation

- `go test ./internal/chc -run 'QuizModelProviderPayloadCarriesAcceptanceReadiness|QuizModelQuestionInstructions|QuizModelRoutingStateCarriesAcceptanceReadiness|OpenAIQuizModelRequestBodySerializesFocusedNextQuestionPayload'` (pass)
- `go test ./internal/chc -run 'QuizArtifactRepositoryWritesCuratedNotes|CurateQuizEntriesPreservesAcceptanceReadiness|QuizArtifactRepositoryLoadsRawCaptureEntries'` (pass)
- `go test ./internal/chcapp -run 'NewQuizModelRequest.*Acceptance'` (pass)
- `go test ./internal/chc` (pass)
- `go test ./internal/chcapp` (pass)

## Risks and Follow-ups

- Readiness inference is heuristic in this slice; later generated feature files
  must still pass deterministic static checks before being accepted.
- This slice does not generate `.feature` files and does not execute `godog`.
