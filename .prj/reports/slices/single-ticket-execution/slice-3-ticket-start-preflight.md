# Slice 3: Ticket Start Preflight

Status: delivered
Slicing: Single Ticket Execution
Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
Branch: `feat/slice-3-ticket-start-preflight`
PR: `#36`

## Purpose

Prepare one ticket execution from the CLI without invoking an agent harness.

## Delivered Behavior

`chr ticket start <ticket-key>` resolves a project, reconciles the ticket index,
selects one indexed ticket, validates the runtime boundary, and creates a
`pending` execution record for that ticket.

`--dry-run` prints the same preflight and runtime validation output without
creating an execution record.

## Implementation Notes

The command uses the existing project resolver and ticket reconciler before
ticket lookup. Runtime validation uses the generated execution ID as the runtime
run ID, which scopes result, log, and validation artifact paths to the execution.

Runtime validation failures print the same preflight output and checks, then exit
with an error before persistence.

## Contracts Added or Changed

New command:

- `chr ticket start <ticket-key>`

New CLI output fields:

- `execution`
- `ticket`
- `ticket_path`
- `ticket_relative_path`
- `status`

The command also prints the existing runtime validation plan and check output.

## Files of Interest

- `cmd/chr/main.go`
- `cmd/chr/main_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./cmd/chr ./internal/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice only prepares execution metadata. Harness command construction and
agent invocation remain in later slices.
