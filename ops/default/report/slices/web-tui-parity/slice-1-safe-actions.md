# Slice 1: Safe queue selection and canonical moves

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `fix/slice-1-tui-safe-actions`
PR: [#455](https://forge.adrianpk.com/adrianpk/chr/pulls/455)
Base: `b1052249998cf4c432adf1f9082ce211f52533b3`

## Purpose

Resolve G-026, G-034 and I-001: terminal actions must use visible targets,
canonical ticket guards and graceful shutdown while busy.

## Delivered Behavior

- Dash keeps its own selected queue key. Arrow navigation, refresh, ordering,
  removal and empty queues cannot silently select the Board's ticket. The marked
  row remains visible when navigating beyond the first queue window.
- Work commands capture the displayed ticket source revision. Native preview
  rejects a newer source image; execution retains the shared selected-record and
  snapshot guards. A pending command never rereads mutable selection.
- The palette offers `backlog`, `ready_for_dev`, `in_progress`,
  `ready_for_review`, `ready_for_merge` and `done`. Moves use the same
  revision-bound document operation as Web. Review evidence, active ownership,
  native phases and verified integration still control allowed transitions.
- Ctrl+C reaches Bubble Tea shutdown before busy modes. Pending adapter IO uses
  the managed UI context; exiting does not request cancellation of retained work.

## Implementation Notes

T1.1 reproduced three failures before correction: Dash selected the Board key
instead of its queued key in both authorities; a legacy `reviewing` move wrote
without canonical review evidence; four busy modes swallowed Ctrl+C. Reproduction
tests and their corresponding fixes share task commits so no failing test commit
is introduced.

T1.2 is `61a300444ff4814d9d2fbd61b4378577900060e4`; T1.3 is
`450d25ab66885018c476092716f9af6cc6823ed4`; T1.4 is
`3b8baf497125dd908d20011c4f0b5d66dea56d97`. The task ledger records the remaining
PR preparation and controller integration boundary.

The PTY fixture uses the production model, adapters and managed host. Only the
provider response is delayed. It waits for provider entry, sends actual Ctrl+C,
joins the child and provider, and compares retained workflow inspections. The
driver navigates to Specs before filtering its view-scoped palette and sends
separate input events after observable renders.

## Contracts Added or Changed

`TicketView.Revision`, `TicketActionRequest.Revision` and
`WorkActionRequest.TicketRevision` carry the observed source hash. Terminal moves
call `SaveTicketDocument`; native work still calls `PreviewSelectedWork` and
`StartBoundWorkExtent`. Legacy spellings remain available to existing non-TUI
interfaces. No persistence schema or domain lifecycle changes are introduced.

## Files of Interest

- `internal/tui/dash_selection.go`, `dash.go`, `navigation.go`: independent key
  selection and visible queue navigation.
- `internal/app/tui_adapters_workflow.go`, `tui_adapters_spec_plan_ticket.go`:
  source binding and shared ticket mutations.
- `internal/tui/model_update.go`, `internal/app/tui_guided_host.go`: busy quit
  routing and adapter lifetime.
- `internal/app/tui_dash_target_test.go`, `tui_safe_actions_test.go`,
  `tui_busy_terminal_linux_test.go`: owned production regression fixtures.
- `docs/how-to/user-guide/index.md`, `CHANGELOG.md`: user-visible behavior.

## Validation

Successful implementation checks:

- `go test ./internal/tui -run '^TestDash(Target|SelectionRefresh)$' -count=1`
- `go test ./internal/app -run '^TestTUIDashTarget$' -count=1 -timeout=1m`
- `go test ./internal/app -run '^Test(TUIWorkflow(Preview|Stale|Start)|TUIDashRevision|ServiceWorkflowCoordinator)$' -count=1 -timeout=3m`
- `go test ./internal/app -run '^Test(TUITicketMove|TUIBoard(Moves|Stale|Owned)|WebTicket(Editor|Revision))$' -count=1 -timeout=3m`
- `go test ./internal/tui -run '^TestBusyQuit$' -count=1`
- `go test ./internal/app -run '^TestTUIBusyTerminal$' -count=1 -timeout=1m`

The final clean PR candidate is checked after PR-reference follow-up commits.
Its exact HEAD, ordered commands and results are recorded in the controller's
validation ledger; successful final commands are listed in the canonical PR.
The required batch starts with `make verify`, followed by focused TUI and
application race regressions, then `git diff --check`. This is the development
gate and named local integration evidence, not nightly or live-forge coverage.

## Risks and Follow-ups

Controller-owned integration into dev remains pending. T1.5 and the ticket remain
open until that merge is verified. Broader queue observation and Board refresh
presentation belong to their recorded later slices.
