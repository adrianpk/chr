# Tictactoe Conventions

- Keep game rules in `internal/game`.
- Keep `cmd/tictactoe` focused on argument parsing, prompts, and process I/O.
- Use only Go's standard library.
- Prefer table-driven tests for rules and input parsing.
- Use Go fuzz tests for user input parsing when useful.
- Use `testing/quick` for small invariant-style checks when the behavior is
  stable enough to express as a property.
- Do not add mutation-testing tools or third-party assertion libraries.
