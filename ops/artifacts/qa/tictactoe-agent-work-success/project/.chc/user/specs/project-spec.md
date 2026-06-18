---
title: "Tictactoe CLI Gameplay Specification"
project_id: "project-7c98c4f4cf644493"
case_slug: ""
status: "draft"
source_request_id: "spec-generation-project"
source_response_id: "response-spec-generation-project"
created_at: "2026-06-18T08:45:31Z"
---

# Tictactoe CLI Gameplay Specification

**Purpose**

This specification defines the first useful version of the Tictactoe project: a small Go command-line tic-tac-toe game for two human players sharing one terminal. The game must run locally, use standard input and standard output, and remain simple enough to serve as a ChrysoCoord QA reference project.

**Decision Discovery**

Workspace discovery found no ADR, RFC, decision-log, engineering-principles, or similarly named architecture decision document. The highest-authority constraints for this spec are the project agent protocol, `.agents/CONVENTIONS.md`, the curated project document, the curated CLI gameplay dump, and the curated CLI gameplay quiz.

No discovered decision conflicts with the curated dumps or quiz notes.

**Scope**

The implementation must provide a complete local two-player tic-tac-toe CLI:

- The command is run as `tictactoe`.
- The game displays a 3x3 board.
- Empty cells are shown as available move positions `1` through `9`.
- Two human players alternate turns in one terminal.
- Player `X` always starts.
- Player `O` follows after each valid `X` move.
- The game accepts moves from standard input.
- The game rejects invalid input, out-of-range positions, and occupied cells.
- Invalid moves do not advance the turn.
- The game detects wins, draws, and clean termination.
- The game prints the final board and a clear final result before exiting.

**Non-Goals**

The first useful version must not add:

- External Go dependencies.
- Persistence or saved games.
- Networking or services.
- Non-CLI interfaces.
- Terminal UI frameworks.
- Colors or terminal control behavior.
- AI opponents.
- Player accounts or profiles.
- Required replay after game end.

Replay after game end remains an open optional extension and must not block the first useful version.

**Architecture Constraints**

The code must keep responsibilities separated:

- Domain game rules belong in `internal/game`.
- Command argument handling, prompts, rendering, stdin/stdout, and process behavior belong in `cmd/tictactoe`.
- Domain rule code must not depend on process I/O.
- Command code may orchestrate I/O and call domain behavior, but should not own game-rule decisions such as winner detection, occupied-cell validation, or turn advancement.
- Only Go's standard library may be used.

The current seed already has `cmd/tictactoe/main.go`, `internal/game/game.go`, and `internal/game/game_test.go`. The gameplay loop is not implemented yet, and `Board.Winner` / `Board.Full` are currently incomplete. The spec requires completing behavior without changing the repository into a different application shape.

**Domain Behavior**

The game board has exactly nine cells arranged as a 3x3 grid. Position input maps to cells as positions `1` through `9`. Existing parser behavior maps valid user-facing positions to zero-based internal indexes.

A move is valid only when:

- The input parses as a position from `1` through `9`.
- The target cell is not already occupied.
- The game has not already ended.
- The mark being placed is either `X` or `O`.

A valid move places the current player's mark and then advances the turn only if the game is not over. Invalid moves must leave the board and current player unchanged.

Win detection must cover all eight winning lines:

- Three rows.
- Three columns.
- Two diagonals.

A draw occurs when the board is full and there is no winner.

The game ends immediately when a player wins or when a draw is reached. After the game has ended, no further move should mutate game state.

**CLI Behavior**

Running `tictactoe` with no help argument must start the local game. The initial output must show the empty board with available positions. Each turn must make the current player visible and accept one move from standard input.

Help behavior already exists for `-h`, `--help`, and `help`; implementation work should preserve that behavior unless a generated ticket explicitly narrows it.

Invalid input must be visible and recoverable. Examples of invalid input include non-numeric text, positions outside `1` through `9`, and choosing an occupied cell. After any invalid move, the same player remains on turn.

