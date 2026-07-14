# Slice 3: Seeded CHR Context And ADR Docs

Status: reviewing
Slicing: QA Agent Work Scenario
Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`
Tracker: `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`
Branch: `feat/slice-3-seeded-chr-context-adr-docs`
PR: pending

## Purpose

Seed the blog QA reference with the Chryso context needed to generate the
comments spec, plan, tracking document, and tickets while also providing a
manual bug ticket for single-ticket execution.

## Delivered Behavior

- The blog reference includes `.chr/default/config.yml` and `.chr/default/manifest.md`.
- Project metadata describes the blog QA scenario.
- Use-case metadata defines `comments`.
- Comments context is seeded through a curated dump and quiz notes.
- A manual `ready_for_dev` bug ticket exists for the unknown-post status issue.
- An accepted ADR under `docs/explanation/adr/` records API boundary decisions
  that should influence comments generation.
- The reference intentionally has no generated comments specs, plans, tracking
  documents, or generated tickets.

## Implementation Notes

All Chryso-owned project artifacts live under `.chr/default/...`. ADRs are
normal project documentation under the Diataxis explanation quadrant instead of
hidden CHR-only context.

The seeded bug ticket points at the existing failing test and instructs the
agent to observe the failure before changing code.

## Contracts Added or Changed

- Active project after reset contains:
  - `.chr/default/project.md`
  - `.chr/default/cases/comments.md`
  - `.chr/default/dumps/comments-context.md`
  - `.chr/default/quizzes/comments-quiz.md`
  - `.chr/default/tickets/CHR-1-fix-unknown-post-status.md`
  - `docs/explanation/adr/0001-api-boundaries.md`
- `chr ticket list --dir qa/project/blog` shows the seeded manual bug ticket.
- `chr spec list --dir qa/project/blog` shows `comments` as missing.

## Files of Interest

- `qa/ref/001/blog/.chr/default/project.md`
- `qa/ref/001/blog/.chr/default/cases/comments.md`
- `qa/ref/001/blog/.chr/default/dumps/comments-context.md`
- `qa/ref/001/blog/.chr/default/quizzes/comments-quiz.md`
- `qa/ref/001/blog/.chr/default/tickets/CHR-1-fix-unknown-post-status.md`
- `qa/ref/001/blog/docs/explanation/adr/0001-api-boundaries.md`
- `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`

## Validation

- `make qa-reset-blog`
- `go run ./cmd/chr ticket list --dir qa/project/blog`
- `go run ./cmd/chr spec list --dir qa/project/blog`
- `find qa/project/blog/.chr/default/specs qa/project/blog/.chr/default/plans -maxdepth 2 -type f 2>/dev/null | sort`
- `make qa-test-blog` expected failure: unknown post returns `500`, test
  expects `404`
- `make verify`
- `git diff --check`

## Risks and Follow-ups

- Slice 4 still needs to add explicit test-based agent guidance.
- Comments generation itself is intentionally left for manual QA after later
  slices complete the scenario.
