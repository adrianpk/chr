# Slice 5: Context-Aware Quiz Generation Foundation

Status: delivered
Slicing: UX Operation Surface
Plan: `.prj/plans/slicing-ux-operation-surface-plan.md`
Tracker: `.prj/trackers/slicing-ux-operation-surface-tracker.md`
Branch: `feat/slice-5-context-aware-quiz-generation`
PR: #138

## Purpose

Replace generic quiz templates with a model-backed quiz generation boundary that
can ask project-specific questions from structured project context.

## Delivered Behavior

- Added contextual quiz request and result models.
- Added context sections for project identity, authoring session, `AGENTS.md`,
  `.chr/` artifacts, docs, project tree, and Go inventory.
- Added an output contract for structured quiz questions.
- Added strict response validation before writing quiz artifacts.
- Added a Codex-backed question provider for the CLI path.
- Added a deterministic fake provider for tests.
- Removed static quiz generation from executable code.
- Changed `author quiz new` to generate context-aware quizzes only.
- Preserved quiz-to-draft promotion over generated Markdown quiz artifacts.
- Updated documentation and TUI specs to remove static quiz wording.

## Implementation Notes

The quiz provider receives a structured request with action
`authoring_quiz_generation`, strategy `deep`, target artifact, topic, requested
question count, context sections, and output contract.

The Codex provider invokes `codex exec --skip-git-repo-check` from the project
directory. It expects strict JSON with a top-level `questions` array. The command
fails if the provider response cannot be parsed, has no valid questions, or asks
for sensitive material.

`--dry-run` still calls the provider because the useful preview is the generated
question set, not a local template.

## Contracts Added or Changed

- Added `AuthoringSessionModeContextual`.
- Added `AuthoringQuizGenerationRequest`.
- Added `AuthoringQuizGenerationResult`.
- Added `AuthoringQuizQuestionProvider`.
- Added `CodexAuthoringQuizQuestionProvider`.
- Removed `StaticQuizGenerationOptions`, `StaticQuizGenerationPlan`,
  `PlanStaticQuizGeneration`, `RenderStaticQuizMarkdown`, and `WriteStaticQuiz`.

## Files of Interest

- `internal/chr/authoring_contextual_quiz.go`
- `internal/chr/authoring_contextual_quiz_test.go`
- `internal/chrapp/author.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/authoring-sessions.md`
- `docs/explanation/document-generation-model.md`

## Validation

```sh
go test ./internal/chr ./internal/chrapp
make verify
```

All commands pass.

## Risks and Follow-ups

- The first provider boundary shells out to Codex directly; later slices can add
  richer provider configuration and trace persistence.
- Fast/batch quiz modes remain a later UX optimization. The foundation keeps
  `deep` as the quality path.
