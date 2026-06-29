# Slice 6: Report Lookup

Status: delivered
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-6-report-lookup`
PR: #123

## Purpose

Expose direct report lookup through the CLI.

This slice lets users inspect one execution or loop report without manually
opening the report path.

## Delivered Behavior

- Added report lookup by report id, ticket key, execution id, loop id, kind, or
  status.
- Added `chr report show`.
- Added Markdown body output for selected reports.
- Added metadata output with `--metadata`.
- Added explicit errors for missing and ambiguous lookups.

## Implementation Notes

Report lookup reuses the report discovery boundary from `chr report list`.
Lookup reads the selected Markdown report artifact and prints it without
rewriting or re-rendering the report.

A report id can be a project-relative report path, an absolute report path, a
ticket key, an execution id, or a loop id. Broad lookups must resolve to exactly
one report.

## Contracts Added or Changed

- Added `ReportLookupRequest`.
- Added `ReportLookupResult`.
- Added `ReportLookupService`.
- Added `chr report show`.
- Added `chr report show --metadata`.
- Updated `docs/reference/reports.md`.

## Files of Interest

- `internal/chr/report_lookup.go`
- `internal/chr/report_lookup_test.go`
- `internal/chrapp/report.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/reports.md`

## Validation

```sh
make verify
```

All commands pass.

## Risks and Follow-ups

- Broad ticket-key lookups can be ambiguous when several executions share the
  same ticket. This is intentional and requires the user to narrow the lookup.
- Review handoff summaries remain deferred to Slice 7.
