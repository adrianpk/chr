---
title: "CLI Gameplay Implementation Plan"
project_id: "project-7c98c4f4cf644493"
case_slug: "cli-gameplay"
status: "planned"
source_request_id: "plan-generation-cli-gameplay"
source_response_id: "response-plan-generation-cli-gameplay"
created_at: "2026-06-18T08:47:29Z"
---

# CLI Gameplay Implementation Plan

## High-level implementation strategy

Close the CLI gameplay gap inside the existing Go seed project. The current repository already has `go.mod`, `cmd/tictactoe/main.go`, `internal/game/game.go`, and `internal/game/game_test.go`; the command still prints seed text, `Board.Winner` and `Board.Full` are incomplete, and tests currently cover parsing plus basic placement only.

Workspace decision discovery found no ADR, RFC, decision-log, engineering-principles, or similarly named architecture decision documents. The binding constraints are `AGENTS.md`, `.agents/CONVENTIONS.md`, the curated project and CLI gameplay docs, the quiz answers, and the existing spec and plan context: use only Go's standard library, keep rules in `internal/game`, keep process I/O in `cmd/tictactoe`, reject invalid moves without advancing the turn, and exclude replay, AI, persistence, networking, colors, profiles, and terminal UI packages.

Implement from the domain outward: finish board outcomes, add turn-aware state transitions, add deterministic CLI rendering, wire the interactive loop, then harden test coverage around rules, parsing, state transitions, and scripted CLI flows.

## Slice decomposition

### S01 - Game outcome rules

- slice identifier: `S01`
- title: Game outcome rules
- objective: Complete board-level win and full-board detection in `internal/game`.
- implementation rationale: The turn loop and terminal-state handling depend on reliable outcome detection before higher-level gameplay state can be implemented.
- dependencies: None.
- implementation notes: Add table-driven tests for all rows, columns, diagonals, no-winner states, full boards, and non-full boards. Implement `Board.Winner` and `Board.Full` without adding I/O or external packages.
- acceptance criteria: All eight win lines are detected; mixed or incomplete boards do not report false winners; full boards return true; boards with empty cells return false; `go test ./...` passes.
- estimated execution complexity: Low.
- associated tasks: `test(game): cover board outcome detection`; `feat(game): implement winner and full-board rules`.
- ticket metadata: `cli-gameplay-s01-game-outcomes`, labels `case:cli-gameplay`, `area:internal/game`, `type:implementation`, priority `P1`.
- branch metadata: base `dev`, branch `feature/cli-gameplay/s01-game-outcomes`, target `dev`.
- PR expectations: One focused PR limited to board outcome behavior and tests.

### S02 - Move state transitions

- slice identifier: `S02`
- title: Move state transitions
- objective: Add domain-level game state for current player, valid move application, invalid move rejection, and terminal states.
- implementation rationale: Turn advancement and invalid move behavior are game rules and should not be owned by the command loop.
- dependencies: `S01`.
- implementation notes: Introduce a small `Game` or equivalent state type around `Board`, current player, and outcome accessors. Start with `X`; alternate only after valid non-terminal moves; reject occupied cells, invalid indexes, invalid marks, and moves after game end without mutating state.
- acceptance criteria: `X` starts; valid moves alternate `X` and `O`; invalid moves preserve board and current player; occupied cells are rejected; wins and draws prevent further mutation; table-driven transition tests pass.
- estimated execution complexity: Medium.
- associated tasks: `test(game): cover move state transitions`; `feat(game): add turn-aware move application`.
- ticket metadata: `cli-gameplay-s02-move-state`, labels `case:cli-gameplay`, `area:internal/game`, `type:implementation`, priority `P1`.
- branch metadata: base `dev`, branch `feature/cli-gameplay/s02-move-state`, target `dev`.
- PR expectations: One PR limited to domain state, state transition tests, and outcome accessors needed by the CLI.

### S03 - CLI board rendering

- slice identifier: `S03`
- title: CLI board rendering
- objective: Add deterministic presentation helpers for board display, prompts, recoverable errors, and final result text.
- implementation rationale: Rendering should be testable before the interactive loop is introduced and should remain separate from domain rule decisions.
- dependencies: `S02`.
- implementation notes: Keep helpers in `cmd/tictactoe`. Display empty cells as positions `1` through `9` and occupied cells as `X` or `O`. Avoid colors, cursor control, terminal UI packages, and domain-rule duplication.
- acceptance criteria: Initial board clearly shows available positions; partial boards show player marks; prompts include current player; winner and draw messages are clear; rendering tests use only standard-library assertions.
- estimated execution complexity: Low.
- associated tasks: `test(cmd): cover board prompt and result rendering`; `feat(cmd): add CLI rendering helpers`.
- ticket metadata: `cli-gameplay-s03-cli-rendering`, labels `case:cli-gameplay`, `area:cmd/tictactoe`, `type:implementation`, priority `P2`.
- branch metadata: base `dev`, branch `feature/cli-gameplay/s03-cli-rendering`, target `dev`.
- PR expectations: One PR for presentation helpers and rendering tests only.

