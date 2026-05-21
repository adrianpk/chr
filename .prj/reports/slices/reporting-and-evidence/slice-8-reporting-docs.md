# Slice 8: Reporting Documentation

Status: delivered
Slicing: Reporting and Evidence
Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
Branch: `feat/slice-8-reporting-docs`
PR: #125

## Purpose

Make reporting behavior discoverable without reading implementation code.

This slice completes the user-facing documentation layer for execution reports,
loop reports, report commands, and review handoff summaries.

## Delivered Behavior

- Added a how-to guide for finding execution and loop evidence.
- Linked reporting docs from the documentation index.
- Added report command examples to the user guide.
- Linked ticket execution documentation to reporting references.
- Linked ticket loop documentation to reporting references.

## Implementation Notes

This slice is documentation-only. It does not change runtime behavior, report
formats, CLI behavior, or persistence.

The new how-to guide keeps operational steps separate from exact report format
contracts, which remain under `docs/reference/`.

## Contracts Added or Changed

- Added `docs/how-to/find-execution-and-loop-evidence.md`.
- Updated `docs/index.md`.
- Updated `docs/how-to/user-guide/index.md`.
- Updated `docs/reference/ticket-execution.md`.
- Updated `docs/reference/ticket-loop.md`.

## Files of Interest

- `docs/how-to/find-execution-and-loop-evidence.md`
- `docs/how-to/user-guide/index.md`
- `docs/index.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-loop.md`

## Validation

```sh
make verify
```

All commands pass.

## Risks and Follow-ups

- This slice does not add new command behavior.
- Future reporting commands should update both reference documentation and the
  evidence lookup how-to.
