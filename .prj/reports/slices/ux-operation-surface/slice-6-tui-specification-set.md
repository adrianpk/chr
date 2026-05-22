# Slice 6: TUI Specification Set

Status: delivered

Plan: `.prj/plans/slicing-ux-operation-surface-plan.md`
Tracker: `.prj/trackers/slicing-ux-operation-surface-tracker.md`
Ticket: `.prj/tickets/open/20260522071859-align-cli-authoring-and-tui-operation-surface.md`

## Summary

Created the first granular specification set for the Bubble Tea TUI.

The specs define the TUI as the local human operation surface over the existing
CLI/domain foundation. They cover product role, navigation, record screens,
authoring and quiz workflows, ticket execution evidence, action safety, and
implementation boundaries.

## Created Specs

- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-navigation-and-layout.md`
- `.prj/specs/specs/tui-record-screens.md`
- `.prj/specs/specs/tui-authoring-and-quiz-workflow.md`
- `.prj/specs/specs/tui-ticket-execution-and-evidence.md`
- `.prj/specs/specs/tui-actions-and-safety.md`
- `.prj/specs/specs/tui-implementation-boundary.md`

## Notes

- The specs keep the CLI as the executable foundation.
- The TUI is defined as the ergonomic local human workflow surface.
- Static quiz generation is treated as a fallback/template mode.
- Context-aware quiz generation is specified as a separate product capability.
- The first TUI implementation is intentionally read-only.
- Mutating and runtime actions require explicit confirmation and evidence.

## Validation

- Documentation diff check passed before commit.
- No runtime code was changed in this slice.

