# Slice 8: Quiz Curation Polish

Status: delivered
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-8-quiz-curation-polish`
PR: #190

## Purpose

Finish the interactive quiz loop by turning completed quiz sessions into
curated notes and draft spec artifacts while keeping quiz behavior owned by the
application service.

## Delivered Behavior

- Completing a quiz refreshes curated quiz notes from the raw capture.
- Curated notes group answered turns by topic instead of mirroring raw
  questions one-to-one.
- Completed quiz plus dump context generates a spec artifact with
  `status: draft`.
- Draft spec synthesis uses the formal quiz model request/response boundary.
- The Quizzes detail panel renders curated notes when no active question is
  running.
- The interactive quiz specs document the implemented curation and draft spec
  paths.

## Implementation Notes

Raw quiz capture remains append-only Markdown. The repository can read it back
into quiz entries for curation, and missing raw capture resolves to an empty
entry set so done can still produce an inspectable curated file.

Draft specs are written under `.chc/specs/drafts/` as spec artifacts, not as a
separate draft document family. The service supplies provenance to the curated
quiz, raw quiz, dump context, model request, and model response.

## Contracts Added or Changed

- `internal/chc.ParseQuizRawCaptureEntries`
- `internal/chc.CurateQuizEntries`
- `internal/chc.QuizDraftSpec`
- `internal/chc.WriteQuizDraftSpec`
- `internal/chc.QuizModelActionDraftSpec`
- `internal/chc.QuizModelResponse.DraftSpec`
- `internal/chc.QuizArtifactRepository.LoadRawCaptureEntries`

## Files of Interest

- `internal/chc/quiz_curation.go`
- `internal/chc/quiz_draft_spec.go`
- `internal/chc/quiz_artifact_repository.go`
- `internal/chc/quiz_model.go`
- `internal/chcapp/service_quiz.go`
- `internal/chctui/quizzes.go`
- `.prj/specs/specs/tui-interactive-quiz-workflow.md`
- `.prj/specs/specs/tui-interactive-quiz-workflow-model.md`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chc ./internal/chcapp ./internal/chctui`
- `make verify`

## Risks and Follow-ups

- The initial curation pass is deterministic and intentionally compact; richer
  semantic curation can be layered later behind the same service boundary.
