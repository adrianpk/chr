# Slice 1: Result Application Model

Status: delivered
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-1-result-application-model`
PR: `#41`

## Purpose

Define a side-effect free model for applying one execution result to one ticket.

## Delivered Behavior

The core package now exposes a result application model that combines an
execution record and execution result envelope into ticket-facing state.

The model maps:

- succeeded execution with succeeded result to ticket status `reviewing`;
- failed, blocked, or failed-result outcomes to ticket status `blocked`.

It carries the fields later writer slices need for comments, validation
evidence, changed-file summaries, and blocked reasons.

## Implementation Notes

The model validates that the execution record and result envelope refer to the
same execution ID and ticket key. It does not mutate ticket Markdown or refresh
the ticket index.

## Contracts Added or Changed

New core types:

- `ResultApplication`
- `ResultApplicationExecution`
- `ResultApplicationComment`
- `ResultApplicationEvidence`
- `ResultApplicationChangedFile`
- `ResultApplicationBlockedReason`
- `ResultApplicationOptions`

New function:

- `NewResultApplication`

## Files of Interest

- `internal/chr/result_application.go`
- `internal/chr/result_application_test.go`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

The model only prepares ticket-facing state. Markdown mutation starts in Slice 2.
