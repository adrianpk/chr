# Slice 10: Responsive Board and retained workflows

Status: reviewing
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-10-tui-board-workflows`
PR: pending

## Purpose

Resolve G-016, G-028 and G-031 through responsive Board presentation and retained
workflow navigation, previews and guarded controls inside Dash.

## Delivered Behavior

- Wide Board shows all six normal columns and Attention beside ticket detail.
  Narrow and maximized views remain usable. Resize, filtering and pane focus
  retain the selected ticket key and canonical column.
- Dash `w` opens an independent retained workflow list/detail browser. Empty
  lists, bounded invocation pages and revision-bound member pages remain
  read-only. Detail includes evidence, policy, current action availability and
  separate host-validated, confirmed-ticket and integrated-unit counters.
- Dash/browser `b` previews ticket, slice, set, named case or loose work. Loose
  scope accepts canonical priority/severity/area filters and order; hierarchy
  scopes retain canonical ordering. Confirmation displays actual frozen policy,
  later-ticket inclusion and snapshot, then admits at most one checkpoint.
- Browser keys and palette commands expose continue, pause, cancel, resume,
  exact-candidate approval and owner reconciliation. Unavailable actions explain
  their restriction. Allowance and decision forms collect explicit operator
  input, preview it and require separate confirmation. Stale confirmed form
  failures retain input and its original guards. Linked ticket/case navigation
  follows retained identities independently of queue selection.

## Implementation Notes

T10.1: `2e6f89388d5dd615c22765c6c1d76ca7fd61489d`.
T10.2: `8d348e4d15ab9beef1ec9581df53753b188057c8`.
T10.3: `6745653689130740a1238a5b4ba9b176f9dcb182`.
T10.4: `6f391411c3390b3392ebf3b640a53f67895c0c09`.

Responsive Board reuses canonical columns, Attention classification and existing
card/detail renderers. The retained browser keeps its own list/member offsets;
late queue observations cannot replace another selected workflow or page.

Scope preview calls the shared selection/admission owner. Filters and ranked
order are accepted only for loose work, matching its canonical request contract.
Selected ticket source fencing uses the canonical document revision independently
of the bounded display window; the complete snapshot still binds admission.

Integration forms reuse the transport-neutral input mapper and control owner.
Captured invocation, entry, budget, candidate, approval and question guards are
never replaced by a later refresh. Grant record identities rotate when input
changes. Consumption and admission limits remain retained; decision resolution
or deferral does not approve publication or record delivery.

The production PTY journey uses the root loader and managed host, widens Board,
then confirms pause/resume/cancel and checks terminal-state restrictions. Its
input driver uses the actual F6 key and waits for exact completed status messages
rather than matching words inside availability guidance. The existing 45-second
owned fixture and 40-second child limits and all observable assertions remain.
Only its owned terminal child is stopped after checks.

## Contracts Added or Changed

Terminal DTOs carry bounded workflow/member pages, application-owned action
availability and captured integration/question fences. New forms and browser
controls delegate to existing preview, status and control services; no persistence
schema, claim authority or UI-specific execution owner is added. Service files
remain transport-neutral. Terminal states remain read-only and an open PR or
worker exit cannot establish confirmed delivery.

## Files of Interest

- `internal/tui/board_responsive.go`, `shell_tickets.go`: responsive Board.
- `internal/tui/workflow_browser.go`, `workflow_controls.go`: retained navigation,
  paging, availability and linked identities.
- `internal/tui/workflow_preview_form.go`, `workflow_control_form.go`,
  `workflow_actions.go`: captured scope/control inputs and draft retention.
- `internal/app/tui_adapters_workflow.go`, `tui_adapters_workflow_view.go`,
  `tui_workflow_integration.go`: shared-owner calls and independent TUI mapping.
- `internal/app/service_workflow_action_options.go`: neutral action guidance.
- `internal/tui/*workflow*test.go`, `internal/app/tui_workflow_*test.go`:
  real retained workflows, stale controls, member windows, forms and PTY checks.
- User Guide, TUI reference, action matrix and changelog: delivered interactions.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including responsive Board,
  retained pages, captured forms, draft restoration and existing catalog frames.
- `go test ./internal/app -run '^TestTUIWorkflow(Pages|Refresh|Controls|Preview)$' -count=1 -timeout=3m` — passed for retained paging and read-only observation.
- `go test ./internal/app -run '^Test(TUIWorkScopeForm|TUIWorkAnchorWindow|TUIWorkflowPreview|TUIWorkflowStale)$' -count=1 -timeout=3m` — passed for canonical preview, cancellation and stale source/snapshot rejection.
- `go test ./internal/app -run '^TestTUIWorkflow(Grant|Decision|NativeApproval|Controls|Approval|Terminal)$' -count=1 -timeout=3m` — passed for real invocation/entry/budget/candidate/question guards, accounting, exact approval and production terminal lifecycle.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app` — passed.
- `git diff --check` — passed before task commits.

Required `make verify`, focused race checks and final whitespace evidence will be
recorded by the controller for the clean PR-reference HEAD. They are pending at
report introduction. Integration and closure remain controller-owned.

## Risks and Follow-ups

None recorded.
