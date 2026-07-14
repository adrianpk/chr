---
title: "CLI Gameplay Context"
status: "curated"
type: "dump"
case: "cli-gameplay"
raw_path: "ops/default/dump/<id>/artifact.mdcli-gameplay-context-raw.md"
---

# CLI Gameplay Context

## Summary

The CLI gameplay use case should turn the current seed into a complete local
two-player tic-tac-toe game.

## Required Behavior

- Run from the terminal with `tictactoe`.
- Display a 3x3 board with clear available cells.
- Accept moves using positions `1` through `9`.
- Start with player `X` and alternate with player `O`.
- Reject invalid input, out-of-range positions, and occupied cells.
- Keep the same player on turn after an invalid move.
- Detect all row, column, and diagonal wins.
- Detect a draw when the board is full with no winner.
- Print the final board and result before exiting.

## Constraints

- Use only the Go standard library.
- Keep game rules separate from process I/O.
- Do not add colors, terminal control packages, persistence, networking, player
  profiles, or an AI opponent.

## Test Expectations

- Use table-driven tests for rules, move validation, wins, and draws.
- Use Go fuzz tests for move input parsing if the implementation keeps parsing
  as a separate boundary.
- Use `testing/quick` only for simple natural invariants.
- Do not add mutation testing tools or third-party assertions.