### S04 - Interactive gameplay loop

- slice identifier: `S04`
- title: Interactive gameplay loop
- objective: Replace the seed output with a complete stdin/stdout two-player game loop.
- implementation rationale: Once domain behavior and rendering are stable, the command can orchestrate the full game while keeping `main` thin and testable.
- dependencies: `S03`.
- implementation notes: Add an injectable command boundary such as `run(args []string, in io.Reader, out io.Writer, errOut io.Writer) int`. Preserve `-h`, `--help`, and `help`. Read moves line by line, parse positions through `internal/game`, report invalid text, out-of-range positions, and occupied cells as recoverable errors, and print the final board plus winner or draw before exit.
- acceptance criteria: Running `tictactoe` starts a full game; help still prints usage; invalid input is visible and does not advance the turn; occupied cells do not advance the turn; wins and draws exit cleanly with status `0`; scripted CLI tests pass; `go test ./...` passes.
- estimated execution complexity: Medium.
- associated tasks: `test(cmd): cover scripted gameplay flows`; `feat(cmd): implement interactive gameplay loop`.
- ticket metadata: `cli-gameplay-s04-gameplay-loop`, labels `case:cli-gameplay`, `area:cmd/tictactoe`, `type:implementation`, priority `P1`.
- branch metadata: base `dev`, branch `feature/cli-gameplay/s04-gameplay-loop`, target `dev`.
- PR expectations: One PR wiring command I/O to domain state; review should focus on turn preservation, clean termination, and help compatibility.

### S05 - Test coverage hardening

- slice identifier: `S05`
- title: Test coverage hardening
- objective: Bring completed behavior in line with the requested finished-test expectations.
- implementation rationale: The final slice closes gaps after implementation shape is stable, avoiding brittle tests against behavior that is still moving.
- dependencies: `S04`.
- implementation notes: Expand table-driven tests for any remaining parser, rule, validation, win, draw, and transition gaps. Add a Go fuzz test for `ParsePosition` because parsing is already a separate boundary. Add `testing/quick` only if a simple stable invariant is natural, such as valid parsed positions always mapping to indexes `0` through `8`. Do not add mutation testing tools or third-party assertions.
- acceptance criteria: Rule tests cover all win lines and draw states; state tests cover invalid moves without turn advancement; parser tests cover valid positions, whitespace, non-numeric input, and boundaries; CLI tests cover win, draw, help, invalid input recovery, and occupied-cell recovery; fuzz target compiles and runs with a short fuzz time; `go test ./...` passes.
- estimated execution complexity: Medium.
- associated tasks: `test(game): harden parser rule and transition coverage`; `test(cmd): complete CLI gameplay regression coverage`.
- ticket metadata: `cli-gameplay-s05-test-hardening`, labels `case:cli-gameplay`, `area:test`, `type:test`, priority `P1`.
- branch metadata: base `dev`, branch `feature/cli-gameplay/s05-test-hardening`, target `dev`.
- PR expectations: One PR for coverage only; review should reject brittle overmatching and any non-standard-library test helper.

## Slice ordering

1. `S01` completes foundational board outcomes.
2. `S02` builds turn-aware game state on those outcomes.
3. `S03` adds stable CLI rendering against the domain shape.
4. `S04` wires stdin/stdout gameplay using the domain and rendering helpers.
5. `S05` hardens final behavior coverage after implementation settles.

Execution should remain sequential. Later branches should start from updated `dev` after the previous slice merges.

## Slice dependency analysis

| Slice | Depends on | Reason |
| --- | --- | --- |
| `S01` | None | Outcome rules are required before state and CLI terminal behavior. |
| `S02` | `S01` | Move application must know when wins and draws end the game. |
| `S03` | `S02` | Rendering should reflect finalized board, player, and outcome accessors. |
| `S04` | `S03` | The loop needs stable board, prompt, error, and result text helpers. |
| `S05` | `S04` | Final test gaps can only be closed after full CLI behavior exists. |

No slice is meaningfully parallel without increasing rework risk.

## Tasks inside slices

