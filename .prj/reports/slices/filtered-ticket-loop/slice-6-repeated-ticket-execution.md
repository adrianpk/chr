# Slice 6: Repeated Ticket Execution

Status: delivered
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-6-repeated-ticket-execution`
PR: #61

## Purpose

Run eligible loop tickets sequentially by reusing the existing single-ticket
execution and result application paths.

## Delivered Behavior

`chc ticket loop` can now run without `--dry-run`.

The command:

- resolves project context;
- refreshes the ticket index;
- queries eligible tickets;
- creates a loop run record;
- executes each eligible ticket through the existing ticket-start execution
  path;
- applies each execution result through the existing result application path;
- updates loop run counters;
- prints a loop execution summary and per-ticket execution IDs.

## Implementation Notes

The implementation keeps single-ticket behavior centralized by extracting
helpers that build ticket-start preflight for a known ticket and build result
application plans for a known execution record.

This slice intentionally does not add configurable stop conditions or
thresholds. That remains Slice 7.

## Contracts Added or Changed

Changed CLI behavior:

- `ticket loop` no longer requires `--dry-run`.

Changed internals:

- loop execution creates and updates `LoopRunRecord` values;
- loop execution stores per-ticket `ExecutionRecord` values through the
  existing execution repository.

## Files of Interest

- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/ticket-loop.md`

## Validation

- `go test ./cmd/chc -run 'TestTicketLoop'`: pass
- `make verify`: pass

## Risks and Follow-ups

Loop execution currently processes the selected eligible set without configurable
thresholds. Slice 7 adds stop conditions and retry or iteration limits.
