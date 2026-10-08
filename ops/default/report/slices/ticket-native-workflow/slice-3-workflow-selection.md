# Slice 3: Deterministic Work Selection

Status: delivered
Delivery set: ticket-native-workflow
Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)
Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)
Branch: `feat/slice-3-workflow-selection`
PR: [#366](https://forge.adrianpk.com/adrianpk/chr/pulls/366)
Implementation candidate: `7d4a35032ad432ebf307140ec692038031665262`

## Purpose

Resolve exact ticket-native extents through one bounded application operation,
with hierarchy ordering, prerequisite proof and native loose-ticket filters.

## Delivered Behavior

- Exact ticket, slice, set and use-case identities reject missing input or
  missing records. Loose selection uses an explicit extent. Legacy records,
  unsliced use-case tickets, slice ownership and integrated work have separate
  exclusion counts. Previews preserve source files and pending journals.
- Pages return at most 128 task identities, revisions, source/contract hashes,
  membership, lifecycle and first prerequisite blocker. Snapshot hashes bind
  normalized request scope, authoritative images and referenced proof bytes.
- Use-case sets use homogeneous explicit local order or immutable creation
  time then set ID. Slice/task orders retain gaps. Interleaved ticket creation,
  classifications and labels cannot reorder hierarchy work. Changed slice
  membership fingerprints reject admission, including exact grouped tickets.
- All earlier tasks and slices are implicit prerequisites. Explicit missing,
  cyclic and forward dependencies reject. Unmet dependencies remain outside
  the requested extent and identify their next-action owner.
- Same-slice task evidence must bind successful host validation to the current
  cumulative branch, contract, commit and tree. Cross-slice or independent work
  needs a confirmed canonical delivery receipt with matching repository, tree
  and actual Git ancestry on the selected target. Labels alone do not prove it.
- Loose filters preserve native areas, types, use cases and all-tag matching,
  AND across fields and OR within lists. Independent priority/severity minima
  reject unclassified ranks. Both ranks sort before time and natural key in
  combined order. Status filters cannot re-admit integrated work.

## Implementation Notes

The central application service delegates to `WorkflowStore.Preview`; files
remain authoritative and the SQLite projection is not required. Preview shares
repository exclusion with writes but rejects a pending journal instead of
replaying it. Cancellation and a ten-second deadline bound the observation.
Record, slice, evidence-reference, filter and prerequisite limits reject
oversized inputs or expose an observation-budget blocker.

Real-files fixtures cover each selector, pagination conflicts and exclusions.
Git fixtures cover cumulative validation, receipt trees, unreachable targets,
foreign repositories and changed proof bytes. Interleaved two-use-case fixtures,
all four loose examples, missing rank/time ties and concurrent guarded writes
exercise deterministic behavior. Pending-journal fixtures prove read-only state.

## Contracts Added or Changed

- `WorkflowSelector`, `WorkflowPreviewRequest`, `WorkflowPreviewMember`,
  `WorkflowPreview` and `PreviewWorkflow` define the shared selection operation.
- `WorkflowPreviewTarget`, `WorkflowTaskValidation` and `WorkflowBlocker`
  bind prerequisite observations to immutable inputs and a selected target.
- Native rank validation accepts `unclassified`. Combined native priority and
  severity order compares both ranks before creation time and natural key.
- The shared coordinator supports a read-only preview preparation step while
  preserving ordinary domain-operation recovery.

## Files of Interest

- `internal/core/workflow_selection*.go`: selection, order, proof and filters.
- `internal/core/workflow_lock.go`: shared exclusion and pending-journal guard.
- `internal/core/ticket_loop_query.go`, `ticket_validation.go`: native ranks.
- `internal/app/service_workflow_selection.go`: adapter-neutral operation.
- [Workflow selection reference](../../../../../docs/reference/workflow-selection.md).

## Validation

Successful implementation checks:

- `make verify`
- `go test ./internal/core ./internal/app -run '^(TestWorkflowSelection|TestServiceWorkflowPreview|TestTicketLoopQuery|TestTicketParser)' -count=1`
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowSelection|TestServiceWorkflowPreview|TestTicketLoopQuery)' -count=1`
- `git diff --check`

The controller passed `make verify`,
`go test ./internal/core ./internal/app -count=1` and `git diff --check` in that
order on clean head `b39a0051818b1fa9c08b88cf2e79658b173cf16f`. Its immutable-head
JSON and log retain the evidence. Canonical PR #366 merged that head into `dev`
on 2026-10-08 at 16:42:08 UTC; all four task tickets are solved. These are
Slice 3 checks; the composed final gate remains in the approved plan.

Repository checks use the retained owned `fixtures/slice-2/dump-curation`
executable through PATH, as documented in the [Slice 2 report](slice-2-workflow-materialization.md).
It preserves the exact existing CLI dump fixture text and delegates every other
request to the installed executable. It has no persistent service to stop.

Acceptance: A04 hierarchy order and unsliced exclusion; A05 loose filters and
ties; A07 scoped prerequisites; A24 bounded query inputs and pages. Later
continuation and integration slices compose these selection contracts.

## Risks and Follow-ups

Preview requires explicitly versioned/adopted records. Missing Git objects,
changed proof images, pending recovery or exhausted observation budgets block
selection. Receipt creation and provider reconciliation belong to delivery;
claims, invocation retention and transport controls remain in their recorded
implementation slices. The CLI curation fixture is a validation dependency.
