# Slice 3: Validation and Changed-File Summaries

Status: delivered
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-3-evidence-summaries`
PR: `#120`

## Purpose

Normalize validation and changed-file evidence for execution reports.

This slice makes report output more useful and bounded without reading full log
contents or duplicating existing evidence artifacts.

## Delivered Behavior

- Added validation summary fields to execution reports.
- Derived validation status from execution status and validation artifact
  presence.
- Sorted changed files by path before rendering.
- Added changed-file summary counts.
- Bounded rendered changed-file entries to 25.
- Marked changed-file summaries as truncated when input exceeds the render
  limit.

## Implementation Notes

Validation status is intentionally inferred from existing execution metadata.
The report does not inspect validation log contents.

Changed-file lists remain deterministic by sorting paths before applying the
render limit.

## Contracts Added or Changed

- Execution reports now include a `Validation` section.
- Changed-file sections now include `total`, `rendered`, and `truncated`
  summary fields.
- Execution reports render at most 25 changed-file entries.

## Files of Interest

- `internal/chc/execution_report.go`
- `internal/chc/execution_report_test.go`
- `docs/reference/execution-reports.md`

## Validation

```sh
go test ./internal/chc -run 'Test(NewExecutionReport|RenderExecutionReport|ExecutionReportWriter)'
make verify
```

Both commands pass.

## Risks and Follow-ups

- Validation status is metadata-derived; parsing validation logs remains out of
  scope.
- Report listing and lookup remain deferred to later slices.
