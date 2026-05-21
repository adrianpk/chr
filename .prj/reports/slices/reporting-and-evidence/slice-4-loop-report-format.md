# Slice 4: Loop Report Format

Status: reviewing
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-4-loop-report-format`
PR: pending

## Purpose

Define the loop report model and Markdown rendering contract.

This slice gives later report persistence, listing, and lookup work a stable
loop report shape.

## Delivered Behavior

- Added `LoopReport` model.
- Added optional related execution summaries.
- Added deterministic Markdown rendering for loop summary, counts, timings,
  related executions, and failure reason.
- Added bounded execution rendering with total, rendered, and truncated counts.
- Added focused tests for empty, normal, and truncated execution lists.

## Implementation Notes

The loop report renderer is side-effect free. It does not write artifacts and
does not query execution records by itself.

Related executions are caller-provided so later slices can decide how to
resolve them from persisted loop/execution state.

## Contracts Added or Changed

- Added `LoopReport`.
- Added `LoopReportOptions`.
- Added `NewLoopReport`.
- Added `RenderLoopReportMarkdown`.
- Added `docs/reference/loop-reports.md`.

## Files of Interest

- `internal/chc/loop_report.go`
- `internal/chc/loop_report_test.go`
- `docs/reference/loop-reports.md`
- `docs/reference/ticket-loop.md`

## Validation

```sh
go test ./internal/chc -run 'Test(NewLoopReport|RenderLoopReport)'
make verify
```

Both commands pass.

## Risks and Follow-ups

- Loop report persistence is intentionally deferred.
- CLI listing and lookup remain deferred to later slices.
