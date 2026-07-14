---
key: "CHR-1"
title: "Game outcome rules"
synopsis: "Finish tic-tac-toe board win and full-board rules."
status: "reviewing"
priority: "high"
severity: "normal"
type: "feature"
tags: ["case-cli-gameplay", "area-internal-game", "type-implementation", "plan-cli-gameplay"]
use_case: "cli-gameplay"
project_area: "plan"
source_kind: "plan_tracking"
source_ref: ".chr/user/plans/cli-gameplay-plan.md"
slice: "S01"
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
updated_at: "2026-06-18T09:15:09Z"
---

# Description

Complete board winner and full-board detection in `internal/game`.

Source plan: `.chr/user/plans/cli-gameplay-plan.md`
Tracking: `.chr/user/plans/cli-gameplay-tracking.md`
Slice: S01

## Objective

Complete board-level win and full-board detection in `internal/game`.

## Tasks

- Add table-driven tests for all three rows, all three columns, both diagonals, no-winner boards, full boards, and non-full boards.
- Implement `Board.Winner` without I/O, CLI concerns, or external packages.
- Implement `Board.Full` without I/O, CLI concerns, or external packages.
- Keep the PR limited to board outcome behavior and tests.

## Dependencies

None. This is the foundational slice for later state and CLI terminal behavior.

## Acceptance Criteria

- All eight win lines are detected.
- Mixed or incomplete boards do not report false winners.
- Full boards return true.
- Boards with empty cells return false.
- Game-rule code remains in `internal/game`.
- No external dependencies or third-party assertion libraries are added.

## Validation Expectations

- Run `go test ./...`.
- Use table-driven tests for winner and full-board input/output behavior.

## Branch and PR

- Base: `dev`
- Branch: `feature/cli-gameplay/s01-game-outcomes`
- Target: `dev`
- PR scope: one focused PR limited to board outcome behavior and tests.

## References

- Tracking: `CLI Gameplay Execution Tracking`
- Source plan: `.chr/user/plans/cli-gameplay-plan.md`
- Project and case docs: `.chr/user/project.md`, `.chr/user/cases/cli-gameplay.md`
- Curated context: `.chr/user/dumps/cli-gameplay-context.md`, `.chr/user/dumps/project-context.md`, `.chr/user/quizzes/cli-gameplay-quiz.md`
- Agent constraints: `AGENTS.md`, `.agents/CONVENTIONS.md`
- Decision discovery: no ADR/RFC/decision-log/engineering-principles documents found.

# Dependencies

None.

# Comments

## 2026-06-18T09:15:09Z chr

Execution `exec-648ea4e465b090f7` finished with status `succeeded`.

Summary: Implemented tic-tac-toe board winner and full-board detection with table-driven tests for all win lines, no-winner boards, full boards, and non-full boards.

## 2026-06-18T09:15:09Z chr

Execution exec-648ea4e465b090f7 changed files:

- internal/game/game.go: Implemented Board.Winner and Board.Full.

- internal/game/game_test.go: Added table-driven outcome tests for winner and full-board behavior.

# Blocker

# Validation

## 2026-06-18T09:15:09Z exec-648ea4e465b090f7

- command: go test ./...
- status: succeeded

