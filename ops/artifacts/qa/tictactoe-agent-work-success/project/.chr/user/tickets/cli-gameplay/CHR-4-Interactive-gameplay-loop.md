---
key: "CHR-4"
title: "Interactive gameplay loop"
synopsis: "Replace seed output with the full two-player CLI game loop."
status: "reviewing"
priority: "high"
severity: "normal"
type: "feature"
tags: ["case-cli-gameplay", "area-cmd-tictactoe", "type-implementation", "plan-cli-gameplay"]
use_case: "cli-gameplay"
project_area: "plan"
source_kind: "plan_tracking"
source_ref: ".chr/user/plans/cli-gameplay-plan.md"
slice: "S04"
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
updated_at: "2026-06-18T09:28:36Z"
---

# Description

Wire the interactive stdin/stdout gameplay loop in `cmd/tictactoe`.

Source plan: `.chr/user/plans/cli-gameplay-plan.md`
Tracking: `.chr/user/plans/cli-gameplay-tracking.md`
Slice: S04

## Objective

Replace the seed output with a complete stdin/stdout two-player gameplay loop.

## Tasks

- Add scripted command tests for help, one `X` win, one draw, invalid input recovery, and occupied-cell recovery.
- Implement an injectable command boundary such as `run(args []string, in io.Reader, out io.Writer, errOut io.Writer) int`.
- Keep `main` thin and delegate process behavior to the injectable boundary.
- Preserve `-h`, `--help`, and `help` behavior.
- Read moves line by line and parse positions through `internal/game`.
- Report invalid text, out-of-range positions, and occupied cells as recoverable errors.
- Ensure invalid input and occupied-cell attempts do not advance the turn.
- Print the final board plus winner or draw before clean exit.

## Dependencies

- `S03` must be merged because the loop depends on stable board, prompt, error, and result rendering helpers.

## Acceptance Criteria

- Running `tictactoe` starts a full local two-player game.
- Help still prints usage.
- Invalid input is visible and does not advance the turn.
- Occupied cells do not advance the turn.
- Wins and draws exit cleanly with status `0`.
- The final board is printed before the final result.
- Process I/O remains in `cmd/tictactoe`; game rules remain in `internal/game`.

## Validation Expectations

- Run `go test ./...`.
- Add scripted CLI tests for help, win, draw, invalid input recovery, and occupied-cell recovery.
- Include manual or scripted `go run ./cmd/tictactoe --help` validation.
- Keep tests standard-library only.

## Branch and PR

- Base: `dev`
- Branch: `feature/cli-gameplay/s04-gameplay-loop`
- Target: `dev`
- PR scope: one PR wiring command I/O to domain state. Review should focus on turn preservation, clean termination, final board printing, and help compatibility.

## References

- Tracking: `CLI Gameplay Execution Tracking`
- Source plan: `.chr/user/plans/cli-gameplay-plan.md`
- Project and case docs: `.chr/user/project.md`, `.chr/user/cases/cli-gameplay.md`
- Curated context: `.chr/user/dumps/cli-gameplay-context.md`, `.chr/user/dumps/project-context.md`, `.chr/user/quizzes/cli-gameplay-quiz.md`
- Agent constraints: `AGENTS.md`, `.agents/CONVENTIONS.md`
- Decision discovery: no ADR/RFC/decision-log/engineering-principles documents found.

# Dependencies

- S03

# Comments

## 2026-06-18T09:28:36Z chr

Execution `exec-6326a1544f559788` finished with status `succeeded`.

Summary: Implemented the interactive two-player CLI gameplay loop with injectable I/O, preserved help behavior, and added scripted command tests for wins, draw, invalid input, out-of-range input, and occupied-cell recovery.

## 2026-06-18T09:28:36Z chr

Execution exec-6326a1544f559788 changed files:

- cmd/tictactoe/main.go: Added run(args, in, out, errOut), delegated main through it, wired stdin/stdout gameplay, recoverable move errors, and final result rendering.

- cmd/tictactoe/main_test.go: Added scripted CLI tests for help, X win, draw, invalid text recovery, out-of-range recovery, and occupied-cell recovery.

# Blocker

# Validation

## 2026-06-18T09:28:36Z exec-6326a1544f559788

- command: go test ./...
- status: succeeded

