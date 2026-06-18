---
project_id: "project-tictactoe-qa"
case_slug: "cli-gameplay"
scope: "cli-gameplay"
status: "raw"
runtime_id: "seed-cli-gameplay"
curated_path: ".chc/user/quizzes/cli-gameplay-quiz.md"
updated_at: "2026-06-18T07:52:42Z"
---

## Turn

- question_id: q1
- status: answered
- mode: deep
- topic: gameplay-scope
- depth: 1
- tags: cli, scope

Question:

What is the smallest useful tic-tac-toe version?

Answer:

A local two-player CLI game. One person runs the command and two humans share
the terminal. The app should show the board, ask for moves, alternate turns,
and end with a win or draw.

## Turn

- question_id: q2
- status: answered
- mode: deep
- topic: input-model
- depth: 1
- tags: cli, input

Question:

How should players choose cells?

Answer:

Use numbers 1 through 9. Invalid text, out-of-range numbers, and occupied cells
should be rejected without advancing the turn.

## Turn

- question_id: q3
- status: answered
- mode: deep
- topic: exclusions
- depth: 1
- tags: constraints

Question:

What should stay out of the first implementation?

Answer:

No AI opponent, no persistence, no networking, no terminal UI libraries, no
colors, no player profiles, and no dependencies outside Go's standard library.

## Turn

- question_id: q4
- status: answered
- mode: deep
- topic: testing
- depth: 1
- tags: testing

Question:

What kind of tests should the finished implementation have?

Answer:

Table-driven tests for rules and state transitions, fuzz tests for move input
parsing when useful, and a small property-style test only if there is a natural
invariant. Do not pull in mutation testing tools.
