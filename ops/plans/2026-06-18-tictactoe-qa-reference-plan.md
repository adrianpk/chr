# Tictactoe QA Reference Plan

Status: planned

Active Slice: Slice 1 - tictactoe qa reference

Tracker: `ops/trackers/2026-06-18-tictactoe-qa-reference-tracker.md`

## Purpose

Add a resettable QA reference project that exercises Chryso's full
generation and agent-work loop against a small but realistic CLI application.

The reference project is a Go standard-library tic-tac-toe CLI. It should give
the user enough preseeded context to generate specs, plans, tracking, and
tickets without answering a long quiz every reset, while still leaving the
implementation work for generated agent tickets.

## Delivery Rules

- One branch and one PR for the slice.
- PR targets `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- The slice must update its implementation report under
  `.prj/reports/slices/tictactoe-qa-reference/`.
- Each task gets one conventional commit.
- After the merged slice, return to `dev` before follow-up work.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | tictactoe qa reference | `feat/slice-1-tictactoe-qa-reference` | `feat(slice-1): tictactoe qa reference` | `.prj/reports/slices/tictactoe-qa-reference/slice-1-tictactoe-qa-reference.md` |

## Slice 1: Tictactoe QA Reference

Goal: add a new resettable QA reference project under `qa/ref/003/tictactoe`
for validating end-to-end Chryso generation and agent execution.

Behavior:

- `make qa-reset QA_CASE=tictactoe QA_REF=003` creates
  `qa/project/tictactoe` as an ignored active project with its own Git repo.
- The reference project is a small Go module using only the standard library.
- The active project can run `go test ./...` before generated work starts.
- The project includes enough `.chr/default` context for spec and plan generation
  without requiring repeated manual quiz answers.
- The primary use case does not include prewritten generated specs, plans,
  tracking, or tickets, so Chryso can create them during the exercise.
- The generated agent-work tickets should be able to run through
  `chr work run-all --use-case cli-gameplay --order key`.

Tasks:

- T1.1 Add the tictactoe Go reference project with a compilable incomplete CLI.
- T1.2 Seed Chryso project, use-case, dump, and quiz artifacts for the
  CLI gameplay workflow.
- T1.3 Add reset/test/smoke wiring and documentation for the tictactoe QA case.
- T1.4 Validate the resettable scenario and add the slice implementation report.
