# Slice 5: Report Listing

Status: delivered
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-5-report-listing`
PR: #122

## Purpose

Expose report discovery through the CLI.

This slice lets users find existing execution and loop report artifacts without
knowing their paths.

## Delivered Behavior

- Added core report artifact discovery under `.chr/runtime/runs/`.
- Added report metadata extraction from Markdown list fields.
- Added filters for kind, ticket key, execution id, loop id, and status.
- Added `chr report` command group.
- Added `chr report list`.
- Added stable line-oriented output for report entries.

## Implementation Notes

The lister reads only lightweight Markdown metadata. It does not inspect runtime
logs, validation output, execution result JSON, or databases.

The command discovers `execution-report.md` and `loop-report.md` files. Direct
body rendering remains deferred to `report show`.

## Contracts Added or Changed

- Added `ReportKind`.
- Added `ReportListFilter`.
- Added `ReportListEntry`.
- Added `ReportLister`.
- Added `chr report list`.
- Added `docs/reference/reports.md`.

## Files of Interest

- `internal/chr/report_listing.go`
- `internal/chr/report_listing_test.go`
- `internal/chrapp/report.go`
- `internal/chrapp/app.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/reports.md`

## Validation

```sh
go test ./internal/chr -run 'TestReportLister'
go test ./internal/chrapp -run 'TestReportList'
make verify
```

All commands pass.

## Risks and Follow-ups

- `report show` remains deferred to Slice 6.
- Listing depends on report Markdown metadata staying compatible with the
  current report renderers.
