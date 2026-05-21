# Slice 1: Execution Report Format

Status: reviewing
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-1-execution-report-format`
PR: `#118`

## Purpose

Define the first stable reporting contract for ticket execution evidence.

This slice gives later writer, listing, and lookup slices a side-effect free
execution report model and deterministic Markdown renderer.

## Delivered Behavior

- Added an `ExecutionReport` model derived from an `ExecutionRecord`.
- Added optional structured result inclusion from `ExecutionResult`.
- Added changed-file summary inclusion from execution results.
- Added deterministic Markdown rendering for summary, timings, runtime,
  artifacts, result, changed files, and failure sections.
- Added validation for mismatched execution IDs and ticket keys.

## Implementation Notes

The renderer does not write files and does not read artifact contents. It only
summarizes already-known evidence.

Missing optional values render as `none`, and missing result or changed-file
evidence is represented explicitly in the Markdown body.

## Contracts Added or Changed

- Added `ExecutionReport`.
- Added `ExecutionReportOptions`.
- Added `NewExecutionReport`.
- Added `RenderExecutionReportMarkdown`.
- Added `docs/reference/execution-reports.md`.

## Files of Interest

- `internal/chc/execution_report.go`
- `internal/chc/execution_report_test.go`
- `docs/reference/execution-reports.md`

## Validation

```sh
go test ./internal/chc -run 'Test(NewExecutionReport|RenderExecutionReport)'
make verify
```

Both commands pass.

## Risks and Follow-ups

- Report persistence is intentionally deferred to Slice 2.
- CLI listing and lookup are intentionally deferred to later slices.
- Loop reports are intentionally deferred to Slice 4.
