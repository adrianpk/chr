---
key: "CHR-3"
title: "CLI board rendering"
synopsis: "Add testable board, prompt, error, and result rendering for the CLI."
status: "reviewing"
priority: "normal"
severity: "normal"
type: "feature"
tags: ["case-cli-gameplay", "area-cmd-tictactoe", "type-implementation", "plan-cli-gameplay"]
use_case: "cli-gameplay"
project_area: "plan"
source_kind: "plan_tracking"
source_ref: ".chr/default/plans/cli-gameplay-plan.md"
slice: "S03"
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
updated_at: "2026-06-18T09:24:37Z"
---

# Description

Add deterministic CLI rendering helpers and tests in `cmd/tictactoe`.

Source plan: `.chr/default/plans/cli-gameplay-plan.md`
Tracking: `.chr/default/plans/cli-gameplay-tracking.md`
Slice: S03

## Objective

Add deterministic presentation helpers for board display, prompts, recoverable errors, and final result text.

## Tasks

- Add rendering tests for empty boards, partial boards, prompts, recoverable error text, winner messages, and draw messages.
- Add board rendering helpers in `cmd/tictactoe`.
- Display empty cells as positions `1` through `9` and occupied cells as `X` or `O`.
- Add prompt helpers that include the current player.
- Add result helpers for winner and draw outcomes.
- Avoid colors, cursor control, terminal UI packages, and domain-rule duplication.

## Dependencies

- `S02` must be merged because rendering should reflect finalized board, player, and outcome accessors.

## Acceptance Criteria

- Initial board clearly shows available positions.
- Partial boards show player marks.
- Prompts include the current player.
- Winner and draw messages are clear and deterministic.
- Rendering code stays in `cmd/tictactoe` and does not own domain rules.
- Tests use only Go's standard library.

## Validation Expectations

- Run `go test ./...`.
- Use standard-library assertions for deterministic rendering tests.
- Avoid brittle overmatching beyond text that is part of the user-facing contract.

## Branch and PR

- Base: `dev`
- Branch: `feature/cli-gameplay/s03-cli-rendering`
- Target: `dev`
- PR scope: one PR for presentation helpers and rendering tests only.

## References

- Tracking: `CLI Gameplay Execution Tracking`
- Source plan: `.chr/default/plans/cli-gameplay-plan.md`
- Project and case docs: `.chr/default/project.md`, `.chr/default/cases/cli-gameplay.md`
- Curated context: `.chr/default/dumps/cli-gameplay-context.md`, `.chr/default/dumps/project-context.md`, `.chr/default/quizzes/cli-gameplay-quiz.md`
- Agent constraints: `AGENTS.md`, `.agents/CONVENTIONS.md`
- Decision discovery: no ADR/RFC/decision-log/engineering-principles documents found.

# Dependencies

- S02

# Comments

## 2026-06-18T09:24:37Z chr

Execution `exec-c4b72e7ae4e31f4d` finished with status `succeeded`.

Summary: Added deterministic CLI board, prompt, recoverable error, and result rendering helpers with standard-library tests. Validation passed with go test ./...

## 2026-06-18T09:24:37Z chr

Execution exec-c4b72e7ae4e31f4d changed files:

- cmd/tictactoe/render.go: Adds deterministic board, prompt, recoverable error, and game result rendering helpers.

- cmd/tictactoe/render_test.go: Covers empty and partial boards, prompt text, recoverable error text, winner messages, and draw messages.

# Blocker

# Validation

## 2026-06-18T09:24:37Z exec-c4b72e7ae4e31f4d

- command: go test ./...
- status: succeeded

