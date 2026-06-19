# Slice 6: Evidence Visibility Docs

Status: delivered
Slicing: Agent Test/Evidence Strategy Pass
Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`
Tracker: `ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md`
Branch: `feat/slice-6-evidence-visibility-docs`
PR: `#325`

## Purpose

Make the evidence strategy visible from normal ticket and work-observation
surfaces without making runtime logs the only way to understand what happened.

This slice does not add new review policy. It exposes concise status derived
from the ticket evidence checklist and documents how to find the detailed
runtime evidence.

## Delivered Behavior

- Added a core evidence summary model over the anchored ticket checklist.
- Board selected-ticket detail now shows overall evidence state plus strategy,
  Gherkin, tests, validation, and evidence-summary item states.
- Dash work-action details now surface ticket, execution report, runtime log,
  validation log, and `chc work logs` lookup references when event refs exist.
- `chc ticket show` prints detailed evidence fields when a ticket has evidence
  state.
- `chc ticket list` and `chc ticket board` append a compact evidence status
  cell while preserving normal Board columns.
- User-facing documentation now describes the visibility surfaces, defaults,
  config overrides, and review handoff expectations.
- The tic-tac-toe QA reference now includes resettable evidence examples for
  docs-only no-test work, CLI behavior with active Gherkin, and blocked
  validation.

## Implementation Notes

- Evidence visibility is derived from ticket Markdown at render time. No SQLite
  migration was needed.
- Missing evidence is explicit in Board detail and compact as `-` in list/board
  CLI rows.
- Malformed evidence blocks surface parse errors in detail views.
- QA examples live outside the initial ticket directory, so the reference
  scenario remains free to generate its own tickets.

## Contracts Added or Changed

- Added `core.EvidenceStrategySummary` and
  `core.SummarizeEvidenceStrategyChecklist`.
- Extended TUI `TicketView` with an evidence summary derived from the ticket
  Markdown file.
- Extended TUI event views with ticket path and execution report refs.
- Extended ticket CLI output with evidence status fields.
- Extended tictactoe QA smoke checks to verify the resettable evidence fixtures.

## Files of Interest

- `internal/core/evidence_strategy_summary.go`
- `internal/tui/model.go`
- `internal/tui/work_actions.go`
- `internal/tui/observability.go`
- `internal/app/ticket.go`
- `docs/reference/evidence-strategy-config.md`
- `docs/reference/commands.md`
- `docs/how-to/user-guide/index.md`
- `qa/ref/003/tictactoe/.chc/user/evidence-examples/`
- `scripts/qa/smoke.sh`

## Validation

- `go test ./internal/core -run 'SummarizeEvidenceStrategyChecklist|NewEvidenceStrategySummary'` (pass)
- `go test ./internal/tui -run 'TestModelRendersTicketEvidenceDetail|TestModelRendersWorkActionEventStream'` (pass)
- `go test ./internal/app -run 'TestTicketShowPrintsEvidenceStrategySummary|TestTicketListPrintsIndexedTickets|TestTicketBoardPrintsBoardColumns'` (pass)
- `make qa-smoke-tictactoe` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Evidence summary is derived per rendered ticket. If a later view needs bulk
  evidence filtering, that should become a dedicated indexed read model rather
  than another ad hoc scan.
