# Slice 7: Dump Editing UX Polish

Status: delivered
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-7-dump-editing-ux-polish`
PR: #181

## Purpose

Polish the dump editing surface after full-document Markdown editing became
available in the TUI.

## Delivered Behavior

- The right-side dump panel is titled `Preview` in read mode and `Edit Dump` in
  edit mode.
- Internal Markdown editing uses panel-derived width and height instead of a
  fixed editor size.
- Edit mode displays save, cancel, and external editor hints inside the detail
  panel.
- The selected dump remains stable after save, cancel, and refresh flows.

## Implementation Notes

- Dump detail rendering now returns a panel title plus panel lines so shell
  layout can switch labels without parsing rendered text.
- The editor height is derived from the available detail panel height after
  metadata and hint lines are accounted for.
- Legacy non-shell dump rendering keeps the `Selected Dump` heading.

## Contracts Added or Changed

- No service contracts changed.
- TUI dump detail rendering now accepts panel dimensions.

## Files of Interest

- `internal/chrtui/dumps.go`
- `internal/chrtui/shell.go`
- `internal/chrtui/model_test.go`

## Validation

- `go test ./internal/chrtui`
- `go test ./internal/chr ./internal/chrapp ./internal/chrtui`
- `make verify`

## Risks and Follow-ups

- The editor remains intentionally simple and does not provide rich Markdown
  syntax highlighting.
