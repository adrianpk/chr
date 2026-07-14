# Slice 1: Tictactoe QA Reference

Status: delivered
Slicing: Tictactoe QA Reference
Plan: `ops/plans/2026-06-18-tictactoe-qa-reference-plan.md`
Tracker: `ops/trackers/2026-06-18-tictactoe-qa-reference-tracker.md`
Branch: `feat/slice-1-tictactoe-qa-reference`
PR: #306

## Purpose

Add a resettable tic-tac-toe QA reference project for validating Chryso's
generation and Codex-backed agent-work loop against a small CLI application.

## Delivered Behavior

The repository now contains `qa/ref/003/tictactoe`, a Go standard-library CLI
seed project. The seed is intentionally incomplete but compilable: it exposes a
placeholder command, a small `internal/game` package, and baseline tests for
move parsing and placement.

The reference includes project-local Chryso context under `.chr/default`:
project metadata, one `cli-gameplay` use case, raw and curated project dumps,
raw and curated use-case dumps, and raw and curated quiz notes. It does not
include generated specs, plans, tracking documents, or tickets for the primary
gameplay flow.

QA reset/test/smoke aliases were added for the tictactoe case. The generic
smoke script now preserves the existing blog failing-baseline expectation while
allowing passing-baseline reference projects such as tictactoe.

## Implementation Notes

The fixture is designed for manual TUI or CLI exercise:

```sh
make qa-reset-tictactoe
cd qa/project/tictactoe
chr spec generate --use-case cli-gameplay
chr plan generate --use-case cli-gameplay
chr plan tracking create --use-case cli-gameplay
chr work run-all --use-case cli-gameplay --order key
```

An empty non-Markdown `.chr/default/tickets/.gitkeep` preserves the tickets
directory without creating artificial tickets.

## Contracts Added or Changed

- `qa/ref/003/tictactoe` is the versioned reference source.
- `qa/project/tictactoe` is the ignored active reset target.
- `make qa-reset-tictactoe`, `make qa-test-tictactoe`, and
  `make qa-smoke-tictactoe` are supported.
- `scripts/qa/smoke.sh` supports both failing-baseline and passing-baseline QA
  references.

## Files of Interest

- `qa/ref/003/tictactoe/`
- `Makefile`
- `scripts/qa/smoke.sh`
- `docs/how-to/qa-tictactoe-agent-work-scenario.md`
- `docs/reference/commands.md`

## Validation

- `go -C qa/ref/003/tictactoe test ./...` (pass)
- `make qa-smoke-tictactoe` (pass)
- `make qa-smoke-blog` (pass)
- `go run ./cmd/chr spec list --dir qa/project/tictactoe` (pass)
- `go run ./cmd/chr plan list --dir qa/project/tictactoe` (pass)
- `go run ./cmd/chr ticket list --dir qa/project/tictactoe` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The scenario has not run model-backed spec or plan generation in this slice;
  that is the intended manual evaluation workflow.
- Generated ticket quality depends on the active model/backend configuration.
