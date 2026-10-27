# Slice 8: Quiz domain and model adapter

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-8-quality-quiz`
PR: pending

## Purpose

Separate clarification state and model contracts from provider transport while
preserving finite questions, answer meaning, generation identity and guarded
authoring publication.

## Delivered Behavior

The first T8.1 checkpoint moves runtime values, validation and raw-capture
curation into core/quiz. Callers use that owner directly. State spellings,
question identity, Deep/Fast limits and retained entry decoding are unchanged.
The model and provider extraction remains in progress.

## Implementation Notes

The mechanical move has 71 declaration proofs. A separate naming step removes
repeated Quiz prefixes. The parser now returns a named metadata/body result,
preserving the empty-body case and both legacy and retained entry decoding.
Shared durable size/count limits have one owner in core/workflow/state.

## Contracts Added or Changed

Scope, RuntimeState, Question, Candidate, Entry and CuratedNotes retain their
original representation. Temporary signature debt decreases from 307 to 306.
Import fixtures reject a dependency from the quiz domain back to its coordinator.
Remaining model, provider, persistence-result and operation-input work is pending.

## Files of Interest

- `internal/core/quiz/` owns runtime validation and curation.
- `internal/core/workflow/state/limits.go` retains the original shared bounds.
- [Move evidence](../../../plan/code-quality-remediation-slice-8-moves.csv)
  records the mechanical declarations before the separate naming step.
- [Signature transitions](../../../plan/code-quality-remediation-slice-8-signatures.csv)
  records direct consumer qualifications and the parser result conversion.

## Validation

- `go test ./internal/core/quiz ./internal/core ./internal/app -run '^Test(ValidateQuiz|QuizRuntime|QuizArtifact|QuizCuration|QuizModel)' -count=1 -timeout=3m`
  — passed after runtime naming and the metadata/body conversion.
- `go test ./scripts/goquality -count=1` — passed with quiz import-bound fixtures.
- `make check-go-style` — passed with 306 exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/quiz ./internal/core/workflow/state`
  — passed, zero issues.

Final clean-head make verify and focused slice validation have not run. Their
ordered evidence will be recorded through the controller after PR references.

## Risks and Follow-ups

T8.1 remains in progress; T8.2-T8.6 are pending. The next implementation unit
resolves the shared generation value dependencies before moving model payloads
and the OpenAI adapter. This checkpoint is not PR readiness.
