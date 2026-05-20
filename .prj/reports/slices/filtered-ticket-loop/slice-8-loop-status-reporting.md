# Slice 8: Loop Status Reporting

Status: in review
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-8-loop-status-reporting`
PR: #63

## Purpose

Expose foreground loop outcomes and iteration evidence directly in command
output.

## Delivered Behavior

`chc ticket loop` now prints a complete loop execution summary for successful,
blocked, and failed loop runs.

The summary includes:

- loop run ID;
- loop run status;
- eligible ticket count;
- processed ticket count;
- skipped ticket count;
- blocked ticket count;
- failed ticket count;
- execution IDs for attempted tickets.

## Implementation Notes

The runner preserves the total deduplicated eligible count before applying
`--max-iterations`, then computes skipped tickets from the loop run counters.

Execution failures now return an error after printing the failed loop summary,
so the CLI remains scriptable while still exposing the persisted loop outcome.

No separate loop report artifact was added. The persisted `LoopRunRecord` and
per-ticket `ExecutionRecord` values remain the lookup boundary for this slice.

## Contracts Added or Changed

Changed CLI output:

- loop execution output includes `skipped`;
- failed loop execution prints the loop summary before returning the failure.

## Files of Interest

- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/ticket-loop.md`

## Validation

- `go test ./cmd/chc -run 'TestTicketLoop'`: pass
- `make verify`: pass

## Risks and Follow-ups

This completes the planned foreground loop reporting surface for the current
slicing. Richer historical lookup commands remain outside this slice.
