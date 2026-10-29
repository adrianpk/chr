# Slice 1: Safe queue selection and canonical moves

Status: delivered
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `fix/slice-1-tui-safe-actions`
PR: [#455](https://forge.adrianpk.com/adrianpk/chr/pulls/455)
Base: `b1052249998cf4c432adf1f9082ce211f52533b3`
Integrated: `161622891a376efe97e3e2d0b3bbebaaf8235660` into dev through PR #455, verified 2026-10-29

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

Looper validation `YCBB7LWNNQR5OLKQDKXKLQ62HN` passed the following ordered
batch on clean candidate `161622891a376efe97e3e2d0b3bbebaaf8235660`, which is the verified
canonical integration commit:

- `make verify` — passed.
- `go test -race ./internal/tui -run '^Test(Dash(Target|SelectionRefresh)|BoardMoveCommands|BusyQuit|Model(Ownership|AsyncRefresh|PlansTicketActionCommand|WorkflowBinding|WorkflowStopAvailable))$' -count=1` — passed.
- `go test -race ./internal/app -run '^Test(TUIDash(Target|Revision)|TUITicketMove|TUIBoard(Moves|Stale|Owned)|TUIBusyTerminal|TUIWorkflow(Preview|Stale|Controls|Start)|GuidedUseCaseTUITerminal|ServiceWorkflowOwner)$' -count=1 -timeout=3m` — passed.
- `git diff --check` — passed.

The production PTY exercises actual terminal input and joins the delayed provider
on shutdown. Native and legacy adapter fixtures use owned filesystem, Git and
SQLite state. The development gate is distinct from exhaustive nightly coverage.
The initial strict-lint failure was corrected without changing acceptance checks.

## Risks and Follow-ups

None recorded.
