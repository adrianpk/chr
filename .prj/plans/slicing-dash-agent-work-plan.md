# Dash Agent Work Slicing Plan

Status: active

## Purpose

Add a coherent agent-work surface where users can tell ChrysoCoord to resolve
one ticket or start resolving all eligible tickets, with matching CLI and TUI
behavior and a Dash view that shows live progress.

The goal is not to create a second runner. The goal is to put a clearer product
surface over the existing ticket execution and ticket loop engine.

## Source

Primary sources:

- `.agents/playbooks/slice-delivery.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-loop.md`
- `docs/reference/cli-tui-action-matrix.md`
- Existing `chc ticket start <ticket-key>` command
- Existing `chc ticket loop` command
- Existing `service.StartSelectedTicket`
- Existing `service.StartReadyTicketLoop`
- Existing loop run, execution record, runtime artifact, and report models

Conversation decisions:

- Users need two primary agent-work intents:
  - resolve a specific ticket;
  - start resolving all eligible tickets.
- "All" initially means all tickets ready for agent development, oldest first,
  without filters.
- Dash/Work is the right home for the operational view.
- Board remains the ticket inspection and organization surface.
- Dash should observe live progress instead of owning execution.
- Filters and advanced ordering are later variants.

## Current State

Delivered on `dev`:

- CLI can start one ticket with `chc ticket start <ticket-key>`.
- CLI can run a foreground ticket loop with `chc ticket loop`.
- CLI loop supports status, priority, severity, area, tag, type, use-case, order,
  and max-iteration flags.
- TUI Work actions can start a selected ticket.
- TUI Work actions can start one ready-ticket loop iteration.
- Loop and execution records already persist state for observability.
- Reports can list and show execution and loop evidence.

Remaining gaps:

- The default agent-work intent is not named as a first-class product surface.
- The default "run all ready tickets" behavior is not presented clearly in Dash.
- TUI Work currently wraps a one-ticket ready loop, not a full "run all eligible"
  flow.
- Dash does not yet show a live queue/current/recent view for agent work.
- The CLI and TUI work terminology is not centered around the same selection
  contract.
- Future filter/order variants need a stable base contract before UI expansion.

## Target Model

One shared agent-work engine provides:

- single-ticket execution by key;
- all-ready-ticket foreground loop;
- deterministic oldest-first default order;
- shared status/filter/order model for future variants;
- persisted execution and loop records as the source of truth.

Dash/Work provides:

- ready queue preview;
- current running ticket or loop status;
- recent execution outcomes;
- actions to run a selected ticket and run all eligible tickets;
- refresh or polling while work is running.

CLI provides the same intents:

- start one ticket;
- start all eligible tickets with the default policy;
- preview selection before executing.

Future filtering and ordering should reuse the existing ticket loop query model
rather than adding a separate Dash-only selector.

## Slice Map

Slicing name: Dash Agent Work

Active slice number: Slice 1

Active slice short name: agent work selection contract

Expected branch name: `feat/slice-1-agent-work-selection-contract`

Expected PR title: `feat(slice-1): agent work selection contract`

Expected report path:
`.prj/reports/slices/dash-agent-work/slice-1-agent-work-selection-contract.md`

Prerequisite PRs or delivered work:

- `dev` includes ticket execution and ticket loop support.
- `dev` includes TUI Work actions for selected-ticket and ready-loop execution.
- `dev` includes execution, loop, and report persistence.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | agent work selection contract | `feat/slice-1-agent-work-selection-contract` | `feat(slice-1): agent work selection contract` | `.prj/reports/slices/dash-agent-work/slice-1-agent-work-selection-contract.md` |
| Slice 2 | agent work service facade | `feat/slice-2-agent-work-service-facade` | `feat(slice-2): agent work service facade` | `.prj/reports/slices/dash-agent-work/slice-2-agent-work-service-facade.md` |
| Slice 3 | cli agent work commands | `feat/slice-3-cli-agent-work-commands` | `feat(slice-3): cli agent work commands` | `.prj/reports/slices/dash-agent-work/slice-3-cli-agent-work-commands.md` |
| Slice 4 | dash agent work panel | `feat/slice-4-dash-agent-work-panel` | `feat(slice-4): dash agent work panel` | `.prj/reports/slices/dash-agent-work/slice-4-dash-agent-work-panel.md` |
| Slice 5 | dash realtime progress | `feat/slice-5-dash-realtime-progress` | `feat(slice-5): dash realtime progress` | `.prj/reports/slices/dash-agent-work/slice-5-dash-realtime-progress.md` |
| Slice 6 | agent work filters and ordering | `feat/slice-6-agent-work-filters-ordering` | `feat(slice-6): agent work filters ordering` | `.prj/reports/slices/dash-agent-work/slice-6-agent-work-filters-ordering.md` |

## Slice 1: Agent Work Selection Contract

Define the shared selection semantics for agent work.

Expected behavior:

- "Run all" selects tickets that are ready for agent development.
- The default status is `ready_for_dev`.
- Legacy `ready` tickets remain supported through existing alias behavior where
  applicable.
- The default execution order is oldest first by `created_at`, with ticket key
  tie-breakers.
- Selection is deterministic and reusable by CLI, service, and TUI.
- Dry-run or preview output can show the same selected queue without execution.

Candidate tasks:

- T1.1 Audit current ticket loop status and order defaults.
- T1.2 Add or update tests for `ready_for_dev` default selection.
- T1.3 Add or update tests for oldest-first default ordering.
- T1.4 Update ticket loop reference docs for agent-work default semantics.
- T1.5 Add slice implementation report and run quality gates.

