# Slice 3: Seeded CHC Context And ADR Docs

Status: reviewing
Slicing: QA Agent Work Scenario
Plan: `.prj/plans/slicing-qa-agent-work-scenario-plan.md`
Tracker: `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`
Branch: `feat/slice-3-seeded-chc-context-adr-docs`
PR: pending

## Purpose

Seed the blog QA reference with the ChrysoCoord context needed to generate the
comments spec, plan, tracking document, and tickets while also providing a
manual bug ticket for single-ticket execution.

## Delivered Behavior

- The blog reference includes `.chc/user/config.yml` and `.chc/user/manifest.md`.
- Project metadata describes the blog QA scenario.
- Use-case metadata defines `comments`.
- Comments context is seeded through a curated dump and quiz notes.
- A manual `ready_for_dev` bug ticket exists for the unknown-post status issue.
- An accepted ADR under `docs/explanation/adr/` records API boundary decisions
  that should influence comments generation.
- The reference intentionally has no generated comments specs, plans, tracking
  documents, or generated tickets.

## Implementation Notes

All ChrysoCoord-owned project artifacts live under `.chc/user/...`. ADRs are
normal project documentation under the Diataxis explanation quadrant instead of
hidden CHC-only context.

The seeded bug ticket points at the existing failing test and instructs the
agent to observe the failure before changing code.

## Contracts Added or Changed

- Active project after reset contains:
  - `.chc/user/project.md`
  - `.chc/user/cases/comments.md`
  - `.chc/user/dumps/comments-context.md`
  - `.chc/user/quizzes/comments-quiz.md`
  - `.chc/user/tickets/CHC-1-fix-unknown-post-status.md`
  - `docs/explanation/adr/0001-api-boundaries.md`
- `chc ticket list --dir qa/project/blog` shows the seeded manual bug ticket.
- `chc spec list --dir qa/project/blog` shows `comments` as missing.

## Files of Interest

- `qa/ref/001/blog/.chc/user/project.md`
- `qa/ref/001/blog/.chc/user/cases/comments.md`
- `qa/ref/001/blog/.chc/user/dumps/comments-context.md`
- `qa/ref/001/blog/.chc/user/quizzes/comments-quiz.md`
- `qa/ref/001/blog/.chc/user/tickets/CHC-1-fix-unknown-post-status.md`
- `qa/ref/001/blog/docs/explanation/adr/0001-api-boundaries.md`
- `.prj/trackers/slicing-qa-agent-work-scenario-tracker.md`

## Validation

- `make qa-reset-blog`
- `go run ./cmd/chc ticket list --dir qa/project/blog`
- `go run ./cmd/chc spec list --dir qa/project/blog`
- `find qa/project/blog/.chc/user/specs qa/project/blog/.chc/user/plans -maxdepth 2 -type f 2>/dev/null | sort`
- `make qa-test-blog` expected failure: unknown post returns `500`, test
  expects `404`
- `make verify`
- `git diff --check`

## Risks and Follow-ups

- Slice 4 still needs to add explicit test-based agent guidance.
- Comments generation itself is intentionally left for manual QA after later
  slices complete the scenario.
