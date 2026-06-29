---
project_id: "project-tictactoe-qa"
case_slug: "cli-gameplay"
scope: "cli-gameplay"
status: "curated"
raw_path: ".chr/user/quizzes/cli-gameplay-quiz-raw.md"
updated_at: "2026-06-18T07:52:42Z"
---

# CLI Gameplay Quiz

## Decisions

- Build a local two-player tic-tac-toe CLI.
- Use positions `1` through `9` for moves.
- `X` starts and turns alternate after valid moves.
- Invalid input does not advance the turn.
- The game ends on a win or draw and prints a clear final result.

## Constraints

- Use only Go's standard library.
- Keep domain rules in `internal/game`.
- Keep command I/O in `cmd/tictactoe`.
- Exclude AI, persistence, networking, colors, terminal UI frameworks, and
  player accounts.

## Acceptance Signals

- A user can complete a full game from stdin/stdout.
- Invalid input is visible and recoverable.
- All win lines and draw states are covered by tests.
- The package-level rule tests and CLI-facing tests can run with `go test ./...`.

## Testing Guidance

- Prefer table-driven tests for rule cases.
- Add fuzz coverage for parsing if parsing remains a separate function.
- Add `testing/quick` coverage only for simple invariants that do not make the
  test suite brittle.
- Do not add mutation-testing tools.

## Open Questions

- Replay after game end is optional and should not block the first useful
  version.

## Provenance

- raw quiz: `.chr/user/quizzes/cli-gameplay-quiz-raw.md`
