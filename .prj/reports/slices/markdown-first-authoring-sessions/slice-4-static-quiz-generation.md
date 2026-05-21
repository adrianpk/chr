# Slice 4: Static Quiz Generation

Status: delivered
Slicing: Markdown-First Authoring Sessions
Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
Branch: `feat/slice-4-static-quiz-generation`
PR: #94

## Purpose

Generate editable Markdown questionnaires for Markdown-first authoring sessions
without requiring adaptive model-backed questioning.

## Delivered Behavior

The repository now has a static quiz generation service that can:

- validate static authoring session inputs;
- select target-specific question templates for every authoring target;
- plan a quiz artifact using the authoring artifact layout contract;
- render deterministic editable Markdown with answer slots;
- append the generated quiz output and provenance link to the session model;
- support dry-run planning without writing files;
- write quiz artifacts while respecting collision policy.

## Implementation Notes

Static quiz generation is a domain service. It does not add CLI commands and it
does not mutate session storage. Slice 5 can decide how command execution
persists the updated session and generated quiz artifact.

The templates are intentionally compact. They provide a useful starting point
for manual editing while leaving adaptive follow-up questioning to a later
assisted authoring slicing.

## Contracts Added or Changed

New internal contracts:

- `StaticQuizGenerationOptions`
- `StaticQuizGenerationPlan`
- `StaticQuizQuestion`
- `PlanStaticQuizGeneration`
- `RenderStaticQuizMarkdown`
- `WriteStaticQuiz`

## Files of Interest

- `internal/chc/authoring_static_quiz.go`
- `internal/chc/authoring_static_quiz_test.go`
- `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestPlanStaticQuizGeneration|TestWriteStaticQuiz'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice creates static questionnaire content only. It does not expose CLI
commands, persist updated sessions, support interactive questioning, or promote
quiz answers into drafts or specs.
