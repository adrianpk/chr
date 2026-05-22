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
| Slice 4: Agent And Runtime Observability | Delivered | `feat/slice-4-tui-agent-runtime-observability` | `feat(slice-4): tui agent runtime observability` | `.prj/reports/slices/tui-operational-console/slice-4-tui-agent-runtime-observability.md` | PR #142 merged. |
| Slice 5: Command Palette Foundation | Delivered | `feat/slice-5-tui-command-palette` | `feat(slice-5): tui command palette` | `.prj/reports/slices/tui-operational-console/slice-5-tui-command-palette.md` | PR #143 merged. |
| Slice 6: Safe Artifact Creation Actions | Delivered | `feat/slice-6-tui-artifact-creation-actions` | `feat(slice-6): tui artifact creation actions` | `.prj/reports/slices/tui-operational-console/slice-6-tui-artifact-creation-actions.md` | PR #144 merged. |
| Slice 7: Ticket Mutation Actions | Delivered | `feat/slice-7-tui-ticket-mutation-actions` | `feat(slice-7): tui ticket mutation actions` | `.prj/reports/slices/tui-operational-console/slice-7-tui-ticket-mutation-actions.md` | PR #145 merged. |
| Slice 8: Agent Work Control | Reviewing | `feat/slice-8-tui-agent-work-control` | `feat(slice-8): tui agent work control` | `.prj/reports/slices/tui-operational-console/slice-8-tui-agent-work-control.md` | PR pending review. |

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

- [x] Load daemon status where available.
- [x] Load loop and execution summaries.
- [x] Load runtime readiness summary.
- [x] Render agent view.
- [x] Render runtime/evidence view.
- [x] Show blocked/failure summaries.
- [x] Add tests for observability view models.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `af90147` loads daemon status where available.
- T4.2: `af90147` loads loop and execution summaries.
- T4.3: `af90147` loads runtime readiness summary.
- T4.4: `af90147` renders the Agent view.
- T4.5: `af90147` renders the Runtime view.
- T4.6: `af90147` shows blocked and failure summaries.
- T4.7: `af90147` adds observability view model tests.
- T4.8: `bf12912` adds the slice implementation report.
- T4.9: `make verify` passed.

## Slice 5: Command Palette Foundation

Candidate tasks:

- [x] Add command palette state.
- [x] Add command registry model.
- [x] Add text filtering.
- [x] Register navigation commands.
- [x] Register refresh/help commands.
- [x] Register safe open-source-path command where supported.
- [x] Add tests for palette filtering and command dispatch.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `6e6b228` adds command palette state.
- T5.2: `6e6b228` adds the command registry model.
- T5.3: `6e6b228` adds text filtering.
- T5.4: `6e6b228` registers navigation commands.
- T5.5: `6e6b228` registers refresh and help commands.
- T5.6: `6e6b228` registers selected source path reporting.
- T5.7: `6e6b228` adds palette filtering and dispatch tests.
- T5.8: `7e16e9d` adds the slice implementation report.
- T5.9: `make verify` passed.

## Slice 6: Safe Artifact Creation Actions

Candidate tasks:

- [x] Add action confirmation model.
- [x] Wire create ticket.
- [x] Wire create authoring session.
- [x] Wire create context-aware quiz.
- [x] Wire create dump or draft where supported.
- [x] Reload state after writes.
- [x] Render validation errors.
- [x] Add tests for action planning.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: `c985c16` adds artifact action confirmation state.
- T6.2: `c985c16` wires confirmed ticket creation through the ticket repository.
- T6.3: `c985c16` wires confirmed authoring session creation through the authoring repository.
- T6.4: `c985c16` registers context-aware quiz creation and reports provider-backed validation when unavailable from the TUI action.
- T6.5: `c985c16` registers draft creation and reports the supported promotion requirement.
- T6.6: `c985c16` reloads tickets, documents, and authoring sessions after writes.
- T6.7: `c985c16` renders action validation errors as TUI status messages.
- T6.8: `c985c16` adds action planning, cancellation, confirmation, and failure tests.
- T6.9: `d74a52d` adds the slice implementation report.
- T6.10: `make verify` passed.

## Slice 7: Ticket Mutation Actions

Candidate tasks:

- [x] Add ticket action confirmation model.
- [x] Wire add ticket comment.
- [x] Wire move ticket state.
- [x] Wire checklist toggle where safe.
- [x] Wire checklist creation only if supported by domain operations.
- [x] Validate and reload ticket after mutation.
- [x] Add tests for ticket action planning.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T7.1: `8afafbe` adds ticket action confirmation state.
- T7.2: `8afafbe` wires selected ticket comments through the ticket commenter.
- T7.3: `8afafbe` wires selected ticket status moves through the status updater.
- T7.4: `8afafbe` registers checklist toggle validation for unavailable safe writes.
- T7.5: `8afafbe` registers checklist creation validation for unavailable safe writes.
- T7.6: `8afafbe` reloads ticket state after mutation.
- T7.7: `8afafbe` adds ticket action planning, cancellation, confirmation, and failure tests.
- T7.8: `2e348c7` adds the slice implementation report.
- T7.9: `make verify` passed.

## Slice 8: Agent Work Control

Candidate tasks:

- [x] Add selected-ticket execution action.
- [x] Add loop start action with explicit filters.
- [x] Add stop/request-stop action where supported.
- [x] Show confirmation for long-running operations.
- [x] Show completion/failure state after operation.
- [x] Surface evidence paths after execution.
- [x] Add tests for action planning and result handling.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T8.1: `58c6f9d` adds selected-ticket execution action.
- T8.2: `58c6f9d` adds a ready-ticket loop action with explicit `status=ready` and `max_iterations=1`.
- T8.3: `58c6f9d` adds daemon stop request action.
- T8.4: `58c6f9d` shows confirmation for work actions.
- T8.5: `58c6f9d` renders completion or failure status after execution.
- T8.6: `58c6f9d` surfaces execution result paths in completion messages.
- T8.7: `58c6f9d` adds work action planning, cancellation, confirmation, and failure tests.
- T8.8: pending report commit.
- T8.9: `make verify` passed.
