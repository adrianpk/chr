---
title: "CLI Gameplay Context Raw Capture"
status: "raw"
type: "dump"
case: "cli-gameplay"
curated_path: "ops/default/dump/<id>/artifact.mdcli-gameplay-context.md"
---

# CLI Gameplay Context Raw Capture

The actual feature I want to test is the game loop. A person runs the command,
sees a board, and enters moves. Use numbers 1 through 9 for cells because that
is easy to explain in a CLI and easy to test. The board should show which cells
are still available.

X starts. O follows. If someone types junk, an out-of-range number, or an
occupied cell, the CLI should say what is wrong and ask the same player again.
Bad input should not advance the turn.

The game ends when X or O gets three in a row, or when the board fills with no
winner. The app should print the final board and a clear result. It does not
need colors, screen clearing, player names, save files, or an AI opponent.

For tests, I care more about confidence than ceremony. Table tests for wins,
draws, and invalid moves are expected. Fuzzing makes sense for parsing move
input. A small property-style test is useful if it checks something natural,
like accepted moves always point inside the board. Do not add mutation testing
tools because this fixture should stay standard-library only.
