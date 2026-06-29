# Slice 7: Review Handoff Summary

Status: delivered
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-7-review-handoff-summary`
PR: #124

## Purpose

Add a compact review handoff summary derived from existing execution evidence.

This slice gives later user-facing surfaces a stable way to show what happened,
what changed, whether validation passed, and what the reviewer should inspect
next.

## Delivered Behavior

- Added `ReviewHandoffSummary`.
- Added `NewReviewHandoffSummary`.
- Added deterministic Markdown rendering.
- Included status, result summary, validation, changed files, blocker or failure
  reason, artifact paths, and next action.
- Added tests for succeeded, failed, and blocked executions.
- Documented handoff summary semantics.

## Implementation Notes

The handoff summary is derived from `ExecutionReport`. It does not read ticket
files, result JSON, validation logs, runtime logs, or Git state.

The summary is intentionally concise. Full execution details remain in the
execution report and source artifacts.

## Contracts Added or Changed

- Added `ReviewHandoffSummary`.
- Added `RenderReviewHandoffSummaryMarkdown`.
- Added `docs/reference/review-handoff-summaries.md`.

## Files of Interest

- `internal/chr/review_handoff_summary.go`
- `internal/chr/review_handoff_summary_test.go`
- `docs/reference/review-handoff-summaries.md`
- `docs/reference/reports.md`

## Validation

```sh
go test ./internal/chr -run 'TestReviewHandoffSummary|TestNewExecutionReport|TestRenderExecutionReport'
make verify
```

All commands pass.

## Risks and Follow-ups

- Handoff summaries are currently a rendering contract, not a persisted report
  artifact.
- Reporting documentation remains deferred to Slice 8.
