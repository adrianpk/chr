# Dash Agent Work Slicing Tracker

Status: completed

Plan: `.prj/plans/slicing-dash-agent-work-plan.md`

## Active Slice

Slicing name: Dash Agent Work

Active slice number: Slice 6

Active slice short name: agent work filters ordering

Expected branch name: `feat/slice-6-agent-work-filters-ordering`

Expected PR title: `feat(slice-6): agent work filters ordering`

Expected report path:
`.prj/reports/slices/dash-agent-work/slice-6-agent-work-filters-ordering.md`

Prerequisite PRs or delivered work:

- `dev` includes ticket execution and ticket loop support.
- `dev` includes TUI Work actions for selected-ticket and ready-loop execution.
- `dev` includes execution, loop, and report persistence.

## Slicing Checklist

- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [x] Complete all planned slices.
- [x] Validate completed slicing with repository quality gates.
- [x] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Agent Work Selection Contract | Completed | `feat/slice-1-agent-work-selection-contract` | `feat(slice-1): agent work selection contract` | `.prj/reports/slices/dash-agent-work/slice-1-agent-work-selection-contract.md` | Merged into `dev`. |
| Slice 2: Agent Work Service Facade | Completed | `feat/slice-2-agent-work-service-facade` | `feat(slice-2): agent work service facade` | `.prj/reports/slices/dash-agent-work/slice-2-agent-work-service-facade.md` | Merged into `dev`. |
| Slice 3: CLI Agent Work Commands | Completed | `feat/slice-3-cli-agent-work-commands` | `feat(slice-3): cli agent work commands` | `.prj/reports/slices/dash-agent-work/slice-3-cli-agent-work-commands.md` | Merged into `dev`. |
| Slice 4: Dash Agent Work Panel | Completed | `feat/slice-4-dash-agent-work-panel` | `feat(slice-4): dash agent work panel` | `.prj/reports/slices/dash-agent-work/slice-4-dash-agent-work-panel.md` | Merged into `dev`. |
| Slice 5: Dash Realtime Progress | Completed | `feat/slice-5-dash-realtime-progress` | `feat(slice-5): dash realtime progress` | `.prj/reports/slices/dash-agent-work/slice-5-dash-realtime-progress.md` | Merged into `dev`. |
| Slice 6: Agent Work Filters And Ordering | Completed | `feat/slice-6-agent-work-filters-ordering` | `feat(slice-6): agent work filters ordering` | `.prj/reports/slices/dash-agent-work/slice-6-agent-work-filters-ordering.md` | Merged into `dev`. |

## Slice 1: Agent Work Selection Contract

Candidate tasks:

- [x] T1.1 Audit current ticket loop status and order defaults.
- [x] T1.2 Add or update tests for `ready_for_dev` default selection.
- [x] T1.3 Add or update tests for oldest-first default ordering.
- [x] T1.4 Update ticket loop reference docs for agent-work default semantics.
- [x] T1.5 Add slice implementation report and run quality gates.

Task commits:

- `42c6f9f feat(slice-1): agent work selection contract`

Validation:

- `go test ./internal/chc ./internal/chcapp`
- `make verify`
- `git diff --check`

## Slice 2: Agent Work Service Facade

Candidate tasks:

- [x] T2.1 Define agent-work request/result structs.
- [x] T2.2 Implement selected-ticket facade over existing selected-ticket
      execution.
- [x] T2.3 Implement run-all facade over existing loop execution.
- [x] T2.4 Add tests proving the facade reuses existing execution and loop
      records.
- [x] T2.5 Add slice implementation report and run quality gates.

Task commits:

- `d3ed234 feat(slice-2): agent work service facade`

Validation:

- `go test ./internal/chcapp`
- `make verify`
- `git diff --check`
- `make verify`
- `git diff --check`

## Slice 3: CLI Agent Work Commands

Candidate tasks:

- [x] T3.1 Add `chc work ticket <ticket-key>` or equivalent command.
- [x] T3.2 Add `chc work run-all` or equivalent command.
- [x] T3.3 Add dry-run preview for run-all.
- [x] T3.4 Add CLI tests for selected ticket, run-all, dry-run, and no eligible
      tickets.
- [x] T3.5 Update command/reference docs.
- [x] T3.6 Add slice implementation report and run quality gates.

Task commits:

- `5ed20c3 feat(slice-3): cli agent work commands`

Validation:

- `go test ./internal/chcapp`

## Slice 4: Dash Agent Work Panel

Candidate tasks:

- [x] T4.1 Identify the existing Dash/Work layout and action wiring.
- [x] T4.2 Add queue/current/recent sections with stable dimensions.
- [x] T4.3 Add selected-ticket and run-all controls.
- [x] T4.4 Wire controls to the agent-work service facade.
- [x] T4.5 Add TUI tests for rendering and action dispatch.
- [x] T4.6 Update TUI docs and action matrix.
- [x] T4.7 Add slice implementation report and run quality gates.

Task commits:

- `cab10ff feat(slice-4): dash agent work panel`

Validation:

- `go test ./internal/chctui ./internal/chcapp`
- `make verify`
- `git diff --check`
- `make verify`
- `git diff --check`

## Slice 5: Dash Realtime Progress

Candidate tasks:

- [x] T5.1 Define the polling or refresh cadence for active work.
- [x] T5.2 Add a service read model for agent-work observability state.
- [x] T5.3 Wire Dash refresh to persisted execution and loop state.
- [x] T5.4 Add tests for running, succeeded, failed, and blocked states.
- [x] T5.5 Update docs for realtime Dash behavior.
- [x] T5.6 Add slice implementation report and run quality gates.

Task commits:

- `ea7c138 feat(slice-5): dash realtime progress`

Validation:

- `go test ./internal/chctui ./internal/chcapp`
- `make verify`
- `git diff --check`

## Slice 6: Agent Work Filters And Ordering

Candidate tasks:

- [x] T6.1 Add filter/order request model to the agent-work facade.
- [x] T6.2 Wire CLI filters and order options to the facade.
- [x] T6.3 Add Dash filter controls.
- [x] T6.4 Add tests for filter combinations and order modes.
- [x] T6.5 Update references and action matrix.
- [x] T6.6 Add slice implementation report and run quality gates.

Task commits:

- `641e6d0 feat(slice-6): agent work filters ordering`

Validation:

- `go test ./internal/chcapp ./internal/chctui`
- `make verify`
- `git diff --check`

## Completion

The slicing is complete.

Tuning needed: none recorded.
