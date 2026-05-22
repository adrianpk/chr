# TUI Operational Console Slicing Tracker

Status: planned

Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Spec: `.prj/specs/specs/tui-operational-console.md`

## Slicing Checklist

- [x] Capture TUI design inputs.
- [x] Create TUI operational console spec.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: TUI App Foundation | Delivered | `feat/slice-1-tui-app-foundation` | `feat(slice-1): tui app foundation` | `.prj/reports/slices/tui-operational-console/slice-1-tui-app-foundation.md` | PR #139 merged. |
| Slice 2: Read-Only Board And Ticket Detail | Delivered | `feat/slice-2-tui-board-ticket-detail` | `feat(slice-2): tui board ticket detail` | `.prj/reports/slices/tui-operational-console/slice-2-tui-board-ticket-detail.md` | PR #140 merged. |
| Slice 3: Document And Authoring Explorer | Delivered | `feat/slice-3-tui-document-authoring-explorer` | `feat(slice-3): tui document authoring explorer` | `.prj/reports/slices/tui-operational-console/slice-3-tui-document-authoring-explorer.md` | PR #141 merged. |
| Slice 4: Agent And Runtime Observability | Planned | `feat/slice-4-tui-agent-runtime-observability` | `feat(slice-4): tui agent runtime observability` | `.prj/reports/slices/tui-operational-console/slice-4-tui-agent-runtime-observability.md` | Read-only agent/runtime/evidence state. |
| Slice 5: Command Palette Foundation | Planned | `feat/slice-5-tui-command-palette` | `feat(slice-5): tui command palette` | `.prj/reports/slices/tui-operational-console/slice-5-tui-command-palette.md` | Discoverable commands. |
| Slice 6: Safe Artifact Creation Actions | Planned | `feat/slice-6-tui-artifact-creation-actions` | `feat(slice-6): tui artifact creation actions` | `.prj/reports/slices/tui-operational-console/slice-6-tui-artifact-creation-actions.md` | Low-risk creation flows. |
| Slice 7: Ticket Mutation Actions | Planned | `feat/slice-7-tui-ticket-mutation-actions` | `feat(slice-7): tui ticket mutation actions` | `.prj/reports/slices/tui-operational-console/slice-7-tui-ticket-mutation-actions.md` | Explicit ticket operations. |
| Slice 8: Agent Work Control | Planned | `feat/slice-8-tui-agent-work-control` | `feat(slice-8): tui agent work control` | `.prj/reports/slices/tui-operational-console/slice-8-tui-agent-work-control.md` | Start/stop selected work. |

## Slice 1: TUI App Foundation

Candidate tasks:

- [x] Add Bubble Tea dependencies.
- [x] Add `internal/chctui` package boundary.
- [x] Add `chc tui` command.
- [x] Dispatch `chc` without subcommands to the TUI.
- [x] Preserve `chc --help`.
- [x] Implement base model/update/view.
- [x] Render project-aware navigation shell.
- [x] Add tests for model/update behavior.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: Bubble Tea dependency and TUI package boundary.
- T1.2: `chc tui` and default `chc` routing.
- T1.3: base TUI model, navigation, and project shell.
- T1.4: tests, validation, and slice report.

## Slice 2: Read-Only Board And Ticket Detail

Candidate tasks:

- [x] Load tickets through existing repositories/indexes.
- [x] Group tickets by status.
- [x] Render board/list rows with headers.
- [x] Implement ticket selection.
- [x] Render selected ticket detail.
- [x] Show checklist and comment summaries where available.
- [x] Add text filtering.
- [x] Add tests for ticket view model construction.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `7edc2d8` loads project tickets for TUI startup.
- T2.2: `7edc2d8` groups tickets by status for the Board view.
- T2.3: `7edc2d8` renders board and list rows.
- T2.4: `7edc2d8` implements ticket selection.
- T2.5: `7edc2d8` renders selected ticket detail.
- T2.6: `7edc2d8` shows checklist and comment summaries.
- T2.7: `7edc2d8` adds text filtering.
- T2.8: `7edc2d8` adds ticket view tests.
- T2.9: `1d0e928` adds the slice implementation report.
- T2.10: `make verify` passed.

