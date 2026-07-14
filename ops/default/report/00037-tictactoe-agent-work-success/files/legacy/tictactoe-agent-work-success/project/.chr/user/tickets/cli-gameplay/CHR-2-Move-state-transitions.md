---
key: "CHR-2"
title: "Move state transitions"
synopsis: "Add game state that preserves turns on invalid moves and stops after terminal outcomes."
status: "reviewing"
priority: "high"
severity: "normal"
type: "feature"
tags: ["case-cli-gameplay", "area-internal-game", "type-implementation", "plan-cli-gameplay"]
use_case: "cli-gameplay"
project_area: "plan"
source_kind: "plan_tracking"
source_ref: ".chr/user/plans/cli-gameplay-plan.md"
slice: "S02"
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
updated_at: "2026-06-18T09:19:44Z"
---

# Description

Add turn-aware domain state and move application in `internal/game`.

Source plan: `.chr/user/plans/cli-gameplay-plan.md`
Tracking: `.chr/user/plans/cli-gameplay-tracking.md`
Slice: S02

## Objective

Add domain-level game state for current player, valid move application, invalid move rejection, and terminal states.

## Tasks

- Add table-driven transition tests for starting player, valid move alternation, invalid indexes, invalid marks, occupied cells, and terminal states.
- Introduce a small `Game` or equivalent state type around board, current player, and outcome accessors.
- Ensure `X` starts.
- Alternate turns only after valid non-terminal moves.
- Reject occupied cells, invalid indexes, invalid marks, and moves after game end without mutating state.
- Keep turn advancement and invalid move behavior in `internal/game`, not the command loop.

## Dependencies

- `S01` must be merged because move application depends on reliable winner and full-board detection.

## Acceptance Criteria

- `X` starts.
- Valid moves alternate `X` and `O`.
- Invalid moves preserve board state and current player.
- Occupied cells are rejected.
- Wins and draws prevent further mutation.
- Domain code has no process I/O and uses only Go's standard library.

## Validation Expectations

- Run `go test ./...`.
- Use table-driven tests for state transitions and invalid move behavior.
- Include terminal-state cases for win and draw outcomes.

## Branch and PR

- Base: `dev`
- Branch: `feature/cli-gameplay/s02-move-state`
- Target: `dev`
- PR scope: one PR limited to domain state, state transition tests, and outcome accessors needed by the CLI.

## References

- Tracking: `CLI Gameplay Execution Tracking`
- Source plan: `.chr/user/plans/cli-gameplay-plan.md`
- Project and case docs: `.chr/user/project.md`, `.chr/user/cases/cli-gameplay.md`
- Curated context: `.chr/user/dumps/cli-gameplay-context.md`, `.chr/user/dumps/project-context.md`, `.chr/user/quizzes/cli-gameplay-quiz.md`
- Agent constraints: `AGENTS.md`, `.agents/CONVENTIONS.md`
- Decision discovery: no ADR/RFC/decision-log/engineering-principles documents found.

# Dependencies

- S01

# Comments

## 2026-06-18T09:19:44Z chr

Execution `exec-409755cd38adfe67` finished with status `succeeded`.

Summary: Added turn-aware game state, move application, outcome accessors, and transition tests. Validated with go test ./...

## 2026-06-18T09:19:44Z chr

Execution exec-409755cd38adfe67 changed files:

- internal/game/game.go: Added Game state, Outcome values, current-player accessors, Play/ApplyMove, and terminal outcome handling.

- internal/game/game_test.go: Added table-driven tests for starting player, turn alternation, invalid move preservation, occupied cells, and terminal win/draw behavior.

# Blocker

# Validation

## 2026-06-18T09:19:44Z exec-409755cd38adfe67

- command: go test ./...
- status: succeeded