At game end, the CLI must print the final board and a clear result indicating either the winning player or a draw. Normal win and draw completion should exit cleanly.

**Testing Requirements**

Tests must use only Go's standard library.

Required coverage:

- Position parsing for valid positions, whitespace handling, non-numeric input, and range boundaries.
- Board placement for valid moves, occupied cells, invalid indexes, and invalid marks.
- Winner detection for every row, column, and diagonal.
- Full-board detection.
- Draw detection when the board is full with no winner.
- Turn order starting with `X`.
- Turn advancement only after valid moves.
- Invalid input and occupied-cell behavior preserving the current turn.
- Terminal game states preventing further state mutation.
- CLI-facing scripted flows for at least one win, one draw, help output, invalid input recovery, and occupied-cell recovery.

Table-driven tests are preferred for rules, parsing, validation, wins, draws, and state transitions. A Go fuzz test is appropriate for input parsing if parsing remains a separate function. `testing/quick` may be used only for simple stable invariants. Mutation testing tools and third-party assertion libraries are out of scope.

The baseline validation command is:

```shell
go test ./...
```

Additional CLI validation should include scripted runs for a win, a draw, invalid input recovery, and help behavior once the interactive loop exists.

**Acceptance Criteria**

The first useful version is accepted when:

- `tictactoe` can run a full local two-player game from stdin/stdout.
- The board clearly displays available positions `1` through `9`.
- `X` starts.
- Turns alternate only after valid moves.
- Invalid text is rejected without advancing the turn.
- Out-of-range positions are rejected without advancing the turn.
- Occupied cells are rejected without advancing the turn.
- All row, column, and diagonal wins are detected.
- Draws are detected when the board is full with no winner.
- The final board is printed before exit.
- The final result clearly reports the winner or draw.
- The implementation remains standard-library only.
- Rules remain in `internal/game` and process I/O remains in `cmd/tictactoe`.
- `go test ./...` passes.

**Implementation Slices**

The following slice structure is compatible with the generated project context and should remain secondary to the behavioral requirements above.

| Slice | Area | Objective | Acceptance Gate |
| --- | --- | --- | --- |
| `S01` | `internal/game` | Complete board winner and full-board detection. | All eight win lines and full-board states covered by tests; `go test ./...` passes. |
| `S02` | `internal/game` | Add game state for current player, valid move application, invalid move rejection, and terminal states. | Turn order and invalid move semantics are tested; `go test ./...` passes. |
| `S03` | `cmd/tictactoe` | Add deterministic board, prompt, invalid message, and result rendering. | Empty and marked boards are clear; result text is testable; `go test ./...` passes. |
| `S04` | `cmd/tictactoe` | Wire the interactive stdin/stdout gameplay loop. | Scripted win, draw, invalid input, occupied cell, and help flows pass; `go test ./...` passes. |
| `S05` | tests | Harden parser, domain, and CLI-facing regression coverage. | Coverage matches required behavior; optional fuzz target compiles if added; `go test ./...` passes. |

Generated tickets should keep branch and PR scope aligned with the selected slice. Domain slices should avoid CLI text churn, and CLI slices should avoid changing domain semantics unless the selected ticket explicitly requires it.

**Open Questions**

Replay after game end is optional. It is not required for the first useful version and should not be implemented unless a later generated plan or ticket explicitly selects it as follow-up work.

No other unresolved behavior is required to complete the first useful version.

## Provenance

- source: project_context
- source: request_context
- source: quiz_context
- source: request_context
- source: /home/adrian/Projects/chrysocoord/qa/project/tictactoe/AGENTS.md
- source: /home/adrian/Projects/chrysocoord/qa/project/tictactoe/.chc/user/project.md
- source: .chc/user/dumps/cli-gameplay-context.md
- source: .chc/user/dumps/project-context.md
- source: .chc/user/quizzes/cli-gameplay-quiz.md
