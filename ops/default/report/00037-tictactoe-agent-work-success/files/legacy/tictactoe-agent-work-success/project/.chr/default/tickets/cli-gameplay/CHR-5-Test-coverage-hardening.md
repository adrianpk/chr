---
key: "CHR-5"
title: "Test coverage hardening"
synopsis: "Complete final test coverage for parser, rules, state, and scripted CLI behavior."
status: "reviewing"
priority: "high"
severity: "normal"
type: "task"
tags: ["case-cli-gameplay", "area-test", "type-test", "plan-cli-gameplay"]
use_case: "cli-gameplay"
project_area: "plan"
source_kind: "plan_tracking"
source_ref: ".chr/default/plans/cli-gameplay-plan.md"
slice: "S05"
branch: ""
review_refs: []
review_state: ""
review_result: ""
review_comments_url: ""
review_round: ""
blocking_findings: []
merge_state: ""
merge_result: ""
merge_actor: ""
merge_commit: ""
merge_blocker: ""
merged_at: ""
created_at: "2026-06-18T08:53:02Z"
updated_at: "2026-06-18T09:32:10Z"
---

# Description

Harden parser, domain, and CLI regression coverage after implementation settles.

Source plan: `.chr/default/plans/cli-gameplay-plan.md`
Tracking: `.chr/default/plans/cli-gameplay-tracking.md`
Slice: S05

## Objective

Bring completed behavior in line with the requested finished-test expectations after implementation shape is stable.

## Tasks

- Fill remaining table-driven parser, rule, validation, win, draw, and transition coverage gaps.
- Add or complete CLI regression coverage for win, draw, help, invalid input recovery, and occupied-cell recovery.
- Add a Go fuzz test for `ParsePosition` if parsing remains a separate boundary.
- Add `testing/quick` only if a simple stable invariant is natural, such as valid parsed positions mapping to indexes `0` through `8`.
- Do not add mutation testing tools, third-party assertions, or non-standard-library helpers.
- Avoid brittle overmatching of CLI output beyond stable user-facing behavior.

## Dependencies

- `S04` must be merged because final coverage should target the complete CLI and domain behavior.

## Acceptance Criteria

- Rule tests cover all win lines and draw states.
- State tests cover invalid moves without turn advancement.
- Parser tests cover valid positions, whitespace, non-numeric input, and boundaries.
- CLI tests cover win, draw, help, invalid input recovery, and occupied-cell recovery.
- Any fuzz target compiles and runs with a short fuzz time.
- Test coverage remains standard-library only.

## Validation Expectations

- Run `go test ./...`.
- If `FuzzParsePosition` is added, run `go test -fuzz=FuzzParsePosition -fuzztime=5s ./internal/game`.
- Prefer table-driven tests for input/output behavior.
- Use fuzz or property-style tests only where the invariant is stable and useful.

## Branch and PR

- Base: `dev`
- Branch: `feature/cli-gameplay/s05-test-hardening`
- Target: `dev`
- PR scope: one PR for coverage only. Review should reject brittle overmatching and any non-standard-library test helper.

## References

- Tracking: `CLI Gameplay Execution Tracking`
- Source plan: `.chr/default/plans/cli-gameplay-plan.md`
- Project and case docs: `.chr/default/project.md`, `.chr/default/cases/cli-gameplay.md`
- Curated context: `.chr/default/dumps/cli-gameplay-context.md`, `.chr/default/dumps/project-context.md`, `.chr/default/quizzes/cli-gameplay-quiz.md`
- Agent constraints: `AGENTS.md`, `.agents/CONVENTIONS.md`
- Decision discovery: no ADR/RFC/decision-log/engineering-principles documents found.

# Dependencies

- S04

# Comments

## 2026-06-18T09:32:10Z chr

Execution `exec-7fe25b30b4955e1e` finished with status `succeeded`.

Summary: Hardened internal game test coverage with additional parser boundary cases, a ParsePosition quick-check invariant, a ParsePosition fuzz target, and direct Board invalid-move validation tests. Required go test and short fuzz validation passed.

## 2026-06-18T09:32:10Z chr

Execution exec-7fe25b30b4955e1e changed files:

- internal/game/game_test.go: Added parser boundary cases, testing/quick invariant coverage, FuzzParsePosition, and direct Board At/Place invalid-state coverage.

# Blocker

# Validation

## 2026-06-18T09:32:10Z exec-7fe25b30b4955e1e

- command: go test ./...
- status: succeeded

