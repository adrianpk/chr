# Slice 4: Quiz Parity Lab Validation

Status: delivered
Slicing: Quiz Loop RedWright Parity
Plan: `.prj/plans/slicing-quiz-loop-redwright-parity-plan.md`
Tracker: `.prj/trackers/slicing-quiz-loop-redwright-parity-tracker.md`
Branch: `feat/slice-4-quiz-parity-lab-validation`
PR: #196

## Purpose

Record the concrete lab validation path for the corrected TUI quiz loop and
preserve evidence about why old runtime state must be reset before judging the
new behavior.

## Delivered Behavior

- Adds TUI quiz parity reset guidance to the full-system lab runbook.
- Adds project-level and use-case-level TUI quiz validation steps.
- Records expected evidence for one-question-at-a-time behavior, model activity,
  prompt-intent persistence, and skip-driven trunk recovery.
- Records observed stale lab state that predates this slicing round.

## Implementation Notes

The current `chc-blog-work` lab project contained interactive quiz runtime files
generated before this parity slicing:

- `.chc/tmp/quiz-runtime/project.json`
- `.chc/tmp/quiz-runtime/article-authoring.json`
- `.chc/tmp/quiz-runtime/content-import.json`

The active questions in those files still show the pre-fix symptom: verbose
questions that ask for many details at once. Those files are useful evidence,
but they must not be reused as proof of current behavior after this slicing.

The runbook now defines a quiz-specific reset that removes only interactive quiz
runtime and interactive quiz artifacts. It intentionally keeps project metadata,
use-case metadata, dumps, tickets, specs, and authoring-session quizzes.

## Contracts Added or Changed

- Manual TUI quiz parity validation starts from clean quiz runtime state.
- Old `.chc/tmp/quiz-runtime/*.json` files are not valid evidence for current
  question-generation behavior after prompt and routing fixes.
- The default TUI flow must produce one active question per scope, not a visible
  pre-generated batch.
- A skipped follow-up branch must return to broader coverage.

## Files of Interest

- `docs/how-to/full-system-lab-runbook.md`
  - adds reset and validation steps for TUI quiz parity
- `.prj/reports/slices/quiz-loop-redwright-parity/slice-4-quiz-parity-lab-validation.md`
  - records lab evidence and validation expectations

## Validation

- `go run ./cmd/chc --help`
- `git diff --check`
- `make verify`

## Risks and Follow-ups

- Manual TUI validation still depends on a live model key and user-driven
  interaction.
- Fast mode parity remains a later tuning concern if lab evidence shows the deep
  loop is acceptable.
