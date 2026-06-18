---
title: "CLI Gameplay Execution Tracking"
project_id: "project-7c98c4f4cf644493"
case_slug: "cli-gameplay"
status: "tracking"
plan_path: ".chc/user/plans/cli-gameplay-plan.md"
created_at: "2026-06-18T08:51:15Z"
---

# CLI Gameplay Execution Tracking

## Execution Index

Initial status: planned; no implementation work has started.

Workspace decision discovery found no ADR, RFC, decision-log, engineering-principles, or equivalent accepted/active decision documents. The binding constraints are `AGENTS.md`, `.agents/CONVENTIONS.md`, the CLI gameplay plan, curated ChrysoCoord context, and quiz answers.

## Ordered Slices

| Order | Slice | Ticket | Branch | Depends on | Initial status |
| --- | --- | --- | --- | --- | --- |
| 1 | `S01` Game outcome rules | `cli-gameplay-s01-game-outcomes` | `feature/cli-gameplay/s01-game-outcomes` | None | Not started |
| 2 | `S02` Move state transitions | `cli-gameplay-s02-move-state` | `feature/cli-gameplay/s02-move-state` | `S01` | Not started |
| 3 | `S03` CLI board rendering | `cli-gameplay-s03-cli-rendering` | `feature/cli-gameplay/s03-cli-rendering` | `S02` | Not started |
| 4 | `S04` Interactive gameplay loop | `cli-gameplay-s04-gameplay-loop` | `feature/cli-gameplay/s04-gameplay-loop` | `S03` | Not started |
| 5 | `S05` Test coverage hardening | `cli-gameplay-s05-test-hardening` | `feature/cli-gameplay/s05-test-hardening` | `S04` | Not started |

## Recommended Ticket Order

1. `cli-gameplay-s01-game-outcomes`
2. `cli-gameplay-s02-move-state`
3. `cli-gameplay-s03-cli-rendering`
4. `cli-gameplay-s04-gameplay-loop`
5. `cli-gameplay-s05-test-hardening`

Execution should stay sequential. Create each slice branch from latest `dev` after the prior slice merges.

## Expected Branch and PR References

Base and target branch for every slice: `dev`.

Each slice should produce one focused PR into `dev`. Domain PRs should avoid CLI text churn. CLI PRs should avoid domain semantic changes unless required by the selected ticket. Every PR should include summary, validation commands, and intentionally deferred behavior.

## Global Constraints

- Use only Go's standard library.
- Keep game rules in `internal/game`.
- Keep argument handling, prompts, rendering, stdin/stdout, and process behavior in `cmd/tictactoe`.
- Reject invalid text, out-of-range positions, and occupied cells without advancing the turn.
- Do not add replay, AI opponents, persistence, networking, colors, player profiles, terminal UI packages, or third-party test helpers.
- Prefer table-driven tests for rules, parsing, validation, wins, draws, and state transitions.
- Add fuzz or property-style tests only where the plan explicitly justifies them.

## Completion Gate

The use case is complete when `go test ./...` passes, `tictactoe` runs a full local two-player stdin/stdout game, `X` starts, turns advance only after valid moves, all wins and draws are detected, the final board and result are printed, help behavior still works, and scripted CLI flows cover win, draw, invalid-input recovery, occupied-cell recovery, and help.

## Tickets

- S01: CHC-1 - Game outcome rules (`.chc/user/tickets/cli-gameplay/CHC-1-Game-outcome-rules.md`)
- S02: CHC-2 - Move state transitions (`.chc/user/tickets/cli-gameplay/CHC-2-Move-state-transitions.md`)
- S03: CHC-3 - CLI board rendering (`.chc/user/tickets/cli-gameplay/CHC-3-CLI-board-rendering.md`)
- S04: CHC-4 - Interactive gameplay loop (`.chc/user/tickets/cli-gameplay/CHC-4-Interactive-gameplay-loop.md`)
- S05: CHC-5 - Test coverage hardening (`.chc/user/tickets/cli-gameplay/CHC-5-Test-coverage-hardening.md`)

## Provenance

- plan: .chc/user/plans/cli-gameplay-plan.md
- source: project_context
- source: request_context
- source: quiz_context
- source: request_context
- source: .chc/user/plans/cli-gameplay-plan.md