## Slice 3: Document And Authoring Explorer

Candidate tasks:

- [x] Discover project-local document artifacts.
- [x] Discover authoring sessions.
- [x] Render document list and detail views.
- [x] Render authoring session list and detail views.
- [x] Show plans/slicing/tracking as artifacts.
- [x] Preserve tickets as operational status source.
- [x] Add tests for discovery and view models.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `1be6d40` discovers project-local document artifacts.
- T3.2: `1be6d40` discovers authoring sessions.
- T3.3: `1be6d40` renders document list and detail views.
- T3.4: `1be6d40` renders authoring session list and detail views.
- T3.5: `1be6d40` shows plans, slicing plans, and trackers as artifacts.
- T3.6: `1be6d40` keeps tickets as the operational status source.
- T3.7: `1be6d40` adds discovery and view model tests.
- T3.8: `48a4ae8` adds the slice implementation report.
- T3.9: `make verify` passed.

## Slice 4: Agent And Runtime Observability

Candidate tasks:

- [ ] Load daemon status where available.
- [ ] Load loop and execution summaries.
- [ ] Load runtime readiness summary.
- [ ] Render agent view.
- [ ] Render runtime/evidence view.
- [ ] Show blocked/failure summaries.
- [ ] Add tests for observability view models.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1:
- T4.2:
- T4.3:
- T4.4:
- T4.5:
- T4.6:
- T4.7:
- T4.8:
- T4.9:

## Slice 5: Command Palette Foundation

Candidate tasks:

- [ ] Add command palette state.
- [ ] Add command registry model.
- [ ] Add text filtering.
- [ ] Register navigation commands.
- [ ] Register refresh/help commands.
- [ ] Register safe open-source-path command where supported.
- [ ] Add tests for palette filtering and command dispatch.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1:
- T5.2:
- T5.3:
- T5.4:
- T5.5:
- T5.6:
- T5.7:
- T5.8:
- T5.9:

## Slice 6: Safe Artifact Creation Actions

Candidate tasks:

- [ ] Add action confirmation model.
- [ ] Wire create ticket.
- [ ] Wire create authoring session.
- [ ] Wire create context-aware quiz.
- [ ] Wire create dump or draft where supported.
- [ ] Reload state after writes.
- [ ] Render validation errors.
- [ ] Add tests for action planning.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1:
- T6.2:
- T6.3:
- T6.4:
- T6.5:
- T6.6:
- T6.7:
- T6.8:
- T6.9:
- T6.10:

## Slice 7: Ticket Mutation Actions

Candidate tasks:

- [ ] Add ticket action confirmation model.
- [ ] Wire add ticket comment.
- [ ] Wire move ticket state.
- [ ] Wire checklist toggle where safe.
- [ ] Wire checklist creation only if supported by domain operations.
- [ ] Validate and reload ticket after mutation.
- [ ] Add tests for ticket action planning.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1:
- T7.2:
- T7.3:
- T7.4:
- T7.5:
- T7.6:
- T7.7:
- T7.8:
- T7.9:

## Slice 8: Agent Work Control

Candidate tasks:

- [ ] Add selected-ticket execution action.
- [ ] Add loop start action with explicit filters.
- [ ] Add stop/request-stop action where supported.
- [ ] Show confirmation for long-running operations.
- [ ] Show completion/failure state after operation.
- [ ] Surface evidence paths after execution.
- [ ] Add tests for action planning and result handling.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1:
- T8.2:
- T8.3:
- T8.4:
- T8.5:
- T8.6:
- T8.7:
- T8.8:
- T8.9:
