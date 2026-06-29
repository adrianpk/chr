# Slice 6: Manual Promotion and Provenance

Status: delivered
Slicing: Markdown-First Authoring Sessions
Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
Branch: `feat/slice-6-manual-promotion-provenance`
PR: #96

## Purpose

Promote selected Markdown authoring inputs into a next-step artifact while
preserving provenance.

## Delivered Behavior

The repository now supports one conservative promotion path:

```text
quiz -> draft
```

The implementation can:

- plan a draft artifact from an existing quiz;
- render an editable draft that embeds the curated quiz content;
- preserve source quiz and target draft provenance in the authoring session;
- reject unallowed target collisions;
- update the persisted authoring session after promotion;
- expose the workflow through `chr author promote quiz-to-draft`;
- support dry-run rendering without writing the draft or updating the session.

## Implementation Notes

Promotion is intentionally narrow. It proves the command and provenance shape
without trying to automate every document transition.

The generated draft remains Markdown-first and editable. Later assisted
authoring can replace the simple embedding step with model-backed curation while
keeping the same provenance model.

## Contracts Added or Changed

New internal contracts:

- `QuizToDraftPromotionOptions`
- `QuizToDraftPromotionPlan`
- `PlanQuizToDraftPromotion`
- `RenderDraftFromQuiz`
- `WriteQuizToDraftPromotion`

Repository contract update:

- `AuthoringSessionRepository.UpdateSession`

New user-facing command:

```text
chr author promote quiz-to-draft
```

## Files of Interest

- `internal/chr/authoring_promotion.go`
- `internal/chr/authoring_promotion_test.go`
- `internal/chr/authoring_session_repository.go`
- `internal/chrapp/author.go`
- `internal/chrapp/cli_test.go`
- `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chr -run 'TestPlanQuizToDraftPromotion|TestWriteQuizToDraftPromotion'`: pass
- `go test ./internal/chrapp -run 'TestAuthor'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice only implements quiz-to-draft promotion. Draft-to-spec and other
document transitions remain future work unless selected in a later slicing.
