# Slice 4: Review Evidence Gate

Status: delivered
Slicing: Agent Test/Evidence Strategy Pass
Plan: `ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md`
Tracker: `ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md`
Branch: `feat/slice-4-review-evidence-gate`
PR: `#323`

## Purpose

Prevent successful worker completion or direct status edits from moving a
ticket into review when required evidence is missing, malformed, blocked, or
failing.

This slice enforces the existing evidence strategy defaults:

- `enabled: true`
- `ticket_checklist: required`
- `strict_ready_for_review: true`

Projects can opt out through the existing evidence strategy configuration.

## Delivered Behavior

- Added an evidence readiness evaluator with stable diagnostic codes.
- Result application now evaluates readiness before setting a ticket to
  `reviewing`.
- Successful execution with missing or failing evidence now leaves the ticket
  `blocked` and writes a blocker reason.
- Direct `chr ticket status <key> ready_for_review` now uses the same evidence
  gate.
- Direct status changes rejected by the evidence gate leave the ticket
  unchanged.
- Existing branch and pull request reference requirements still apply.
- Optional evidence strategy configuration preserves the previous transition
  behavior.
- Loop and work-run test fixtures now include complete evidence checklists when
  they model successful worker tickets.
- Reference docs now describe review-transition evidence requirements,
  diagnostic codes, and opt-out settings.

## Diagnostics

The evaluator can classify these blockers:

- `evidence_strategy_missing`
- `evidence_strategy_malformed`
- `evidence_strategy_blocked`
- `gherkin_required_missing`
- `tests_required_missing`
- `validation_missing`
- `validation_failed`
- `runtime_unavailable`
- `codex_runtime_unavailable`
- `evidence_summary_missing`

## Implementation Notes

- The ticket checklist remains the primary operational source.
- Execution result application additionally requires a complete result
  `evidence_summary`.
- Runtime or validation failures are classified through existing execution
  failure kinds where possible.
- Direct status changes do not write blocker sections because rejected commands
  do not mutate the ticket.
- The gate intentionally checks only `ready_for_review`; it does not introduce
  a complete status state machine.

## Files of Interest

- `internal/core/evidence_readiness.go`
- `internal/core/evidence_readiness_test.go`
- `internal/core/result_application.go`
- `internal/core/result_application_test.go`
- `internal/app/ticket_evidence_readiness.go`
- `internal/app/ticket.go`
- `internal/app/service_artifact_ticket.go`
- `internal/app/cli_test.go`
- `docs/reference/evidence-strategy-config.md`
- `docs/reference/ticket-review-workflow.md`
- `docs/reference/ticket-status.md`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/core -run 'ResultApplication|EvidenceReadiness|TicketStatusUpdater'` (pass)
- `go test ./internal/app -run 'TestTicket(ApplyResult|ReviewRef|Status)'` (pass)
- `go test ./internal/app -run 'TestTicketLoopExecutesEligibleTickets|TestTicketLoopHonorsMaxIterations|TestWorkRunAllExecutesReadyTickets|TestTicketLoopDispatchesConfiguredFinishedNotification|TestServiceStartAgentWorkTicketReusesTicketExecution|TestServiceStartAllReadyAgentWork'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Gherkin applicability remains based on existing checklist state; deeper
  policy inference is reserved for the Gherkin test policy slice.
- Evidence status is enforced but not yet surfaced in Board detail; visibility
  work remains in the evidence visibility docs slice.
