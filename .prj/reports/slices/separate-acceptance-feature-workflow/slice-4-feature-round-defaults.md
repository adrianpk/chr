# Slice 4: Feature Round Defaults

Status: reviewing
Slicing: Separate Acceptance Feature Workflow
Plan: `ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md`
Tracker: `ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md`
Branch: `feat/slice-4-feature-round-defaults`
PR: `#297`

## Purpose

Add a dedicated feature clarification round so acceptance feature generation can
ask for missing scenario details after the spec exists, instead of biasing the
general quiz toward Gherkin concepts.

## Delivered Behavior

Acceptance feature generation can now return default-oriented clarification
prompts when it needs more information. The service writes those prompts to an
editable feature-round note under `.chc/user/acceptance/<case>-feature-round.md`.
Future feature generation requests include that note as acceptance context.

## Implementation Notes

`needs_more_info` feature responses now require structured clarification
prompts with a question, inferred default answer, and reason. Generated,
rejected, and not-applicable responses do not retain prompts.

The feature-round note is intentionally separate from quiz notes and generated
`.feature` files. It captures the narrower acceptance formalization step after
dump, quiz, and spec material already exist.

## Contracts Added or Changed

- Added `clarification_prompts` to the acceptance feature model payload.
- Added `.chc/user/acceptance/<case>-feature-round.md` as the feature-round
  note path.
- Added service behavior that writes feature-round notes on `needs_more_info`.
- Added service behavior that includes existing feature-round notes in later
  acceptance feature generation context.

## Files of Interest

- `internal/chc/acceptance_feature.go`
- `internal/chc/quiz_model_payload.go`
- `internal/chc/quiz_openai_model.go`
- `internal/chcapp/service_acceptance_feature.go`
- `internal/chcapp/service_acceptance_feature_test.go`
- `internal/chc/acceptance_feature_test.go`

## Validation

- `go test ./internal/chc -run 'AcceptanceFeature|QuizModelFeature'` (pass)
- `go test ./internal/chcapp -run 'ServiceGenerateAcceptanceFeature'` (pass)
- `go test ./internal/chc ./internal/chcapp` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- TUI rendering and action wiring are pending in Slice 5.
- `godog` execution remains deferred.
