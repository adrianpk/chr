# Slice 2: Execution Report Artifact Writer

Status: reviewing
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-2-execution-report-writer`
PR: `#119`

## Purpose

Persist execution reports as durable project-local artifacts.

Slice 1 defined the model and renderer. This slice adds the artifact writer and
connects it to result application.

## Delivered Behavior

- Added `ExecutionReportPath` for deterministic report placement.
- Added `ExecutionReportWriter`.
- Writes `execution-report.md` under the execution run directory.
- Creates the report directory when needed.
- Refuses to overwrite an existing non-report file.
- Writes execution reports from `chc ticket apply-result`.
- Keeps dry-run result application side-effect free.

## Implementation Notes

The writer derives the target path from the execution record run directory. It
does not accept arbitrary caller paths.

Existing execution reports may be regenerated. Existing unrelated files at the
same path are rejected.

## Contracts Added or Changed

- Execution report artifact path: `<run_dir>/execution-report.md`.
- `chc ticket apply-result` now writes both `result-application.md` and
  `execution-report.md`.
- Dry-run result application does not write either report artifact.

## Files of Interest

- `internal/chc/execution_report.go`
- `internal/chc/execution_report_test.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/execution-reports.md`
- `docs/reference/ticket-result-application.md`

## Validation

```sh
go test ./internal/chc -run 'Test(ExecutionReportWriter|NewExecutionReport|RenderExecutionReport)'
go test ./internal/chcapp -run 'TestTicketApplyResult'
make verify
```

All commands pass.

## Risks and Follow-ups

- Report listing and lookup remain deferred to later slices.
- Loop report artifacts remain deferred to Slice 4.