| Slice | Task | Conventional commit target |
| --- | --- | --- |
| `S01` | Add board outcome tests for wins and full boards. | `test(game): cover board outcome detection` |
| `S01` | Implement winner and full-board rules. | `feat(game): implement winner and full-board rules` |
| `S02` | Add state transition tests for turn order, invalid moves, occupied cells, and terminal states. | `test(game): cover move state transitions` |
| `S02` | Add game state and turn-aware move application. | `feat(game): add turn-aware move application` |
| `S03` | Add rendering tests for empty boards, partial boards, prompts, and final results. | `test(cmd): cover board prompt and result rendering` |
| `S03` | Add board, prompt, invalid-message, and result rendering helpers. | `feat(cmd): add CLI rendering helpers` |
| `S04` | Add scripted command tests for help, win, draw, invalid input, and occupied-cell recovery. | `test(cmd): cover scripted gameplay flows` |
| `S04` | Implement the injectable gameplay loop and thin `main`. | `feat(cmd): implement interactive gameplay loop` |
| `S05` | Add parser fuzz coverage and optional stable invariant coverage. | `test(game): harden parser rule and transition coverage` |
| `S05` | Fill remaining CLI and domain regression gaps. | `test(cmd): complete CLI gameplay regression coverage` |

## Ticket generation data

| Ticket | Slice | Title | Branch | Area | Acceptance gate |
| --- | --- | --- | --- | --- | --- |
| `cli-gameplay-s01-game-outcomes` | `S01` | Game outcome rules | `feature/cli-gameplay/s01-game-outcomes` | `internal/game` | `go test ./...` |
| `cli-gameplay-s02-move-state` | `S02` | Move state transitions | `feature/cli-gameplay/s02-move-state` | `internal/game` | `go test ./...` |
| `cli-gameplay-s03-cli-rendering` | `S03` | CLI board rendering | `feature/cli-gameplay/s03-cli-rendering` | `cmd/tictactoe` | `go test ./...` |
| `cli-gameplay-s04-gameplay-loop` | `S04` | Interactive gameplay loop | `feature/cli-gameplay/s04-gameplay-loop` | `cmd/tictactoe` | `go test ./...` plus scripted run checks |
| `cli-gameplay-s05-test-hardening` | `S05` | Test coverage hardening | `feature/cli-gameplay/s05-test-hardening` | tests | `go test ./...` plus short parser fuzz run if fuzz target is added |

Each ticket should reference the project document, CLI gameplay use-case document, curated CLI gameplay dump, curated project dump, curated quiz, `AGENTS.md`, `.agents/CONVENTIONS.md`, and the discovered absence of ADR/RFC decision documents.

## Branch strategy

Use `dev` as the integration branch. Create one branch per slice from latest `dev` after the prior slice merges: `feature/cli-gameplay/s01-game-outcomes`, `feature/cli-gameplay/s02-move-state`, `feature/cli-gameplay/s03-cli-rendering`, `feature/cli-gameplay/s04-gameplay-loop`, and `feature/cli-gameplay/s05-test-hardening`.

Keep each branch scoped to its ticket. Do not add dependencies, new commands, persistence, networking, colors, replay, AI opponents, player profiles, or terminal UI packages.

## Pull request strategy

Open one pull request per slice into `dev`. Each PR should include a concise summary, validation commands, and intentionally deferred behavior. Domain PRs should not modify CLI text. CLI PRs should not change domain semantics unless the selected ticket explicitly requires it.

## Merge sequencing

Merge in slice order: `S01`, `S02`, `S03`, `S04`, then `S05`. After each merge, update local `dev`, rerun `go test ./...`, and create the next branch from the updated integration branch. Do not batch merge multiple slices unless each PR has passed its own review and validation.

## Review requirements

Reviewers must verify that rules remain in `internal/game`, process I/O remains in `cmd/tictactoe`, and all implementation and tests use only Go's standard library. Review must specifically check invalid text, out-of-range positions, occupied cells, turn preservation after invalid input, all row/column/diagonal wins, draw detection, final board printing, help behavior, and absence of optional replay or non-goal features.

## Completion validation criteria

The use case is complete when `go test ./...` passes, `tictactoe` can run a full local two-player game from stdin/stdout, the initial board shows positions `1` through `9`, `X` starts, turns alternate only after valid moves, invalid text and out-of-range positions are recoverable without advancing the turn, occupied cells are rejected without advancing the turn, all wins and draws produce a clear final result, and the final board is printed before exit.

Additional validation should include `go run ./cmd/tictactoe --help`, one scripted `X` win, one scripted draw, one scripted invalid-input recovery flow, one scripted occupied-cell recovery flow, and a short `go test -fuzz=FuzzParsePosition -fuzztime=5s ./internal/game` run if `S05` adds the parser fuzz target.

## Provenance

- source: project_context
- source: request_context
- source: quiz_context
- source: request_context
- source: /home/adrian/Projects/chryso/qa/project/tictactoe/AGENTS.md
- source: /home/adrian/Projects/chryso/qa/project/tictactoe/.chr/default/project.md
- source: /home/adrian/Projects/chryso/qa/project/tictactoe/.chr/default/cases/cli-gameplay.md
- source: .chr/default/dumps/cli-gameplay-context.md
- source: .chr/default/dumps/project-context.md
- source: .chr/default/plans/cli-gameplay-plan.md
- source: .chr/default/quizzes/cli-gameplay-quiz.md