Out of scope:

- New commands.
- New TUI panels.
- Advanced filters.

## Slice 2: Agent Work Service Facade

Create a service-level facade that names the product intent without duplicating
the runner.

Expected behavior:

- Service exposes a single-ticket agent-work operation.
- Service exposes a run-all-ready agent-work operation.
- Both operations reuse existing execution and loop internals.
- The run-all operation processes all selected tickets unless stopped by
  existing loop stop conditions.
- The facade returns enough state for CLI and Dash to render queue, current
  status, and recent outcomes.
- Existing TUI Work actions can migrate to the facade without behavior drift.

Candidate tasks:

- T2.1 Define agent-work request/result structs.
- T2.2 Implement selected-ticket facade over existing selected-ticket
      execution.
- T2.3 Implement run-all facade over existing loop execution.
- T2.4 Add tests proving the facade reuses existing execution and loop records.
- T2.5 Add slice implementation report and run quality gates.

Out of scope:

- UI layout.
- New CLI command names.
- Polling or realtime refresh.

## Slice 3: CLI Agent Work Commands

Add CLI commands that expose the product intent directly while preserving
existing ticket commands.

Expected behavior:

- CLI can run a specific ticket through the agent-work facade.
- CLI can run all ready-for-dev tickets through the agent-work facade.
- CLI can preview the run-all queue without execution.
- Existing `chc ticket start` and `chc ticket loop` continue to work.
- Command output uses the standard Summary, selection, execution, and evidence
  style already used by ticket commands.

Candidate tasks:

- T3.1 Add `chc work ticket <ticket-key>` or equivalent command.
- T3.2 Add `chc work run-all` or equivalent command.
- T3.3 Add dry-run preview for run-all.
- T3.4 Add CLI tests for selected ticket, run-all, dry-run, and no eligible
      tickets.
- T3.5 Update command/reference docs.
- T3.6 Add slice implementation report and run quality gates.

Out of scope:

- Removing or renaming existing ticket commands.
- Filter and order variants.
- TUI changes.

## Slice 4: Dash Agent Work Panel

Add the Dash/Work panel for agent-work visibility and controls.

Expected behavior:

- Dash shows a ready queue in the default run order.
- Dash shows selected-ticket and run-all controls.
- Dash shows enough safety context before running all eligible tickets.
- Dash can start a selected ticket through the service facade.
- Dash can start all eligible tickets through the service facade.
- Board remains the ticket organization surface; Dash becomes the operational
  execution surface.

Candidate tasks:

- T4.1 Identify the existing Dash/Work layout and action wiring.
- T4.2 Add queue/current/recent sections with stable dimensions.
- T4.3 Add selected-ticket and run-all controls.
- T4.4 Wire controls to the agent-work service facade.
- T4.5 Add TUI tests for rendering and action dispatch.
- T4.6 Update TUI docs and action matrix.
- T4.7 Add slice implementation report and run quality gates.

Out of scope:

- Realtime polling.
- Advanced filters.
- Daemon mode.

## Slice 5: Dash Realtime Progress

Make Dash update while agent work is running.

Expected behavior:

- Dash refreshes queue/current/recent state while a loop or execution is
  running.
- The refresh source is persisted execution and loop state, not in-memory runner
  ownership.
- Dash displays progress such as processed/eligible counts.
- Dash displays failure or blocker reasons when available.
- Dash links or points to execution and loop reports.
- The implementation remains foreground-friendly and does not require a daemon.

Candidate tasks:

- T5.1 Define the polling or refresh cadence for active work.
- T5.2 Add a service read model for agent-work observability state.
- T5.3 Wire Dash refresh to persisted execution and loop state.
- T5.4 Add tests for running, succeeded, failed, and blocked states.
- T5.5 Update docs for realtime Dash behavior.
- T5.6 Add slice implementation report and run quality gates.

Out of scope:

- Websocket/event stream infrastructure.
- Background daemon execution.
- Filtering UI.

## Slice 6: Agent Work Filters And Ordering

Expose filters and ordering after the base run-all behavior is stable.

Expected behavior:

- CLI and Dash share the same filter model.
- Filters include priority, severity, project area, type, tag, and use case.
- Ordering includes created-at, priority, severity, and priority-severity.
- Combinations are deterministic and previewable.
- Dash exposes filters without making the base run-all path harder to use.

Candidate tasks:

- T6.1 Add filter/order request model to the agent-work facade.
- T6.2 Wire CLI filters and order options to the facade.
- T6.3 Add Dash filter controls.
- T6.4 Add tests for filter combinations and order modes.
- T6.5 Update references and action matrix.
- T6.6 Add slice implementation report and run quality gates.

Out of scope:

- Persisted saved filter presets.
- Scheduling.
- Daemon automation.

## Completion Criteria

The slicing is complete when:

- CLI and Dash can both start one ticket.
- CLI and Dash can both start all eligible ready-for-dev tickets.
- Dash shows queue/current/recent agent-work state.
- Dash refreshes visible progress while work is running.
- Existing ticket start and ticket loop commands remain compatible.
- Filter and ordering variants are available through the shared facade.
- `make verify` passes after each slice and after final integration.

## Open Questions

- Final CLI command names are proposed but not locked.
- The exact Dash layout should follow the existing TUI layout constraints during
  implementation.
- Whether "stop after current" belongs in Slice 5 or a later daemon/control
  slice remains undecided.
