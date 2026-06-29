---
schema_version: 1
name: "Tictactoe"
description: "Small standard-library Go CLI game used as a Chryso QA reference project."
---

Tictactoe is a local command-line game for two human players sharing one
terminal. The seed project is intentionally small and incomplete so Chryso
can generate implementation work from project context.

The first useful version should run a complete tic-tac-toe game in the terminal:
show a 3x3 board, alternate X and O turns, accept moves from standard input,
reject invalid moves, detect wins and draws, and exit cleanly.

The project should stay standard-library only. Tests should grow with the
implementation and should cover game rules, input parsing, invalid moves, and
end states.
