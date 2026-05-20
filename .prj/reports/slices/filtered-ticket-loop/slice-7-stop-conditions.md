# Slice 7: Stop Conditions and Thresholds

Status: in review
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-7-stop-conditions`
PR: #62

## Purpose

Stop foreground ticket loops predictably at completion, configured limits,
blocked results, or execution failures.

## Delivered Behavior

`chc ticket loop` now supports `--max-iterations`.

The command:

- completes successfully when no eligible tickets are selected;
- caps attempted tickets when `--max-iterations` is greater than zero;
- treats zero or omitted `--max-iterations` as unlimited;
- deduplicates ticket keys before execution;
- stops after the first blocked result and marks the loop run blocked;
- stops after the first execution failure and marks the loop run failed.

## Implementation Notes

Stop-condition handling stays in the CLI loop runner and reuses the existing
single-ticket execution and result application paths.

The selected run list is deduplicated before the max-iteration threshold is
applied, so one ticket key cannot be attempted twice in one run.

## Contracts Added or Changed

Changed CLI behavior:

- `ticket loop --max-iterations N` limits the number of eligible tickets
  attempted by one foreground loop run.

Changed loop run behavior:

- blocked result applications finish the loop with `blocked` status;
- execution failures finish the loop with `failed` status and increment the
  failed counter.

## Files of Interest

- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/ticket-loop.md`

## Validation

- `go test ./cmd/chc -run 'TestTicketLoop'`: pass
- `make verify`: pass

## Risks and Follow-ups

Loop status output remains minimal. Slice 8 expands loop status reporting.
