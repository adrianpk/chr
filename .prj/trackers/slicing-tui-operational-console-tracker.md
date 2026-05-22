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
- [ ] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: TUI App Foundation | Planned | `feat/slice-1-tui-app-foundation` | `feat(slice-1): tui app foundation` | `.prj/reports/slices/tui-operational-console/slice-1-tui-app-foundation.md` | TUI package and entry behavior. |
| Slice 2: Read-Only Board And Ticket Detail | Planned | `feat/slice-2-tui-board-ticket-detail` | `feat(slice-2): tui board ticket detail` | `.prj/reports/slices/tui-operational-console/slice-2-tui-board-ticket-detail.md` | Ticket-centered read-only board. |
| Slice 3: Document And Authoring Explorer | Planned | `feat/slice-3-tui-document-authoring-explorer` | `feat(slice-3): tui document authoring explorer` | `.prj/reports/slices/tui-operational-console/slice-3-tui-document-authoring-explorer.md` | Documents and authoring artifacts. |
| Slice 4: Agent And Runtime Observability | Planned | `feat/slice-4-tui-agent-runtime-observability` | `feat(slice-4): tui agent runtime observability` | `.prj/reports/slices/tui-operational-console/slice-4-tui-agent-runtime-observability.md` | Read-only agent/runtime/evidence state. |
| Slice 5: Command Palette Foundation | Planned | `feat/slice-5-tui-command-palette` | `feat(slice-5): tui command palette` | `.prj/reports/slices/tui-operational-console/slice-5-tui-command-palette.md` | Discoverable commands. |
| Slice 6: Safe Artifact Creation Actions | Planned | `feat/slice-6-tui-artifact-creation-actions` | `feat(slice-6): tui artifact creation actions` | `.prj/reports/slices/tui-operational-console/slice-6-tui-artifact-creation-actions.md` | Low-risk creation flows. |
| Slice 7: Ticket Mutation Actions | Planned | `feat/slice-7-tui-ticket-mutation-actions` | `feat(slice-7): tui ticket mutation actions` | `.prj/reports/slices/tui-operational-console/slice-7-tui-ticket-mutation-actions.md` | Explicit ticket operations. |
| Slice 8: Agent Work Control | Planned | `feat/slice-8-tui-agent-work-control` | `feat(slice-8): tui agent work control` | `.prj/reports/slices/tui-operational-console/slice-8-tui-agent-work-control.md` | Start/stop selected work. |

## Slice 1: TUI App Foundation

Candidate tasks:

- [ ] Add Bubble Tea dependencies.
- [ ] Add `internal/chctui` package boundary.
- [ ] Add `chc tui` command.
- [ ] Dispatch `chc` without subcommands to the TUI.
- [ ] Preserve `chc --help`.
- [ ] Implement base model/update/view.
- [ ] Render project-aware navigation shell.
- [ ] Add tests for model/update behavior.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T1.1:
- T1.2:
- T1.3:
- T1.4:
- T1.5:
- T1.6:
- T1.7:
- T1.8:
- T1.9:
- T1.10:

## Slice 2: Read-Only Board And Ticket Detail

Candidate tasks:

- [ ] Load tickets through existing repositories/indexes.
- [ ] Group tickets by status.
- [ ] Render board/list rows with headers.
- [ ] Implement ticket selection.
- [ ] Render selected ticket detail.
- [ ] Show checklist and comment summaries where available.
- [ ] Add text filtering.
- [ ] Add tests for ticket view model construction.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T2.1:
- T2.2:
- T2.3:
- T2.4:
- T2.5:
- T2.6:
- T2.7:
- T2.8:
- T2.9:
- T2.10:

## Slice 3: Document And Authoring Explorer

Candidate tasks:

- [ ] Discover project-local document artifacts.
- [ ] Discover authoring sessions.
- [ ] Render document list and detail views.
- [ ] Render authoring session list and detail views.
- [ ] Show plans/slicing/tracking as artifacts.
- [ ] Preserve tickets as operational status source.
- [ ] Add tests for discovery and view models.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1:
- T3.2:
- T3.3:
- T3.4:
- T3.5:
- T3.6:
- T3.7:
- T3.8:
- T3.9:

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
